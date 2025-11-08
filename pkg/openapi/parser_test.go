package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test.yaml")

	validYAML := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: Success
`

	if err := os.WriteFile(yamlPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("Load valid YAML file", func(t *testing.T) {
		doc, err := Load(yamlPath)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if doc.OpenAPI != "3.1.0" {
			t.Errorf("Expected OpenAPI version 3.1.0, got %s", doc.OpenAPI)
		}

		if doc.Info.Title != "Test API" {
			t.Errorf("Expected title 'Test API', got %s", doc.Info.Title)
		}

		if doc.Info.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %s", doc.Info.Version)
		}
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		_, err := Load("/nonexistent/file.yaml")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("Load invalid YAML", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		if err := os.WriteFile(invalidPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
			t.Fatalf("Failed to create invalid test file: %v", err)
		}

		_, err := Load(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})

	t.Run("Load JSON file", func(t *testing.T) {
		jsonPath := filepath.Join(tmpDir, "test.json")
		validJSON := `{
			"openapi": "3.0.0",
			"info": {
				"title": "JSON API",
				"version": "2.0.0"
			},
			"paths": {}
		}`

		if err := os.WriteFile(jsonPath, []byte(validJSON), 0644); err != nil {
			t.Fatalf("Failed to create JSON test file: %v", err)
		}

		doc, err := Load(jsonPath)
		if err != nil {
			t.Fatalf("Load JSON failed: %v", err)
		}

		if doc.Info.Title != "JSON API" {
			t.Errorf("Expected title 'JSON API', got %s", doc.Info.Title)
		}
	})
}

func TestLoadFromData(t *testing.T) {
	t.Run("Valid YAML data", func(t *testing.T) {
		data := []byte(`openapi: 3.1.0
info:
  title: Data API
  version: 1.0.0
