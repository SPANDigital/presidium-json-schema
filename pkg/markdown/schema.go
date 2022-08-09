package markdown

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
	log "github.com/sirupsen/logrus"
)

type RawSchema map[string]interface{}

type Schema struct {

	// the Path of the json file this schemas belongs to
	Path string

	*jsonschema.Schema
}

func ToSchema(s *jsonschema.Schema, path string) *Schema {
	if s == nil {
		return nil
	}

	return &Schema{path, s}
}

func (r RawSchema) Id() string {
	id := r["$id"]
	if _, ok := id.(string); ok {
		return id.(string)
	}
	return ""
}

// Definitions returns all references from the current schema
func (s *Schema) Definitions() []*Schema {
	var definitions []*Schema
	s.WalkSchema(true, func(next *Schema) error {
		if next.Ref != nil && next.Location != s.Location {
			log.Debugf("found definition: %s", s.Ref)
			definitions = append(definitions, ToSchema(next.Ref, s.Path))
		}
		return nil
	})
	return definitions
}

// WalkSchema walks the schema tree, calling fn for each schema in the tree, including root.
func (s *Schema) WalkSchema(followRef bool, fn func(s *Schema) error) {
	s.walkSchema(followRef, map[string]bool{}, fn)
}

func (s *Schema) walkSchema(followRef bool, visited map[string]bool, fn func(s *Schema) error) {
	if s == nil || visited[s.Location] {
		return
	}

	if err := fn(s); err != nil {
		return
	}

	visited[s.Location] = true
	walk := func(schema *jsonschema.Schema) {
		ToSchema(schema, s.Path).walkSchema(followRef, visited, fn)
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
