package markdown

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// WalkSchema walks the schema tree, calling fn for each schema in the tree, including root.
func WalkSchema(s *jsonschema.Schema, followRef bool, fn func(s *jsonschema.Schema) error) {
	walkSchema(s, followRef, map[string]bool{}, fn)
}

func walkSchema(s *jsonschema.Schema, followRef bool, visited map[string]bool, fn func(s *jsonschema.Schema) error) {
	if s == nil || visited[s.Location] {
		return
	}

	if err := fn(s); err != nil {
		return
	}

	visited[s.Location] = true
	walk := func(schema *jsonschema.Schema) {
		walkSchema(schema, followRef, visited, fn)
	}

	walkEach := func(schemas ...*jsonschema.Schema) {
		for _, schema := range schemas {
			walk(schema)
		}
	}

	if followRef {
		walkEach(s.DynamicRef, s.Ref, s.RecursiveRef)
	}

	walkEach(s.AnyOf...)
	walkEach(s.AllOf...)
	walkEach(s.OneOf...)
	walkEach(s.PrefixItems...)
	walkEach(
		s.Not, s.Else, s.Then, s.Contains,
		s.PropertyNames, s.UnevaluatedItems,
		s.UnevaluatedProperties, s.Items2020,
	)

	for _, schema := range s.Properties {
		walk(schema)
	}

	for _, schema := range s.DependentSchemas {
		walk(schema)
	}

	for _, schema := range s.PatternProperties {
		walk(schema)
	}

	switch s.Items.(type) {
	case *jsonschema.Schema:
		walk(s.Items.(*jsonschema.Schema))
	case []*jsonschema.Schema:
		walkEach(s.Items.([]*jsonschema.Schema)...)
	}
}