paths: {}
`)

		doc, err := LoadFromData(data, "test.yaml")
		if err != nil {
			t.Fatalf("LoadFromData failed: %v", err)
		}

		if doc.Info.Title != "Data API" {
			t.Errorf("Expected title 'Data API', got %s", doc.Info.Title)
		}
	})

	t.Run("Valid JSON data", func(t *testing.T) {
		data := []byte(`{
			"openapi": "3.0.0",
			"info": {"title": "JSON Data API", "version": "1.0.0"},
			"paths": {}
		}`)

		doc, err := LoadFromData(data, "test.json")
		if err != nil {
			t.Fatalf("LoadFromData failed: %v", err)
		}

		if doc.Info.Title != "JSON Data API" {
			t.Errorf("Expected title 'JSON Data API', got %s", doc.Info.Title)
		}
	})
}

func TestValidateDocument(t *testing.T) {
	t.Run("Valid document", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test",
				Version: "1.0.0",
			},
			Paths: Paths{},
		}

		err := validateDocument(doc)
		if err != nil {
			t.Errorf("Validation failed for valid document: %v", err)
		}
	})

	t.Run("Nil document", func(t *testing.T) {
		err := validateDocument(nil)
		if err == nil {
			t.Error("Expected error for nil document")
		}
	})

	t.Run("Missing OpenAPI version", func(t *testing.T) {
		doc := &Document{
			Info: &Info{
				Title:   "Test",
				Version: "1.0.0",
			},
		}

		err := validateDocument(doc)
		if err == nil {
			t.Error("Expected error for missing OpenAPI version")
		}
	})

	t.Run("Unsupported OpenAPI version", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "2.0.0",
			Info: &Info{
				Title:   "Test",
				Version: "1.0.0",
			},
		}

		err := validateDocument(doc)
		if err == nil {
			t.Error("Expected error for unsupported OpenAPI version")
		}
	})

	t.Run("Missing info", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
		}

		err := validateDocument(doc)
		if err == nil {
			t.Error("Expected error for missing info")
		}
	})

	t.Run("Missing info.title", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Version: "1.0.0",
			},
		}

		err := validateDocument(doc)
		if err == nil {
			t.Error("Expected error for missing info.title")
		}
	})

	t.Run("Missing info.version", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title: "Test",
			},
		}

		err := validateDocument(doc)
		if err == nil {
			t.Error("Expected error for missing info.version")
		}
	})

	t.Run("Missing paths and components", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test",
				Version: "1.0.0",
			},
		}

		err := validateDocument(doc)
		if err == nil {
			t.Error("Expected error for missing paths and components")
		}
	})
}

func TestResolveSchemaRef(t *testing.T) {
	doc := &Document{
		OpenAPI: "3.1.0",
		Info: &Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*SchemaRef{
				"Pet": {
					Value: &Schema{
						Type: []string{"object"},
						Properties: map[string]*SchemaRef{
							"name": {
								Value: &Schema{
									Type: []string{"string"},
								},
							},
						},
					},
				},
			},
		},
		refCache: make(map[string]any),
	}

	t.Run("Resolve direct value", func(t *testing.T) {
		ref := &SchemaRef{
			Value: &Schema{
				Type: []string{"string"},
			},
		}

		schema, err := doc.ResolveSchemaRef(ref)
		if err != nil {
			t.Fatalf("ResolveSchemaRef failed: %v", err)
		}

		if schema.GetSchemaType() != "string" {
			t.Errorf("Expected type 'string', got %s", schema.GetSchemaType())
		}
	})

	t.Run("Resolve reference", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/Pet",
		}

		schema, err := doc.ResolveSchemaRef(ref)
		if err != nil {
			t.Fatalf("ResolveSchemaRef failed: %v", err)
		}

		if schema.GetSchemaType() != "object" {
			t.Errorf("Expected type 'object', got %s", schema.GetSchemaType())
		}
	})

	t.Run("Resolve nil reference", func(t *testing.T) {
		_, err := doc.ResolveSchemaRef(nil)
		if err == nil {
			t.Error("Expected error for nil reference")
		}
	})

	t.Run("Resolve invalid reference", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/NonExistent",
		}

		_, err := doc.ResolveSchemaRef(ref)
		if err == nil {
			t.Error("Expected error for invalid reference")
		}
	})
}

func TestGetSchemaByName(t *testing.T) {
	doc := &Document{
		OpenAPI: "3.1.0",
		Info: &Info{
			Title:   "Test",
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

	t.Run("Get existing schema", func(t *testing.T) {
		schema, err := doc.GetSchemaByName("Pet")
		if err != nil {
			t.Fatalf("GetSchemaByName failed: %v", err)
		}

		if schema.GetSchemaType() != "object" {
			t.Errorf("Expected type 'object', got %s", schema.GetSchemaType())
		}
	})

	t.Run("Get non-existent schema", func(t *testing.T) {
		_, err := doc.GetSchemaByName("NonExistent")
		if err == nil {
			t.Error("Expected error for non-existent schema")
		}
	})

	t.Run("No components", func(t *testing.T) {
		emptyDoc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test",
				Version: "1.0.0",
			},
			refCache: make(map[string]any),
		}

		_, err := emptyDoc.GetSchemaByName("Pet")
		if err == nil {
			t.Error("Expected error when components is nil")
		}
	})
}

func TestResolveReference(t *testing.T) {
	doc := &Document{
		OpenAPI: "3.1.0",
		Info: &Info{
			Title:   "Test",
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
			Responses: map[string]*Response{
				"NotFound": {
					Description: "Not found",
				},
			},
		},
		refCache: make(map[string]any),
	}

	t.Run("Resolve schema reference", func(t *testing.T) {
		obj, err := doc.resolveReference("#/components/schemas/Pet")
		if err != nil {
			t.Fatalf("resolveReference failed: %v", err)
		}

		schema, ok := obj.(*Schema)
		if !ok {
			t.Error("Expected *Schema type")
		}

		if schema.GetSchemaType() != "object" {
			t.Errorf("Expected type 'object', got %s", schema.GetSchemaType())
		}
	})

	t.Run("Resolve response reference", func(t *testing.T) {
		obj, err := doc.resolveReference("#/components/responses/NotFound")
		if err != nil {
			t.Fatalf("resolveReference failed: %v", err)
		}

		response, ok := obj.(*Response)
		if !ok {
			t.Error("Expected *Response type")
		}

		if response.Description != "Not found" {
			t.Errorf("Expected description 'Not found', got %s", response.Description)
		}
	})

	t.Run("External reference not supported", func(t *testing.T) {
		_, err := doc.resolveReference("./external.yaml#/components/schemas/Pet")
		if err == nil {
			t.Error("Expected error for external reference")
		}
	})

	t.Run("Invalid reference format", func(t *testing.T) {
		_, err := doc.resolveReference("invalid")
		if err == nil {
			t.Error("Expected error for invalid reference format")
		}
	})

	t.Run("Cache hit", func(t *testing.T) {
		// First call - should cache
		obj1, err := doc.resolveReference("#/components/schemas/Pet")
		if err != nil {
			t.Fatalf("First resolveReference failed: %v", err)
		}

		// Second call - should hit cache
		obj2, err := doc.resolveReference("#/components/schemas/Pet")
		if err != nil {
			t.Fatalf("Second resolveReference failed: %v", err)
		}

		// Should be the same object
		if obj1 != obj2 {
			t.Error("Expected cached object to be the same")
		}
	})
}
