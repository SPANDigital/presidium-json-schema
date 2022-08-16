package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SPANDigital/presidium-json-schema/templates"
	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type SchemaConverter interface {
	Convert(path string) error
}

type Converter struct {
	config    Config
	compiler  *jsonschema.Compiler
	template  *template.Template
	converted map[string]bool
	patterns  map[string]string
	order     map[string]*orderedmap.OrderedMap
}

type middlewareFunc func(prop interface{}) interface{}

func NewConverter(config Config) *Converter {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	return &Converter{
		config:    config,
		compiler:  compiler,
		converted: map[string]bool{},
		patterns:  map[string]string{},
		order:     map[string]*orderedmap.OrderedMap{},
	}
}

func (c Converter) Clean() error {
	if !PathExist(c.config.Destination) {
		return nil
	}
	if err := AppFS.RemoveAll(c.config.Destination); err != nil {
		return err
	}

	return nil
}

func (c *Converter) Convert(path string) error {
	c.Clean()
	if err := c.parseTemplates(); err != nil {
		return err
	}

	paths, err := FindFiles(path, c.config.Extension, c.config.Recursive)
	if err != nil {
		return err
	}

	for _, path := range paths {
		log.Infof("loading schema: %s", path)
		if err := c.loadSchema(path); err != nil {
			return err
		}
	}

	schemas, err := c.compileSchemas(paths)
	if err != nil {
		return err
	}

	for _, schema := range schemas {
		c.converted[schema.Location] = true
	}

	for _, schema := range schemas {
		definitions := schema.Definitions()
		if err := c.convertToMarkdown("_index", schema); err != nil {
			return err
		}

		for _, def := range definitions {
			if _, ok := c.converted[def.Location]; ok {
				continue
			}

			name := FileName(def.Title, def.Location)
			if err := c.convertToMarkdown(name, def); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseTemplates parses all gohtml templates from the embedded fs
func (c *Converter) parseTemplates() (err error) {
	c.template = template.New("").Funcs(FuncMap(c.config.ReferenceUrl(), c.patterns, c.order))
	c.template, err = c.template.ParseFS(templates.Files, "*.gohtml")
	if err != nil {
		return errors.Wrap(err, "failed to parse templates")
	}
	return nil
}

// loadSchema loads the schema as raw json to apply the middleware
func (c *Converter) loadSchema(path string) error {
	schemaFile, err := AppFS.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to load schema: %s", path)
	}
	defer schemaFile.Close()

	b, err := ioutil.ReadAll(schemaFile)
	if err != nil {
		return err
	}

	var schema RawSchema
	if err = json.Unmarshal(b, &schema); err != nil {
		return errors.Wrapf(err, "failed to decode schema: %s", path)
	}

	if c.config.Ordered {
		c.order[path] = orderedmap.New()
		if err = json.Unmarshal(b, c.order[path]); err != nil {
			return errors.Wrapf(err, "failed to decode schema: %s", path)
		}
	}

	c.applyMiddleware(schema)
	b, err = json.Marshal(schema)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal schema: %s", path)
	}

	url := FirstNonEmpty(schema.Id(), path)
	return c.compiler.AddResource(url, bytes.NewReader(b))
}

// applyMiddleware recursively walks through the json schema and applies the middleware
func (c *Converter) applyMiddleware(m map[string]interface{}) {
	for key, fn := range c.middleware() {
		if prop, ok := m[key]; ok {
			m[key] = fn(prop)
		}
	}

	for _, prop := range m {
		if _, ok := prop.(map[string]interface{}); ok {
			c.applyMiddleware(prop.(map[string]interface{}))
		}
	}
}

// compileSchemas compiles each schema from their Path
func (c *Converter) compileSchemas(paths []string) ([]*Schema, error) {
	var schemas []*Schema
	for _, path := range paths {
		log.Debugf("compiling schema: %s", path)
		schema, err := c.compiler.Compile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to compile schema: %schema", path)
		}
		schemas = append(schemas, ToSchema(schema, path))
	}
	return schemas, nil
}

// convertToMarkdown executes the template to convert the schema to md
func (c *Converter) convertToMarkdown(filename string, schema *Schema) error {
	path := filepath.Join(c.config.Destination, FilePath(schema.Location))
	if err := c.createIndex(path); err != nil {
		return err
	}

	log.Debugf("converting schema to md: %s", path)
	filename = fmt.Sprintf("%s.md", filename)
	path = filepath.Join(path, filename)
	mdFile, err := AppFS.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create md file: %s", path)
	}

	defer mdFile.Close()

	c.converted[schema.Location] = true
	return c.template.ExecuteTemplate(mdFile, "base.gohtml", schema)
}

// createIndex creates a _index.md file for each directory in the Path
func (c *Converter) createIndex(path string) error {
	log.Debugf("creating index: %s", path)
	path = filepath.Clean(path)
	if c.config.Destination == path {
		return nil
	}

	indexPath := filepath.Join(path, "_index.md")
	if _, err := AppFS.Stat(indexPath); !os.IsNotExist(err) {
		return nil
	}

	if err := AppFS.MkdirAll(path, fs.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to created index directory: %s", path)
	}

	dir := filepath.Base(path)
	fm := fmt.Sprintf("---\ntitle: %s\n---", dir)
	if err := afero.WriteFile(AppFS, indexPath, []byte(fm), os.ModePerm); err != nil {
		return err
	}

	return c.createIndex(filepath.Dir(path))
}

// Regex lookahead/behind is not supported in Go and the schema will not compile if the regex is invalid.
// As a workaround, replace the regex with a hash and use the hash to load it from the template
// See https://github.com/santhosh-tekuri/jsonschema/issues/31
func (c *Converter) middleware() map[string]middlewareFunc {
	return map[string]middlewareFunc{
		"patternProperties": func(prop interface{}) interface{} {
			props := prop.(map[string]interface{})
			patterns := map[string]interface{}{}
			for k, v := range props {
				h := Hash(k)
				c.patterns[h] = k
				patterns[h] = v
			}
			return patterns
		},
		"pattern": func(prop interface{}) interface{} {
			pattern := prop.(string)
			h := Hash(pattern)
			c.patterns[h] = pattern
			return h
		},
	}
}
