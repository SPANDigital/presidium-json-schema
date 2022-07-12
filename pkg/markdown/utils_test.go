package markdown

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestSlugify(t *testing.T) {
	testCases := map[string]string{
		"/properties/dimensions": "properties-dimensions",
		"test/ref.schema.json#":  "test-ref-schema-json",
	}

	for val, expected := range testCases {
		actual := Slugify(val)
		assert.Equal(t, actual, expected)
	}
}

func TestDict(t *testing.T) {
	actual, err := Dict("Name", "Joe", "Age", 1)
	assert.Nil(t, err)
	assert.Equal(t, actual, map[string]interface{}{
		"Name": "Joe",
		"Age":  1,
	})

	_, err = Dict("Name", "Joe", "Age")
	assert.NotNil(t, err)

	_, err = Dict("Name", "Joe", 1, "Age")
	assert.NotNil(t, err)
}

func TestJoin(t *testing.T) {
	testCases := map[string][]string{
		"a,b,c": {"a", "b", "c"},
		"a":     {"a"},
	}

	for expected, values := range testCases {
		actual := Join(values, ",")
		assert.Equal(t, actual, expected)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	testCases := map[string][]string{
		"c": {"", "", "c"},
		"b": {"", "b"},
		"":  {"", ""},
	}

	for expected, values := range testCases {
		actual := FirstNonEmpty(values...)
		assert.Equal(t, actual, expected)
	}
}

func TestIsSlice(t *testing.T) {
	actual := IsSlice([]string{""})
	assert.Equal(t, actual, true)

	actual = IsSlice("")
	assert.Equal(t, actual, false)

	actual = IsSlice([]*jsonschema.Schema{})
	assert.Equal(t, actual, true)

	actual = IsSlice([]jsonschema.Schema{})
	assert.Equal(t, actual, true)
}

func TestIsSchema(t *testing.T) {
	actual := IsSchema([]string{""})
	assert.Equal(t, actual, false)

	actual = IsSchema("")
	assert.Equal(t, actual, false)

	actual = IsSchema(&jsonschema.Schema{})
	assert.Equal(t, actual, true)

	actual = IsSchema([]*jsonschema.Schema{})
	assert.Equal(t, actual, false)

	actual = IsSchema(jsonschema.Schema{})
	assert.Equal(t, actual, true)
}

func TestSlice(t *testing.T) {
	values := []string{"a", "b", "c"}
	actual := Slice(values...)
	assert.Equal(t, actual, values)
}

func TestAppend(t *testing.T) {
	values := []string{"a", "b", "c"}
	actual := Append([]string{}, values...)
	assert.Equal(t, actual, values)
}

func TestEscapeRegex(t *testing.T) {
	testCases := map[*regexp.Regexp]string{
		regexp.MustCompile("^(abc|b|c|d)$"): "`^(abc\\|b\\|c\\|d)$`",
		regexp.MustCompile("^\\W$"):         "`^\\W$`",
		regexp.MustCompile("^"):             "`^`",
	}

	for val, expected := range testCases {
		actual := EscapeRegex(val)
		assert.Equal(t, actual, expected)
	}
}

func TestTitle(t *testing.T) {
	testCases := map[string]string{
		"hello World": "Hello World",
		"helloWorld":  "Helloworld",
		"Hello":       "Hello",
		"wORLD":       "World",
	}

	for val, expected := range testCases {
		actual := Title(val)
		assert.Equal(t, actual, expected)
	}
}

func TestIsRemoteRef(t *testing.T) {
	testCases := map[string]bool{
		"https://spandigital.com":                 true,
		"http://ref.schema.json#":                 true,
		"http://json-schema.org/draft-06/schema#": true,
		"file://json-schema.org/draft-06/schema#": false,
		"http:/json-schema.org/draft-06/schema#":  false,
		"https//json-schema.org/draft-06/schema#": false,
	}

	for val, expected := range testCases {
		actual := IsRemoteRef(val)
		assert.Equal(t, actual, expected)
	}
}

func TestIsInternalRef(t *testing.T) {
	location := "http://ref.schema.json"
	testCases := map[string]bool{
		"https://spandigital.com":                       false,
		"http://ref.schema.json#":                       true,
		"http://ref.schema.json#/properties/attributes": true,
		"http://ref.schema.json":                        true,
	}

	for val, expected := range testCases {
		actual := IsInternalRef(location, val)
		assert.Equal(t, actual, expected)
	}
}

func TestAnchor(t *testing.T) {
	testCases := map[string]string{
		"presidium-json-schema/test/sample.schema.json#/properties/dimensions": "/properties/dimensions",
		"#/properties/dimensions": "/properties/dimensions",
		"test/ref.schema.json#":   "ref.schema",
	}

	for val, expected := range testCases {
		actual := Anchor(val)
		assert.Equal(t, actual, expected)
	}
}

func TestAnchorize(t *testing.T) {
	testCases := map[string]string{
		"presidium-json-schema/test/sample.schema.json#/properties/dimensions": "properties-dimensions",
		"test/ref.schema.json#": "ref-schema",
	}

	for val, expected := range testCases {
		actual := Anchorize(val)
		assert.Equal(t, actual, expected)
	}
}

func TestSchemaFilePath(t *testing.T) {
	testCases := map[string]string{
		"/test/sample.schema.json#/properties/dimensions": "/properties/",
		"ref.schema.json#/dimensions/attributes/":         "/dimensions/attributes/",
		"#/dimensions/attributes/ref.json":                "/dimensions/attributes/",
		"test/ref.schema.json#":                           "",
		"ref.schema.json#":                                "",
		"#":                                               "",
	}

	for val, expected := range testCases {
		actual := SchemaFilePath(val)
		assert.Equal(t, actual, expected)
	}
}

func TestSchemaFileName(t *testing.T) {
	actual := SchemaFileName("", "test/ref.schema.json#")
	assert.Equal(t, actual, "ref-schema")

	actual = SchemaFileName("Ref", "test/ref.schema.json#")
	assert.Equal(t, actual, "ref")

	actual = SchemaFileName("#$%#", "test/ref.schema.json#")
	assert.Equal(t, actual, "ref-schema")

	actual = SchemaFileName("#$%#", "test/")
	assert.Equal(t, actual, "test")
}

func TestFilenameWithoutExt(t *testing.T) {
	testCases := map[string]string{
		"test/ref.schema.json#": "ref.schema",
		"/sample/file.ext":      "file",
	}

	for val, expected := range testCases {
		actual := FilenameWithoutExt(val)
		assert.Equal(t, actual, expected)
	}
}

func TestHumanize(t *testing.T) {
	testCases := map[string]string{
		"test/ref.schema.json#":                   "ref.schema",
		"http://json-schema.org/draft-04/schema#": "schema",
	}

	for val, expected := range testCases {
		actual := Humanize(val)
		assert.Equal(t, actual, expected)
	}
}
