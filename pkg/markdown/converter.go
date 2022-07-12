package markdown

import (
	"encoding/json"
	"fmt"
	"github.com/SPANDigital/presidium-json-schema/templates"
	"github.com/fsnotify/fsnotify"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

type Converter struct {
	config      Config
	compiler    *jsonschema.Compiler
	definitions map[string]*jsonschema.Schema
	template    *template.Template
	done        chan struct{}
}

func NewConverter(config Config) *Converter {
	return &Converter{
		config:      config,
		compiler:    jsonschema.NewCompiler(),
		definitions: map[string]*jsonschema.Schema{},
	}
}

func (c *Converter) Convert(path string) error {
	c.compiler.ExtractAnnotations = true
	schemas, err := FindFiles(path, c.config.Extension, c.config.Recursive)
	if err != nil {
		return err
	}

	for _, path := range schemas {
		if err := c.loadSchema(path); err != nil {
			return err
		}
	}

	return c.watch(schemas, c.convertSchemas)
}

// For development only, will be removed
func (c Converter) watch(schemas []string, delegate func(s []string) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := delegate(schemas); err != nil {
						log.Print(err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("./templates")
	if err != nil {
		return err
	}

	<-c.done
	return nil
}

func (c Converter) loadSchema(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var header SchemaHeader
	dec := json.NewDecoder(f)
	if err = dec.Decode(&header); err != nil {
		return err
	}

	var url = header.Id
	if len(url) == 0 {
		url = path
	}

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	return c.compiler.AddResource(url, f)
}

func (c Converter) convertSchemas(schemas []string) (err error) {
	c.template = template.New("").Funcs(FuncMap())

	c.template, err = c.template.ParseFS(templates.Files, "*.gohtml")
	if err != nil {
		return err
	}

	for _, path := range schemas {
		s, err := c.compiler.Compile(path)
		if err != nil {
			return err
		}

		if err = c.convertSchema(s); err != nil {
			return err
		}
	}

	return nil
}

func (c Converter) convertSchema(s *jsonschema.Schema) error {
	path := filepath.Join(c.config.Destination, SchemaFileName(s.Title, s.Location))
	if err := os.MkdirAll(path, fs.ModePerm); err != nil {
		return err
	}

	if err := c.renderSchema(path, "_index", s); err != nil {
		return err
	}

	c.definitions = map[string]*jsonschema.Schema{}
	c.findDefinitions(s, TrimAnchor(s.Location), map[string]bool{})
	for _, schema := range c.definitions {
		fp := filepath.Join(path, SchemaFilePath(schema.Location))
		name := SchemaFileName(schema.Title, schema.Location)
		if err := c.renderSchema(fp, name, schema); err != nil {
			return err
		}
	}

	return nil
}

func (c Converter) renderSchema(path, filename string, schema *jsonschema.Schema) error {
	if err := c.createIndex(path); err != nil {
		return err
	}

	path = filepath.Join(path, filename)
	file, err := os.Create(fmt.Sprintf("%s.md", path))
	if err != nil {
		return err
	}

	defer file.Close()

	return c.template.ExecuteTemplate(file, "base.gohtml", schema)
}

func (c Converter) createIndex(path string) error {
	if c.config.Destination == path {
		return nil
	}

	idx := filepath.Join(path, "_index.md")
	if _, err := os.Stat(idx); !os.IsNotExist(err) {
		return nil
	}

	if err := os.MkdirAll(path, fs.ModePerm); err != nil {
		return err
	}

	dir := filepath.Base(path)
	fm := fmt.Sprintf("---\ntitle: %s\n---", dir)
	file, err := os.Create(idx)
	if err != nil {
		return err
	}

	if _, err := file.Write([]byte(fm)); err != nil {
		return err
	}

	defer file.Close()

	return c.createIndex(filepath.Dir(path))
}

func (c *Converter) findDefinitions(s *jsonschema.Schema, location string, visited map[string]bool) {
	if s == nil || visited[s.Location] {
		return
	}

	visited[s.Location] = true
	if s.Ref != nil {
		if IsInternalRef(location, s.Ref.Location) || IsRemoteRef(s.Ref.Location) {
			c.definitions[s.Ref.Location] = s.Ref
		}
	}

	traverse := func(schema *jsonschema.Schema) {
		c.findDefinitions(schema, location, visited)
	}

	traverseEach := func(schemas ...*jsonschema.Schema) {
		for _, schema := range schemas {
			traverse(schema)
		}
	}

	traverseEach(s.AnyOf...)
	traverseEach(s.AllOf...)
	traverseEach(s.OneOf...)
	traverseEach(s.PrefixItems...)
	traverseEach(
		s.Not, s.Else, s.Then, s.Contains,
		s.DynamicRef, s.Ref, s.RecursiveRef,
		s.PropertyNames, s.UnevaluatedItems,
		s.UnevaluatedProperties, s.Items2020,
	)

	for _, schema := range s.Properties {
		traverse(schema)
	}

	for _, schema := range s.DependentSchemas {
		traverse(schema)
	}

	for _, schema := range s.PatternProperties {
		traverse(schema)
	}

	switch s.Items.(type) {
	case *jsonschema.Schema:
		traverse(s.Items.(*jsonschema.Schema))
	case []*jsonschema.Schema:
		traverseEach(s.Items.([]*jsonschema.Schema)...)
	}

	return
}
