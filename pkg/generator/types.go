package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// TypeGenerator generates Go types from OpenAPI schemas
type TypeGenerator struct {
	spec      *openapi.Document
	generated map[string]bool
	usesTime  bool // tracks if time.Time is used
	usesDate  bool // tracks if date.Date is used
}

// NewTypeGenerator creates a new TypeGenerator instance
func NewTypeGenerator(spec *openapi.Document) *TypeGenerator {
	return &TypeGenerator{
		spec:      spec,
		generated: make(map[string]bool),
	}
}

// Generate generates Go type definitions from the OpenAPI spec
func (g *TypeGenerator) Generate() (string, error) {
	if g.spec.Components == nil || g.spec.Components.Schemas == nil {
		return "package api\n", nil
	}

	// Build template data
	data := TypesTemplateData{
		Types: []TypeDefinition{},
	}

	// Sort schema names for deterministic output
	schemaNames := make([]string, 0, len(g.spec.Components.Schemas))
	for name := range g.spec.Components.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	// Generate type definitions
	for _, name := range schemaNames {
		schemaRef := g.spec.Components.Schemas[name]
		typeDef := g.buildTypeDefinition(name, schemaRef.Value)
		if typeDef != nil {
			data.Types = append(data.Types, *typeDef)
		}
	}

	// Set import flags
	data.UsesTime = g.usesTime
	data.UsesDate = g.usesDate

	// Execute template
	return executeTemplate("types.go.tmpl", data)
}

// buildTypeDefinition builds a TypeDefinition from an OpenAPI schema
func (g *TypeGenerator) buildTypeDefinition(name string, schema *openapi.Schema) *TypeDefinition {
	if g.generated[name] {
		return nil
	}
	g.generated[name] = true

	// If schema is nil, this is a reference-only schema (alias)
	if schema == nil {
		return nil
	}

	typeName := toGoTypeName(name)

	def := &TypeDefinition{
		Name:        typeName,
		Description: schema.Description,
	}

	schemaType := getSchemaType(schema)

	switch schemaType {
	case "object", "":
		def.Kind = "struct"
		def.Fields = g.buildStructFields(schema)
	case "string":
		if len(schema.Enum) > 0 {
			def.Kind = "enum"
			def.EnumValues = g.buildEnumValues(typeName, schema)
		} else {
			def.Kind = "alias"
			def.AliasType = "string"
		}
	case "integer", "number":
		def.Kind = "alias"
		def.AliasType = mapOpenAPITypeToGo(schema)
	case "boolean":
		def.Kind = "alias"
		def.AliasType = "bool"
	case "array":
		if schema.Items != nil {
			def.Kind = "array"
			def.ItemType = g.resolveType(schema.Items.Value)
		}
	}

	return def
}

// buildStructFields builds field definitions for a struct
func (g *TypeGenerator) buildStructFields(schema *openapi.Schema) []FieldDefinition {
	if schema.Properties == nil {
		return nil
	}

	// Sort property names for deterministic output
	propNames := make([]string, 0, len(schema.Properties))
	for propName := range schema.Properties {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	fields := make([]FieldDefinition, 0, len(propNames))
	for _, propName := range propNames {
		propRef := schema.Properties[propName]
		propSchema := propRef.Value

		fieldName := toGoFieldName(propName)
		fieldType := g.resolveTypeWithRef(propRef)

		// Check if field is required
		isRequired := contains(schema.Required, propName)
		if !isRequired && !isPrimitiveType(fieldType) {
			fieldType = "*" + fieldType
		}

		// Build JSON tag
		jsonTag := propName
		if !isRequired {
			jsonTag += ",omitempty"
		}

		field := FieldDefinition{
			Name:    fieldName,
			Type:    fieldType,
			JSONTag: jsonTag,
		}

		// Add description if available
		if propSchema != nil && propSchema.Description != "" {
			field.Description = propSchema.Description
		}

		fields = append(fields, field)
	}

	return fields
}

// buildEnumValues builds enum value definitions
func (g *TypeGenerator) buildEnumValues(typeName string, schema *openapi.Schema) []EnumValue {
	values := make([]EnumValue, 0, len(schema.Enum))
	for _, value := range schema.Enum {
		if strVal, ok := value.(string); ok {
			values = append(values, EnumValue{
				ConstName: toGoConstName(typeName, strVal),
				Value:     strVal,
			})
		}
	}
	return values
}

// resolveTypeWithRef resolves the Go type from a schema reference
func (g *TypeGenerator) resolveTypeWithRef(ref *openapi.SchemaRef) string {
	if ref == nil {
		return "any"
	}

	// If this is a reference to a component schema, extract the type name
	if ref.Ref != "" {
		// Extract type name from reference like "#/components/schemas/Owner"
		parts := strings.Split(ref.Ref, "/")
		if len(parts) > 0 {
			typeName := parts[len(parts)-1]
			return toGoTypeName(typeName)
		}
	}

	return g.resolveType(ref.Value)
}

// resolveType resolves the Go type for an OpenAPI schema
func (g *TypeGenerator) resolveType(schema *openapi.Schema) string {
	if schema == nil {
		return "any"
	}

	schemaType := getSchemaType(schema)

	switch schemaType {
	case "object", "":
		if len(schema.Properties) > 0 {
			return "map[string]any"
		}
		return "any"
	case "array":
		if schema.Items != nil {
			itemType := g.resolveTypeWithRef(schema.Items)
			return "[]" + itemType
		}
		return "[]any"
	case "string":
		if schema.Format == "date-time" {
			g.usesTime = true
			return "time.Time"
		}
		if schema.Format == "date" {
			g.usesDate = true
			return "date.Date"
		}
		return "string"
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}

// mapOpenAPITypeToGo maps OpenAPI types to Go types
func mapOpenAPITypeToGo(schema *openapi.Schema) string {
	schemaType := getSchemaType(schema)

	switch schemaType {
	case "string":
		return "string"
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}

// getSchemaType extracts the type from an OpenAPI schema
// Handles both OpenAPI 3.0 (single type) and 3.1+ (array of types)
func getSchemaType(schema *openapi.Schema) string {
	if schema == nil {
		return ""
	}

	return schema.GetSchemaType()
}

// Helper functions

func toGoTypeName(name string) string {
	return toPascalCase(name)
}

func toGoFieldName(name string) string {
	return toPascalCase(name)
}

func toGoConstName(typeName, value string) string {
	return typeName + toPascalCase(value)
}

func toPascalCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, "")
}

func splitWords(s string) []string {
	// Split by common separators first
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, " ", "_")

	parts := strings.Split(s, "_")
	var words []string

	// Further split camelCase/PascalCase words
	for _, part := range parts {
		if part == "" {
			continue
		}
		words = append(words, splitCamelCase(part)...)
	}

	return words
}

// splitCamelCase splits a camelCase or PascalCase string into words
func splitCamelCase(s string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Found uppercase letter, start new word
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isPrimitiveType(t string) bool {
	primitives := []string{"string", "int", "int32", "int64", "float32", "float64", "bool", "byte"}
	for _, p := range primitives {
		if t == p {
			return true
		}
	}
	return false
}

// Ensure the Generate method returns an error for template issues
func (g *TypeGenerator) generateError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
