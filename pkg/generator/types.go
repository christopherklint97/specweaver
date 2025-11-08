package generator

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// TypeGenerator generates Go types from OpenAPI schemas
type TypeGenerator struct {
	spec      *openapi3.T
	generated map[string]bool
}

// NewTypeGenerator creates a new TypeGenerator instance
func NewTypeGenerator(spec *openapi3.T) *TypeGenerator {
	return &TypeGenerator{
		spec:      spec,
		generated: make(map[string]bool),
	}
}

// Generate generates Go type definitions from the OpenAPI spec
func (g *TypeGenerator) Generate() (string, error) {
	var sb strings.Builder

	sb.WriteString("package api\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"time\"\n")
	sb.WriteString(")\n\n")

	if g.spec.Components == nil || g.spec.Components.Schemas == nil {
		return sb.String(), nil
	}

	// Generate types for all schemas in components
	for name, schemaRef := range g.spec.Components.Schemas {
		if err := g.generateType(&sb, name, schemaRef.Value); err != nil {
			return "", fmt.Errorf("failed to generate type for %s: %w", name, err)
		}
	}

	return sb.String(), nil
}

// generateType generates a Go type from an OpenAPI schema
func (g *TypeGenerator) generateType(sb *strings.Builder, name string, schema *openapi3.Schema) error {
	if g.generated[name] {
		return nil
	}
	g.generated[name] = true

	typeName := toGoTypeName(name)

	// Add description as comment if available
	if schema.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s %s\n", typeName, schema.Description))
	}

	schemaType := getSchemaType(schema)

	switch schemaType {
	case "object", "":
		g.generateStruct(sb, typeName, schema)
	case "string":
		if len(schema.Enum) > 0 {
			g.generateEnum(sb, typeName, schema)
		} else {
			sb.WriteString(fmt.Sprintf("type %s string\n\n", typeName))
		}
	case "integer", "number":
		goType := mapOpenAPITypeToGo(schema)
		sb.WriteString(fmt.Sprintf("type %s %s\n\n", typeName, goType))
	case "boolean":
		sb.WriteString(fmt.Sprintf("type %s bool\n\n", typeName))
	case "array":
		if schema.Items != nil {
			itemType := g.resolveType(schema.Items.Value)
			sb.WriteString(fmt.Sprintf("type %s []%s\n\n", typeName, itemType))
		}
	}

	return nil
}

// generateStruct generates a Go struct from an object schema
func (g *TypeGenerator) generateStruct(sb *strings.Builder, name string, schema *openapi3.Schema) {
	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))

	if schema.Properties != nil {
		for propName, propRef := range schema.Properties {
			propSchema := propRef.Value
			fieldName := toGoFieldName(propName)

			// Check if this is a reference to a component schema
			fieldType := g.resolveTypeWithRef(propRef)

			// Check if field is required
			isRequired := contains(schema.Required, propName)
			if !isRequired && !isPrimitiveType(fieldType) {
				fieldType = "*" + fieldType
			}

			// Add JSON tags
			jsonTag := propName
			if !isRequired {
				jsonTag += ",omitempty"
			}

			// Add field comment if description exists
			if propSchema.Description != "" {
				sb.WriteString(fmt.Sprintf("\t// %s\n", propSchema.Description))
			}

			sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, fieldType, jsonTag))
		}
	}

	sb.WriteString("}\n\n")
}

// generateEnum generates Go constants for enum values
func (g *TypeGenerator) generateEnum(sb *strings.Builder, name string, schema *openapi3.Schema) {
	sb.WriteString(fmt.Sprintf("type %s string\n\n", name))
	sb.WriteString("const (\n")

	for _, value := range schema.Enum {
		if strVal, ok := value.(string); ok {
			constName := toGoConstName(name, strVal)
			sb.WriteString(fmt.Sprintf("\t%s %s = \"%s\"\n", constName, name, strVal))
		}
	}

	sb.WriteString(")\n\n")
}

// resolveTypeWithRef resolves the Go type from a schema reference
func (g *TypeGenerator) resolveTypeWithRef(ref *openapi3.SchemaRef) string {
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
func (g *TypeGenerator) resolveType(schema *openapi3.Schema) string {
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
			return "time.Time"
		}
		if schema.Format == "date" {
			return "string" // or custom Date type
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
func mapOpenAPITypeToGo(schema *openapi3.Schema) string {
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
// Handles both string types (older versions) and *openapi3.Types (newer versions)
func getSchemaType(schema *openapi3.Schema) string {
	if schema == nil {
		return ""
	}

	// In newer versions of kin-openapi, Type is *openapi3.Types (a slice)
	if schema.Type != nil {
		types := *schema.Type
		if len(types) > 0 {
			return types[0]
		}
	}

	return ""
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
