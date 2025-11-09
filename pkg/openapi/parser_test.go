package openapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.NoError(t, os.WriteFile(yamlPath, []byte(validYAML), 0644))

	t.Run("Load valid YAML file", func(t *testing.T) {
		doc, err := Load(yamlPath)
		require.NoError(t, err)

		assert.Equal(t, "3.1.0", doc.OpenAPI)
		assert.Equal(t, "Test API", doc.Info.Title)
		assert.Equal(t, "1.0.0", doc.Info.Version)
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		_, err := Load("/nonexistent/file.yaml")
		assert.Error(t, err)
	})

	t.Run("Load invalid YAML", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		require.NoError(t, os.WriteFile(invalidPath, []byte("invalid: yaml: content: ["), 0644))

		_, err := Load(invalidPath)
		assert.Error(t, err)
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

		require.NoError(t, os.WriteFile(jsonPath, []byte(validJSON), 0644))

		doc, err := Load(jsonPath)
		require.NoError(t, err)

		assert.Equal(t, "JSON API", doc.Info.Title)
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
		require.NoError(t, err)

		assert.Equal(t, "Data API", doc.Info.Title)
	})

	t.Run("Valid JSON data", func(t *testing.T) {
		data := []byte(`{
			"openapi": "3.0.0",
			"info": {"title": "JSON Data API", "version": "1.0.0"},
			"paths": {}
		}`)

		doc, err := LoadFromData(data, "test.json")
		require.NoError(t, err)

		assert.Equal(t, "JSON Data API", doc.Info.Title)
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
		assert.NoError(t, err)
	})

	t.Run("Nil document", func(t *testing.T) {
		err := validateDocument(nil)
		assert.Error(t, err)
	})

	t.Run("Missing OpenAPI version", func(t *testing.T) {
		doc := &Document{
			Info: &Info{
				Title:   "Test",
				Version: "1.0.0",
			},
		}

		err := validateDocument(doc)
		assert.Error(t, err)
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
		assert.Error(t, err)
	})

	t.Run("Missing info", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
		}

		err := validateDocument(doc)
		assert.Error(t, err)
	})

	t.Run("Missing info.title", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Version: "1.0.0",
			},
		}

		err := validateDocument(doc)
		assert.Error(t, err)
	})

	t.Run("Missing info.version", func(t *testing.T) {
		doc := &Document{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title: "Test",
			},
		}

		err := validateDocument(doc)
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		require.NoError(t, err)

		assert.Equal(t, "string", schema.GetSchemaType())
	})

	t.Run("Resolve reference", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/Pet",
		}

		schema, err := doc.ResolveSchemaRef(ref)
		require.NoError(t, err)

		assert.Equal(t, "object", schema.GetSchemaType())
	})

	t.Run("Resolve nil reference", func(t *testing.T) {
		_, err := doc.ResolveSchemaRef(nil)
		assert.Error(t, err)
	})

	t.Run("Resolve invalid reference", func(t *testing.T) {
		ref := &SchemaRef{
			Ref: "#/components/schemas/NonExistent",
		}

		_, err := doc.ResolveSchemaRef(ref)
		assert.Error(t, err)
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
		require.NoError(t, err)

		assert.Equal(t, "object", schema.GetSchemaType())
	})

	t.Run("Get non-existent schema", func(t *testing.T) {
		_, err := doc.GetSchemaByName("NonExistent")
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		require.NoError(t, err)

		schema, ok := obj.(*Schema)
		assert.True(t, ok, "Expected *Schema type")
		assert.Equal(t, "object", schema.GetSchemaType())
	})

	t.Run("Resolve response reference", func(t *testing.T) {
		obj, err := doc.resolveReference("#/components/responses/NotFound")
		require.NoError(t, err)

		response, ok := obj.(*Response)
		assert.True(t, ok, "Expected *Response type")
		assert.Equal(t, "Not found", response.Description)
	})

	t.Run("External reference not supported", func(t *testing.T) {
		_, err := doc.resolveReference("./external.yaml#/components/schemas/Pet")
		assert.Error(t, err)
	})

	t.Run("Invalid reference format", func(t *testing.T) {
		_, err := doc.resolveReference("invalid")
		assert.Error(t, err)
	})

	t.Run("Cache hit", func(t *testing.T) {
		// First call - should cache
		obj1, err := doc.resolveReference("#/components/schemas/Pet")
		require.NoError(t, err)

		// Second call - should hit cache
		obj2, err := doc.resolveReference("#/components/schemas/Pet")
		require.NoError(t, err)

		// Should be the same object
		assert.Same(t, obj1, obj2, "Expected cached object to be the same")
	})
}
