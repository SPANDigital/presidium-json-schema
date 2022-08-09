package markdown

import (
	"github.com/iancoleman/orderedmap"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
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
	testCases := map[string]string{
		"^(abc|b|c|d)$": "`^(abc\\|b\\|c\\|d)$`",
		"^\\W$":         "`^\\W$`",
		"^":             "`^`",
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

func TestSchemaFilePath(t *testing.T) {
	testCases := map[string]string{
		"ref.schema.json#/definitions/pagination/properties/first": "ref-schema/definitions",
		"test/ref.schema.json#/attributes":                         "ref-schema",
		"test/ref.schema.json#/definitions/attributes/dimensions":  "ref-schema/definitions",
		"test/ref.schema.json#/attributes/dimensions":              "ref-schema",
		"ref.schema.json#": "ref-schema",
	}

	for val, expected := range testCases {
		actual := FilePath(val)
		assert.Equal(t, expected, actual)
	}
}

func TestSchemaFileName(t *testing.T) {
	actual := FileName("", "test/ref.schema.json#")
	assert.Equal(t, actual, "ref-schema")

	actual = FileName("", "/test/sample.schema.json#/properties/dimensions")
	assert.Equal(t, actual, "dimensions")

	actual = FileName("Ref", "test/ref.schema.json#")
	assert.Equal(t, actual, "ref")

	actual = FileName("#$%#", "test/ref.schema.json#")
	assert.Equal(t, actual, "ref-schema")

	actual = FileName("#$%#", "test/")
	assert.Equal(t, actual, "test")
}

func TestGetUrl(t *testing.T) {
	link := GetPermalink("reference")
	actual := link(&jsonschema.Schema{
		Location: "/test/sample.schema.json#/properties/dimensions",
	})
	assert.Equal(t, "[dimensions]({{%baseurl%}}/reference/sample-schema/#dimensions)", actual)

	actual = link(&jsonschema.Schema{
		Title:    "sample",
		Location: "/test/sample.schema.json#/properties/dimensions",
	})
	assert.Equal(t, "[sample]({{%baseurl%}}/reference/sample-schema/#sample)", actual)

	actual = link(&jsonschema.Schema{
		Location: "sample.schema.json#/dimensions",
	})
	assert.Equal(t, "[dimensions]({{%baseurl%}}/reference/sample-schema/#dimensions)", actual)
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

func TestGetAnchorPath(t *testing.T) {
	testCases := map[string]string{
		"ref.schema.json#/definitions/pagination/properties/first": "/definitions/pagination/properties/first",
		"test/ref.schema.json#/attributes":                         "/attributes",
		"test/ref.schema.json#/definitions/attributes/dimensions":  "/definitions/attributes/dimensions",
		"test/ref.schema.json#/attributes/dimensions":              "/attributes/dimensions",
		"ref.schema.json#": "",
	}

	for val, expected := range testCases {
		actual := AnchorPath(val)
		assert.Equal(t, expected, actual)
	}
}

func TestTrimAnchor(t *testing.T) {
	testCases := map[string]string{
		"ref.schema.json#/definitions/pagination/properties/first": "ref.schema.json",
		"test/ref.schema.json#/attributes":                         "test/ref.schema.json",
		"test/ref.schema.json#/definitions/attributes/dimensions":  "test/ref.schema.json",
		"ref.schema.json#": "ref.schema.json",
	}

	for val, expected := range testCases {
		actual := TrimAnchorPath(val)
		assert.Equal(t, expected, actual)
	}
}

func TestGetWeight(t *testing.T) {
	a := orderedmap.New()
	b := orderedmap.New()
	c := orderedmap.New()

	c.Set("a", "")
	c.Set("b", "")
	c.Set("c", "")

	b.Set("a", "")
	b.Set("b", c)

	a.Set("a", *b)

	weightFn := GetWeight(map[string]*orderedmap.OrderedMap{"test.json": a})

	testCases := map[string]int{
		"#/a/b/c": 3,
		"#/a/b/z": -1,
		"#/a/b":   2,
		"#/a":     1,
		"#":       -1,
	}

	for loc, expected := range testCases {
		actual := weightFn("test.json", loc)
		assert.Equal(t, expected, actual)
	}
}
