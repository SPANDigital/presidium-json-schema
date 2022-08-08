package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _ = runtime.Caller(0)
	rootPath   = filepath.Join(filepath.Dir(b), "../..")
)

func init() {
	AppFS = afero.NewMemMapFs()
}

var config = Config{
	Destination: "/a",
	Local:       false,
	Extension:   "*.schema.json",
	Recursive:   false,
}

func TestNewConverter(t *testing.T) {
	c := NewConverter(config)
	assert.Equal(t, config, c.config)
	assert.NotNil(t, c.patterns)
	assert.NotNil(t, c.compiler)
}

func TestConverter_Clean(t *testing.T) {
	c := NewConverter(config)
	test_output := filepath.Join(rootpath, "test_output")
	test_file, err := AppFS.Create(test_output)
	assert.Nil(t, err)
	defer test_file.Close()

	c.config.Destination = test_output
	exists, err := afero.Exists(AppFS, test_output)
	assert.Nil(t, err)
	assert.True(t, exists)

	err = c.Clean()
	assert.Nil(t, err)

	exists, err = afero.Exists(AppFS, test_output)
	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestConverter_Convert(t *testing.T) {
	c := NewConverter(config)
	err := c.Convert(filepath.Join(rootPath, "test"))
	assert.Nil(t, err)
}

func TestConverter_parseTemplates(t *testing.T) {
	c := NewConverter(config)
	err := c.parseTemplates()
	assert.Nil(t, err)
	assert.NotNil(t, c.template)

	templates := []string{
		"any.gohtml", "base.gohtml",
		"number.gohtml", "property.gohtml",
		"string.gohtml", "typeof.gohtml", "array.gohtml",
		"object.gohtml", "schema.gohtml",
	}

	for _, template := range templates {
		assert.NotNil(t, c.template.Lookup(template), fmt.Sprintf("expected template %s to be defined", template))
	}
}

func TestConverter_compileSchemas(t *testing.T) {
	c := NewConverter(config)
	paths := []string{
		"https://json-schema.org/draft/2020-12/schema",
		"https://json-schema.org/draft/2019-09/schema",
		"https://json-schema.org/draft-07/schema",
		"https://json-schema.org/draft-06/schema",
		"https://json-schema.org/draft-04/schema",
	}

	schemas, err := c.compileSchemas(paths)
	assert.Nil(t, err)
	assert.Equal(t, len(paths), len(schemas))
}

func TestConverter_convertToMarkdown(t *testing.T) {
	c := NewConverter(config)
	err := c.parseTemplates()
	assert.Nil(t, err)

	err = c.convertToMarkdown("test", &Schema{
		"",
		&jsonschema.Schema{
			Title: "hello",
		},
	})
	assert.Nil(t, err)

	path := filepath.Join(c.config.Destination, "test.md")
	exist, err := afero.Exists(AppFS, path)
	assert.True(t, exist)
	contains, err := afero.FileContainsBytes(AppFS, path, []byte("hello"))
	assert.True(t, contains)
}

func TestConverter_createIndex(t *testing.T) {
	c := NewConverter(config)
	path := filepath.Join(config.Destination, "b/c/d")
	err := c.createIndex(path)
	assert.Nil(t, err)
	validatePath(t, c.config.Destination, path)
}

func TestConverter_applyMiddleware(t *testing.T) {
	c := NewConverter(config)
	h := Hash("a")
	rawSchema := map[string]interface{}{
		"pattern": "a",
		"patternProperties": map[string]interface{}{
			"a": 1,
		},
	}

	actual := map[string]interface{}{
		"pattern": h,
		"patternProperties": map[string]interface{}{
			h: 1,
		},
	}
	c.applyMiddleware(rawSchema)
	assert.Equal(t, rawSchema, actual)
}

func TestConverter_middleware(t *testing.T) {
	c := NewConverter(config)
	m := c.middleware()
	assert.NotNil(t, m["patternProperties"])
	assert.NotNil(t, m["pattern"])
}

func TestConverter_patternPropertyMiddleware(t *testing.T) {
	c := NewConverter(config)
	m := c.middleware()
	h := Hash("a")
	res := m["patternProperties"](map[string]interface{}{
		"a": 1,
	})
	if val, ok := res.(map[string]interface{}); ok {
		assert.NotNil(t, val[h])
		return
	}
	t.Fail()
}

func TestConverter_patternMiddleware(t *testing.T) {
	c := NewConverter(config)
	m := c.middleware()
	h := Hash("a")
	res := m["pattern"]("a")
	if val, ok := res.(string); ok {
		assert.Equal(t, h, val)
		return
	}
	t.Fail()
}

func validatePath(t *testing.T, root, path string) {
	if path == root {
		return
	}

	dst := filepath.Join(path, "_index.md")
	exist, err := afero.Exists(AppFS, dst)
	assert.Nil(t, err)
	assert.True(t, exist)
	validatePath(t, root, filepath.Dir(path))
}

func writeSchema(t *testing.T, path string, content string) {
	err := afero.WriteFile(AppFS, path, []byte(content), os.ModePerm)
	assert.Nil(t, err)
}
