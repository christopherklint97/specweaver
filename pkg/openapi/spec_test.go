package openapi

import (
	"testing"
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

		if doc.OpenAPI != "3.1.0" {
			t.Errorf("Expected OpenAPI '3.1.0', got %s", doc.OpenAPI)
		}

		if doc.Info.Title != "Test API" {
			t.Errorf("Expected title 'Test API', got %s", doc.Info.Title)
		}
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

		if doc.Components == nil {
			t.Error("Expected components to be set")
		}

		if len(doc.Components.Schemas) != 1 {
			t.Errorf("Expected 1 schema, got %d", len(doc.Components.Schemas))
		}
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

		if info.Title != "Complete API" {
			t.Errorf("Expected title 'Complete API', got %s", info.Title)
		}

		if info.Contact == nil {
			t.Error("Expected contact to be set")
		}

		if info.Contact.Email != "team@example.com" {
			t.Errorf("Expected email 'team@example.com', got %s", info.Contact.Email)
		}

		if info.License == nil {
			t.Error("Expected license to be set")
		}

		if info.License.Name != "MIT" {
			t.Errorf("Expected license 'MIT', got %s", info.License.Name)
		}
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

		if pathItem.Get == nil {
			t.Error("Expected GET operation to be set")
		}

		if pathItem.Get.OperationID != "getItem" {
			t.Errorf("Expected operation ID 'getItem', got %s", pathItem.Get.OperationID)
		}

		if pathItem.Post == nil {
			t.Error("Expected POST operation to be set")
		}
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

		if op.OperationID != "testOp" {
			t.Errorf("Expected operation ID 'testOp', got %s", op.OperationID)
		}

		if len(op.Parameters) != 1 {
			t.Errorf("Expected 1 parameter, got %d", len(op.Parameters))
		}

		if op.Parameters[0].Name != "id" {
			t.Errorf("Expected parameter name 'id', got %s", op.Parameters[0].Name)
		}

		if len(op.Responses) != 1 {
			t.Errorf("Expected 1 response, got %d", len(op.Responses))
		}
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

		if op.RequestBody == nil {
			t.Error("Expected request body to be set")
		}

		if !op.RequestBody.Required {
			t.Error("Expected request body to be required")
		}

		if len(op.RequestBody.Content) != 1 {
			t.Errorf("Expected 1 media type, got %d", len(op.RequestBody.Content))
		}
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

		if schema.GetSchemaType() != "string" {
			t.Errorf("Expected type 'string', got %s", schema.GetSchemaType())
		}

		if *schema.MinLength != 1 {
			t.Errorf("Expected minLength 1, got %d", *schema.MinLength)
		}

		if *schema.MaxLength != 100 {
			t.Errorf("Expected maxLength 100, got %d", *schema.MaxLength)
		}
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

		if schema.GetSchemaType() != "array" {
			t.Errorf("Expected type 'array', got %s", schema.GetSchemaType())
		}

		if schema.Items == nil {
			t.Error("Expected items to be set")
		}

		if *schema.MinItems != 1 {
			t.Errorf("Expected minItems 1, got %d", *schema.MinItems)
		}
	})

	t.Run("Number schema with validation", func(t *testing.T) {
		schema := &Schema{
			Type:     []string{"number"},
			Format:   "float",
			Minimum:  float64Ptr(0.0),
			Maximum:  float64Ptr(100.0),
			MultipleOf: float64Ptr(0.5),
		}

		if schema.GetSchemaType() != "number" {
			t.Errorf("Expected type 'number', got %s", schema.GetSchemaType())
		}

		if schema.Format != "float" {
			t.Errorf("Expected format 'float', got %s", schema.Format)
		}

		if *schema.Minimum != 0.0 {
			t.Errorf("Expected minimum 0.0, got %f", *schema.Minimum)
		}
	})

	t.Run("Enum schema", func(t *testing.T) {
		schema := &Schema{
			Type: []string{"string"},
			Enum: []any{"available", "pending", "sold"},
		}

		if len(schema.Enum) != 3 {
			t.Errorf("Expected 3 enum values, got %d", len(schema.Enum))
		}
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

		if len(schema.AllOf) != 2 {
			t.Errorf("Expected 2 allOf schemas, got %d", len(schema.AllOf))
		}
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

		if param.Name != "id" {
			t.Errorf("Expected name 'id', got %s", param.Name)
		}

		if param.In != "path" {
			t.Errorf("Expected in 'path', got %s", param.In)
		}

		if !param.Required {
			t.Error("Expected parameter to be required")
		}
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

		if param.In != "query" {
			t.Errorf("Expected in 'query', got %s", param.In)
		}

		if param.Required {
			t.Error("Expected parameter to be optional")
		}
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

		if len(components.Schemas) != 1 {
			t.Errorf("Expected 1 schema, got %d", len(components.Schemas))
		}

		if len(components.Responses) != 1 {
			t.Errorf("Expected 1 response, got %d", len(components.Responses))
		}

		if len(components.SecuritySchemes) != 1 {
			t.Errorf("Expected 1 security scheme, got %d", len(components.SecuritySchemes))
		}
	})
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
