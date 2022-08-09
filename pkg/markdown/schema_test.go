package markdown

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ref = &jsonschema.Schema{
	Location: "b",
}

var schemas = []*jsonschema.Schema{
	{Location: "c"},
	{Location: "d", Ref: ref},
	{Location: "e", Ref: ref},
}

var root = &jsonschema.Schema{
	Location: "a",
	Ref:      ref,
	AllOf:    schemas,
	AnyOf:    schemas,
	OneOf:    schemas,
}

func TestRawSchema_Id(t *testing.T) {
	var s = RawSchema{"$id": "test"}
	assert.Equal(t, s.Id(), "test")
}

func TestWalkSchema(t *testing.T) {
	var expected = append(schemas, ref, root)
	var definitions []*jsonschema.Schema
	ToSchema(root, "").WalkSchema(true, func(s *Schema) error {
		definitions = append(definitions, s.Schema)
		return nil
	})

	assert.ElementsMatch(t, expected, definitions)
}

func TestWalkNoRefSchema(t *testing.T) {
	var expected = append(schemas, root)
	var definitions []*jsonschema.Schema
	ToSchema(root, "").WalkSchema(false, func(s *Schema) error {
		definitions = append(definitions, s.Schema)
		return nil
	})

	assert.ElementsMatch(t, expected, definitions)
}

func TestWalkRecursiveSchema(t *testing.T) {
	a := &jsonschema.Schema{Location: "a"}
	b := &jsonschema.Schema{Location: "b", Ref: a}
	a.Ref = b

	var expected = []*jsonschema.Schema{a, b}
	var definitions []*jsonschema.Schema
	ToSchema(a, "").WalkSchema(true, func(s *Schema) error {
		definitions = append(definitions, s.Schema)
		return nil
	})

	assert.ElementsMatch(t, expected, definitions)
}
