package generator

import (
	"strings"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// getParamType returns the Go type for a parameter
func getParamType(param *openapi.Parameter) string {
	if param.Schema == nil || param.Schema.Value == nil {
		return "string"
	}

	schema := param.Schema.Value
	schemaType := schema.GetSchemaType()

	switch schemaType {
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		} else if schema.Format == "int32" {
			return "int32"
		}
		return "int"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	case "string":
		return "string"
	default:
		return "string"
	}
}

// resolveSchemaType resolves a schema reference to a Go type
func resolveSchemaType(schemaRef *openapi.SchemaRef) string {
	if schemaRef == nil {
		return "any"
	}

	// If this is a reference, extract the type name
	if schemaRef.Ref != "" {
		parts := strings.Split(schemaRef.Ref, "/")
		if len(parts) > 0 {
			typeName := parts[len(parts)-1]
			return toPascalCase(typeName)
		}
	}

	// Otherwise resolve from schema
	if schemaRef.Value != nil {
		return resolveSchemaTypeFromValue(schemaRef.Value)
	}

	return "any"
}

// resolveSchemaTypeFromValue resolves the Go type from a schema value
func resolveSchemaTypeFromValue(schema *openapi.Schema) string {
	if schema == nil {
		return "any"
	}

	schemaType := schema.GetSchemaType()

	switch schemaType {
	case "array":
		if schema.Items != nil {
			itemType := resolveSchemaType(schema.Items)
			return "[]" + itemType
		}
		return "[]any"
	case "object":
		return "map[string]any"
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
