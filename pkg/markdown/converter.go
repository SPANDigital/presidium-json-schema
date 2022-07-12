package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SPANDigital/presidium-json-schema/templates"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

type Converter struct {
	config   Config
	compiler *jsonschema.Compiler
	template *template.Template
	patterns map[string]string
}

type middlewareFunc func(prop interface{}) interface{}

func NewConverter(config Config) *Converter {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	return &Converter{
		config:   config,
		compiler: compiler,
		patterns: map[string]string{},
	}
}

func (c *Converter) Convert(path string) error {
	schemas, err := FindFiles(path, c.config.Extension, c.config.Recursive)
	if err != nil {
		return err
	}

	for _, path := range schemas {
		if err := c.loadSchema(path); err != nil {
			return err
		}
	}

	return c.compileSchemas(schemas)
}

func (c Converter) loadSchema(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var schema RawSchema
	dec := json.NewDecoder(file)
	if err = dec.Decode(&schema); err != nil {
		return err
	}

	c.applyMiddleware(schema)
	b, err := json.Marshal(schema)
	if err != nil {
		return nil
	}

	url := FirstNonEmpty(schema.Id(), path)
	return c.compiler.AddResource(url, bytes.NewReader(b))
}

func (c Converter) applyMiddleware(m map[string]interface{}) {
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

func (c Converter) compileSchemas(schemas []string) (err error) {
	c.template = template.New("").Funcs(FuncMap(c.config.ReferenceUrl, c.patterns))
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
	var definitions []*jsonschema.Schema
	WalkSchema(s, true, func(s *jsonschema.Schema) error {
		if s.Ref != nil {
			definitions = append(definitions, s.Ref)
		}
		return nil
	})

	if err := c.convertToMarkdown("_index", s); err != nil {
		return err
	}

	for _, def := range definitions {
		fileName := FileName(def.Title, def.Location)
		if err := c.convertToMarkdown(fileName, def); err != nil {
			return err
		}
	}
	return nil
}

func (c Converter) convertToMarkdown(filename string, schema *jsonschema.Schema) error {
	path := filepath.Join(c.config.Destination, FilePath(schema.Location))
	if err := c.createIndex(path); err != nil {
		return err
	}

	filename = fmt.Sprintf("%s.md", filename)
	path = filepath.Join(path, filename)

	file, err := os.Create(path)
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

	_index := filepath.Join(path, "_index.md")
	if _, err := os.Stat(_index); !os.IsNotExist(err) {
		return nil
	}

	if err := os.MkdirAll(path, fs.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(_index)
	if err != nil {
		return err
	}

	defer file.Close()

	dir := filepath.Base(path)
	fm := fmt.Sprintf("---\ntitle: %s\n---", dir)
	if _, err := file.Write([]byte(fm)); err != nil {
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
			hash := Hash(pattern)
			c.patterns[hash] = pattern
			return hash
		},
	}
}
