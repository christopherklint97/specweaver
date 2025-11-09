package openapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSchemaUnmarshalYAML(t *testing.T) {
	t.Run("Single type as string (OpenAPI 3.0)", func(t *testing.T) {
		yamlData := `type: string
description: A string field`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		require.NoError(t, err)

		assert.Len(t, schema.Type, 1)
		assert.Equal(t, "string", schema.Type[0])
		assert.Equal(t, "A string field", schema.Description)
	})

	t.Run("Type as array (OpenAPI 3.1)", func(t *testing.T) {
		yamlData := `type: [string, "null"]
description: A nullable string`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		require.NoError(t, err)

		assert.Len(t, schema.Type, 2)
		assert.Equal(t, "string", schema.Type[0])
		assert.Equal(t, "null", schema.Type[1])
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
		require.NoError(t, err)

		assert.Equal(t, "object", schema.GetSchemaType())
		assert.Len(t, schema.Properties, 2)
		assert.Len(t, schema.Required, 1)
		assert.Equal(t, "name", schema.Required[0])
	})

	t.Run("Schema with enum", func(t *testing.T) {
		yamlData := `type: string
enum:
  - available
  - pending
  - sold`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		require.NoError(t, err)

		assert.Len(t, schema.Enum, 3)
	})

	t.Run("Invalid type format", func(t *testing.T) {
		yamlData := `type: 123`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		assert.Error(t, err)
	})

	t.Run("Schema without type", func(t *testing.T) {
		yamlData := `description: A schema without explicit type
properties:
  id:
    type: integer`

		var schema Schema
		err := yaml.Unmarshal([]byte(yamlData), &schema)
		require.NoError(t, err)

		// Schema without type is valid (inferred from properties)
		assert.Empty(t, schema.Type)
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
		require.NoError(t, err)

		assert.Len(t, schema.Type, 1)
		assert.Equal(t, "string", schema.Type[0])
	})

	t.Run("Type as array (OpenAPI 3.1)", func(t *testing.T) {
		jsonData := `{
			"type": ["string", "null"],
			"description": "A nullable string"
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		require.NoError(t, err)

		assert.Len(t, schema.Type, 2)
		assert.Equal(t, "string", schema.Type[0])
		assert.Equal(t, "null", schema.Type[1])
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
		require.NoError(t, err)

		assert.Equal(t, "object", schema.GetSchemaType())
		assert.Len(t, schema.Properties, 2)
		assert.Len(t, schema.Required, 2)
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
		require.NoError(t, err)

		assert.Equal(t, "array", schema.GetSchemaType())
		assert.NotNil(t, schema.Items)
	})

	t.Run("Invalid type format", func(t *testing.T) {
		jsonData := `{
			"type": 123
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		assert.Error(t, err)
	})
}

func TestHandleTypeField(t *testing.T) {
	t.Run("Handle string type", func(t *testing.T) {
		data := json.RawMessage(`"string"`)
		var schema Schema

		err := handleTypeField(data, &schema)
		require.NoError(t, err)

		assert.Len(t, schema.Type, 1)
		assert.Equal(t, "string", schema.Type[0])
	})

	t.Run("Handle array type", func(t *testing.T) {
		data := json.RawMessage(`["string", "null"]`)
		var schema Schema

		err := handleTypeField(data, &schema)
		require.NoError(t, err)

		assert.Len(t, schema.Type, 2)
	})

	t.Run("Handle invalid type", func(t *testing.T) {
		data := json.RawMessage(`123`)
		var schema Schema

		err := handleTypeField(data, &schema)
		assert.Error(t, err)
	})
}

func TestSchemaGetSchemaType(t *testing.T) {
	t.Run("Single type", func(t *testing.T) {
		schema := &Schema{
			Type: []string{"string"},
		}

		assert.Equal(t, "string", schema.GetSchemaType())
	})

	t.Run("Multiple types", func(t *testing.T) {
		schema := &Schema{
			Type: []string{"string", "null"},
		}

		// Should return the first type
		assert.Equal(t, "string", schema.GetSchemaType())
	})

	t.Run("No type", func(t *testing.T) {
		schema := &Schema{
			Type: []string{},
		}

		assert.Empty(t, schema.GetSchemaType())
	})

	t.Run("Nil schema", func(t *testing.T) {
		var schema *Schema

		assert.Empty(t, schema.GetSchemaType())
	})
}

func TestSchemaRefIsRefOnly(t *testing.T) {
	t.Run("Reference only", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/Pet",
		}

		assert.True(t, ref.IsRefOnly())
	})

	t.Run("Value with reference", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/Pet",
			Value: &Schema{
				Type: []string{"object"},
			},
		}

		assert.True(t, ref.IsRefOnly(), "Expected IsRefOnly to be true when Ref is set")
	})

	t.Run("Value only", func(t *testing.T) {
		ref := &SchemaRef{
			Value: &Schema{
				Type: []string{"string"},
			},
		}

		assert.False(t, ref.IsRefOnly())
	})

	t.Run("Nil reference", func(t *testing.T) {
		var ref *SchemaRef

		assert.False(t, ref.IsRefOnly())
	})
}
