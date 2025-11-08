package openapi

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSchemaUnmarshalYAML(t *testing.T) {
	t.Run("Single type as string (OpenAPI 3.0)", func(t *testing.T) {
		yamlData := `type: string
description: A string field`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(schema.Type) != 1 || schema.Type[0] != "string" {
			t.Errorf("Expected type ['string'], got %v", schema.Type)
		}

		if schema.Description != "A string field" {
			t.Errorf("Expected description 'A string field', got %s", schema.Description)
		}
	})

	t.Run("Type as array (OpenAPI 3.1)", func(t *testing.T) {
		yamlData := `type: [string, "null"]
description: A nullable string`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(schema.Type) != 2 {
			t.Errorf("Expected 2 types, got %d", len(schema.Type))
		}

		if schema.Type[0] != "string" || schema.Type[1] != "null" {
			t.Errorf("Expected types ['string', 'null'], got %v", schema.Type)
		}
	})

	t.Run("Object schema with properties", func(t *testing.T) {
		yamlData := `type: object
properties:
  name:
    type: string
  age:
    type: integer
required:
  - name`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if schema.GetSchemaType() != "object" {
			t.Errorf("Expected type 'object', got %s", schema.GetSchemaType())
		}

		if len(schema.Properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
		}

		if len(schema.Required) != 1 || schema.Required[0] != "name" {
			t.Errorf("Expected required ['name'], got %v", schema.Required)
		}
	})

	t.Run("Schema with enum", func(t *testing.T) {
		yamlData := `type: string
enum:
  - available
  - pending
  - sold`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(schema.Enum) != 3 {
			t.Errorf("Expected 3 enum values, got %d", len(schema.Enum))
		}
	})

	t.Run("Invalid type format", func(t *testing.T) {
		yamlData := `type: 123`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		if err == nil {
			t.Error("Expected error for invalid type format")
		}
	})

	t.Run("Schema without type", func(t *testing.T) {
		yamlData := `description: A schema without explicit type
properties:
  id:
    type: integer`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Schema without type is valid (inferred from properties)
		if len(schema.Type) != 0 {
			t.Errorf("Expected empty type array, got %v", schema.Type)
		}
	})
}

func TestSchemaUnmarshalJSON(t *testing.T) {
	t.Run("Single type as string (OpenAPI 3.0)", func(t *testing.T) {
		jsonData := `{
			"type": "string",
			"description": "A string field"
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(schema.Type) != 1 || schema.Type[0] != "string" {
			t.Errorf("Expected type ['string'], got %v", schema.Type)
		}
	})

	t.Run("Type as array (OpenAPI 3.1)", func(t *testing.T) {
		jsonData := `{
			"type": ["string", "null"],
			"description": "A nullable string"
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(schema.Type) != 2 {
			t.Errorf("Expected 2 types, got %d", len(schema.Type))
		}

		if schema.Type[0] != "string" || schema.Type[1] != "null" {
			t.Errorf("Expected types ['string', 'null'], got %v", schema.Type)
		}
	})

	t.Run("Object schema", func(t *testing.T) {
		jsonData := `{
			"type": "object",
			"properties": {
				"id": {
					"type": "integer",
					"format": "int64"
				},
				"name": {
					"type": "string"
				}
			},
			"required": ["id", "name"]
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if schema.GetSchemaType() != "object" {
			t.Errorf("Expected type 'object', got %s", schema.GetSchemaType())
		}

		if len(schema.Properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
		}

		if len(schema.Required) != 2 {
			t.Errorf("Expected 2 required fields, got %d", len(schema.Required))
		}
	})

	t.Run("Array schema", func(t *testing.T) {
		jsonData := `{
			"type": "array",
			"items": {
				"type": "string"
			}
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if schema.GetSchemaType() != "array" {
			t.Errorf("Expected type 'array', got %s", schema.GetSchemaType())
		}

		// Items should be set
		if schema.Items == nil {
			t.Error("Expected items to be set")
		}
	})

	t.Run("Invalid type format", func(t *testing.T) {
		jsonData := `{
			"type": 123
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		if err == nil {
			t.Error("Expected error for invalid type format")
		}
	})
}

func TestHandleTypeField(t *testing.T) {
	t.Run("Handle string type", func(t *testing.T) {
		data := json.RawMessage(`"string"`)
		var schema Schema

		err := handleTypeField(data, &schema)
		if err != nil {
			t.Fatalf("handleTypeField failed: %v", err)
		}

		if len(schema.Type) != 1 || schema.Type[0] != "string" {
			t.Errorf("Expected type ['string'], got %v", schema.Type)
		}
	})

	t.Run("Handle array type", func(t *testing.T) {
		data := json.RawMessage(`["string", "null"]`)
		var schema Schema

		err := handleTypeField(data, &schema)
		if err != nil {
			t.Fatalf("handleTypeField failed: %v", err)
		}

		if len(schema.Type) != 2 {
			t.Errorf("Expected 2 types, got %d", len(schema.Type))
		}
	})

	t.Run("Handle invalid type", func(t *testing.T) {
		data := json.RawMessage(`123`)
		var schema Schema

		err := handleTypeField(data, &schema)
		if err == nil {
			t.Error("Expected error for invalid type")
		}
	})
}

func TestSchemaGetSchemaType(t *testing.T) {
	t.Run("Single type", func(t *testing.T) {
		schema := &Schema{
			Type: []string{"string"},
		}

		if schema.GetSchemaType() != "string" {
			t.Errorf("Expected 'string', got %s", schema.GetSchemaType())
		}
	})

	t.Run("Multiple types", func(t *testing.T) {
		schema := &Schema{
			Type: []string{"string", "null"},
		}

		// Should return the first type
		if schema.GetSchemaType() != "string" {
			t.Errorf("Expected 'string', got %s", schema.GetSchemaType())
		}
	})

	t.Run("No type", func(t *testing.T) {
		schema := &Schema{
			Type: []string{},
		}

		if schema.GetSchemaType() != "" {
			t.Errorf("Expected empty string, got %s", schema.GetSchemaType())
		}
	})

	t.Run("Nil schema", func(t *testing.T) {
		var schema *Schema

		if schema.GetSchemaType() != "" {
			t.Errorf("Expected empty string for nil schema, got %s", schema.GetSchemaType())
		}
	})
}

func TestSchemaRefIsRefOnly(t *testing.T) {
	t.Run("Reference only", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/Pet",
		}

		if !ref.IsRefOnly() {
			t.Error("Expected IsRefOnly to be true")
		}
	})

	t.Run("Value with reference", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/Pet",
			Value: &Schema{
				Type: []string{"object"},
			},
		}

		if !ref.IsRefOnly() {
			t.Error("Expected IsRefOnly to be true when Ref is set")
		}
	})

	t.Run("Value only", func(t *testing.T) {
		ref := &SchemaRef{
			Value: &Schema{
				Type: []string{"string"},
			},
		}

		if ref.IsRefOnly() {
			t.Error("Expected IsRefOnly to be false")
		}
	})

	t.Run("Nil reference", func(t *testing.T) {
		var ref *SchemaRef

		if ref.IsRefOnly() {
			t.Error("Expected IsRefOnly to be false for nil")
		}
	})
}
