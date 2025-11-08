package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentStructure(t *testing.T) {
	t.Run("Create basic document", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths:    make(Paths),
			refCache: make(map[string]any),
		}

		assert.Equal(t, "3.1.0", doc.OpenAPI)
		assert.Equal(t, "Test API", doc.Info.Title)
	})

	t.Run("Create document with components", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Components: &Components{
				Schemas: map[string]*SchemaRef{
					"Pet": {
						Value: &Schema{
							Type: []string{"object"},
						},
					},
				},
			},
			refCache: make(map[string]any),
		}

		assert.NotNil(t, doc.Components)
		assert.Len(t, doc.Components.Schemas, 1)
	})
}

func TestInfoStructure(t *testing.T) {
	t.Run("Info with all fields", func(t *testing.T) {
		info := &Info{
			Title:       "Complete API",
			Version:     "2.0.0",
			Description: "A complete API description",
			Contact: &Contact{
				Name:  "API Team",
				Email: "team@example.com",
				URL:   "https://example.com",
			},
			License: &License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		}

		assert.Equal(t, "Complete API", info.Title)
		assert.NotNil(t, info.Contact)
		assert.Equal(t, "team@example.com", info.Contact.Email)
		assert.NotNil(t, info.License)
		assert.Equal(t, "MIT", info.License.Name)
	})
}

func TestPathItemStructure(t *testing.T) {
	t.Run("PathItem with multiple operations", func(t *testing.T) {
		pathItem := &PathItem{
			Get: &Operation{
				OperationID: "getItem",
				Summary:     "Get an item",
			},
			Post: &Operation{
				OperationID: "createItem",
				Summary:     "Create an item",
			},
		}

		assert.NotNil(t, pathItem.Get)
		assert.Equal(t, "getItem", pathItem.Get.OperationID)
		assert.NotNil(t, pathItem.Post)
	})
}

func TestOperationStructure(t *testing.T) {
	t.Run("Operation with parameters and responses", func(t *testing.T) {
		op := &Operation{
			OperationID: "testOp",
			Summary:     "Test operation",
			Description: "A test operation",
			Parameters: []*Parameter{
				{
					Name:     "id",
					In:       "path",
					Required: true,
					Schema: &SchemaRef{
						Value: &Schema{
							Type: []string{"integer"},
						},
					},
				},
			},
			Responses: Responses{
				"200": {
					Description: "Success",
				},
			},
		}

		assert.Equal(t, "testOp", op.OperationID)
		assert.Len(t, op.Parameters, 1)
		assert.Equal(t, "id", op.Parameters[0].Name)
		assert.Len(t, op.Responses, 1)
	})

	t.Run("Operation with request body", func(t *testing.T) {
		op := &Operation{
			OperationID: "create",
			RequestBody: &RequestBody{
				Description: "Request body",
				Required:    true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: &SchemaRef{
							Ref: "#/components/schemas/Pet",
						},
					},
				},
			},
		}

		assert.NotNil(t, op.RequestBody)
		assert.True(t, op.RequestBody.Required)
		assert.Len(t, op.RequestBody.Content, 1)
	})
}

func TestSchemaStructure(t *testing.T) {
	t.Run("Simple string schema", func(t *testing.T) {
		schema := &Schema{
			Type:        []string{"string"},
			Description: "A string field",
			MinLength:   intPtr(1),
			MaxLength:   intPtr(100),
		}

		assert.Equal(t, "string", schema.GetSchemaType())
		assert.Equal(t, 1, *schema.MinLength)
		assert.Equal(t, 100, *schema.MaxLength)
	})

	t.Run("Object schema with properties", func(t *testing.T) {
		schema := &Schema{
			Type:        []string{"object"},
			Description: "An object",
			Properties: map[string]*SchemaRef{
				"id": {
					Value: &Schema{
						Type: []string{"integer"},
					},
				},
				"name": {
					Value: &Schema{
						Type: []string{"string"},
					},
				},
			},
			Required: []string{"id", "name"},
		}

		assert.Equal(t, "object", schema.GetSchemaType())
		assert.Len(t, schema.Properties, 2)
		assert.Len(t, schema.Required, 2)
	})

	t.Run("Array schema", func(t *testing.T) {
		schema := &Schema{
			Type: []string{"array"},
			Items: &SchemaRef{
				Value: &Schema{
					Type: []string{"string"},
				},
			},
			MinItems: intPtr(1),
			MaxItems: intPtr(10),
		}

		assert.Equal(t, "array", schema.GetSchemaType())
		assert.NotNil(t, schema.Items)
		assert.Equal(t, 1, *schema.MinItems)
	})

	t.Run("Number schema with validation", func(t *testing.T) {
		schema := &Schema{
			Type:       []string{"number"},
			Format:     "float",
			Minimum:    float64Ptr(0.0),
			Maximum:    float64Ptr(100.0),
			MultipleOf: float64Ptr(0.5),
		}

		assert.Equal(t, "number", schema.GetSchemaType())
		assert.Equal(t, "float", schema.Format)
		assert.Equal(t, 0.0, *schema.Minimum)
	})

	t.Run("Enum schema", func(t *testing.T) {
		schema := &Schema{
			Type: []string{"string"},
			Enum: []any{"available", "pending", "sold"},
		}

		assert.Len(t, schema.Enum, 3)
	})

	t.Run("Schema with composition (allOf)", func(t *testing.T) {
		schema := &Schema{
			AllOf: []*SchemaRef{
				{Ref: "#/components/schemas/Base"},
				{
					Value: &Schema{
						Properties: map[string]*SchemaRef{
							"extraField": {
								Value: &Schema{
									Type: []string{"string"},
								},
							},
						},
					},
				},
			},
		}

		assert.Len(t, schema.AllOf, 2)
	})
}

func TestParameterStructure(t *testing.T) {
	t.Run("Path parameter", func(t *testing.T) {
		param := &Parameter{
			Name:     "id",
			In:       "path",
			Required: true,
			Schema: &SchemaRef{
				Value: &Schema{
					Type: []string{"integer"},
				},
			},
		}

		assert.Equal(t, "id", param.Name)
		assert.Equal(t, "path", param.In)
		assert.True(t, param.Required)
	})

	t.Run("Query parameter", func(t *testing.T) {
		param := &Parameter{
			Name:        "limit",
			In:          "query",
			Required:    false,
			Description: "Limit results",
			Schema: &SchemaRef{
				Value: &Schema{
					Type: []string{"integer"},
				},
			},
		}

		assert.Equal(t, "query", param.In)
		assert.False(t, param.Required)
	})
}

func TestComponentsStructure(t *testing.T) {
	t.Run("Components with schemas and responses", func(t *testing.T) {
		components := &Components{
			Schemas: map[string]*SchemaRef{
				"Pet": {
					Value: &Schema{
						Type: []string{"object"},
					},
				},
			},
			Responses: map[string]*Response{
				"NotFound": {
					Description: "Not found",
				},
			},
			SecuritySchemes: map[string]*SecurityScheme{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
				},
			},
		}

		assert.Len(t, components.Schemas, 1)
		assert.Len(t, components.Responses, 1)
		assert.Len(t, components.SecuritySchemes, 1)
	})
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
