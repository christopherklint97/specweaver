package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("Expected parser to be created")
	}

	if p.spec != nil {
		t.Error("Expected spec to be nil for new parser")
	}
}

func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()

	validSpec := `openapi: 3.1.0
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

	t.Run("Parse valid file", func(t *testing.T) {
		specPath := filepath.Join(tmpDir, "valid.yaml")
		if err := os.WriteFile(specPath, []byte(validSpec), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		p := New()
		err := p.ParseFile(specPath)
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		if p.spec == nil {
			t.Error("Expected spec to be set after parsing")
		}

		if p.spec.OpenAPI != "3.1.0" {
			t.Errorf("Expected OpenAPI version 3.1.0, got %s", p.spec.OpenAPI)
		}
	})

	t.Run("Parse non-existent file", func(t *testing.T) {
		p := New()
		err := p.ParseFile("/nonexistent/file.yaml")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		if p.spec != nil {
			t.Error("Expected spec to remain nil after failed parse")
		}
	})

	t.Run("Parse invalid spec", func(t *testing.T) {
		invalidSpec := `openapi: 2.0.0
info:
  title: Invalid
  version: 1.0.0
`
		specPath := filepath.Join(tmpDir, "invalid.yaml")
		if err := os.WriteFile(specPath, []byte(invalidSpec), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		p := New()
		err := p.ParseFile(specPath)
		if err == nil {
			t.Error("Expected error for unsupported OpenAPI version")
		}
	})

	t.Run("Parse JSON file", func(t *testing.T) {
		jsonSpec := `{
			"openapi": "3.0.0",
			"info": {
				"title": "JSON API",
				"version": "1.0.0"
			},
			"paths": {}
		}`

		specPath := filepath.Join(tmpDir, "spec.json")
		if err := os.WriteFile(specPath, []byte(jsonSpec), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		p := New()
		err := p.ParseFile(specPath)
		if err != nil {
			t.Fatalf("ParseFile failed for JSON: %v", err)
		}

		if p.spec.Info.Title != "JSON API" {
			t.Errorf("Expected title 'JSON API', got %s", p.spec.Info.Title)
		}
	})
}

func TestGetSpec(t *testing.T) {
	t.Run("Get spec before parsing", func(t *testing.T) {
		p := New()
		spec := p.GetSpec()
		if spec != nil {
			t.Error("Expected nil spec before parsing")
		}
	})

	t.Run("Get spec after parsing", func(t *testing.T) {
		tmpDir := t.TempDir()
		specPath := filepath.Join(tmpDir, "test.yaml")

		validSpec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths: {}`

		if err := os.WriteFile(specPath, []byte(validSpec), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		p := New()
		if err := p.ParseFile(specPath); err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		spec := p.GetSpec()
		if spec == nil {
			t.Error("Expected spec to be available after parsing")
		}

		if spec.Info.Title != "Test API" {
			t.Errorf("Expected title 'Test API', got %s", spec.Info.Title)
		}
	})
}

func TestGetVersion(t *testing.T) {
	t.Run("Get version before parsing", func(t *testing.T) {
		p := New()
		version := p.GetVersion()
		if version != "" {
			t.Errorf("Expected empty version before parsing, got %s", version)
		}
	})

	t.Run("Get version after parsing", func(t *testing.T) {
		tmpDir := t.TempDir()
		specPath := filepath.Join(tmpDir, "test.yaml")

		validSpec := `openapi: 3.2.0
info:
  title: Test API
  version: 1.0.0
paths: {}`

		if err := os.WriteFile(specPath, []byte(validSpec), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		p := New()
		if err := p.ParseFile(specPath); err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		version := p.GetVersion()
		if version != "3.2.0" {
			t.Errorf("Expected version '3.2.0', got %s", version)
		}
	})

	t.Run("Get version for different OpenAPI versions", func(t *testing.T) {
		versions := []string{"3.0.0", "3.0.3", "3.1.0", "3.1.1", "3.2.0"}

		for _, expectedVersion := range versions {
			tmpDir := t.TempDir()
			specPath := filepath.Join(tmpDir, "test.yaml")

			spec := `openapi: ` + expectedVersion + `
info:
  title: Test API
  version: 1.0.0
paths: {}`

			if err := os.WriteFile(specPath, []byte(spec), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			p := New()
			if err := p.ParseFile(specPath); err != nil {
				t.Fatalf("ParseFile failed for version %s: %v", expectedVersion, err)
			}

			version := p.GetVersion()
			if version != expectedVersion {
				t.Errorf("Expected version '%s', got %s", expectedVersion, version)
			}
		}
	})
}

func TestParserWithComplexSpec(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "complex.yaml")

	complexSpec := `openapi: 3.1.0
info:
  title: Complex API
  version: 2.0.0
  description: A complex API with multiple features
  contact:
    name: API Team
    email: team@example.com
servers:
  - url: https://api.example.com
    description: Production server
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: User found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        email:
          type: string
          format: email
`

	if err := os.WriteFile(specPath, []byte(complexSpec), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := New()
	if err := p.ParseFile(specPath); err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	spec := p.GetSpec()

	// Verify document structure
	if spec.Info.Title != "Complex API" {
		t.Errorf("Expected title 'Complex API', got %s", spec.Info.Title)
	}

	if spec.Info.Description != "A complex API with multiple features" {
		t.Errorf("Unexpected description: %s", spec.Info.Description)
	}

	if spec.Info.Contact == nil || spec.Info.Contact.Email != "team@example.com" {
		t.Error("Expected contact email to be set")
	}

	if len(spec.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(spec.Servers))
	}

	if len(spec.Paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(spec.Paths))
	}

	// Verify path operation
	userPath, ok := spec.Paths["/users/{userId}"]
	if !ok {
		t.Fatal("Expected /users/{userId} path to exist")
	}

	if userPath.Get == nil {
		t.Fatal("Expected GET operation to exist")
	}

	if userPath.Get.OperationID != "getUser" {
		t.Errorf("Expected operation ID 'getUser', got %s", userPath.Get.OperationID)
	}

	// Verify parameters
	if len(userPath.Get.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(userPath.Get.Parameters))
	}

	// Verify components
	if spec.Components == nil || spec.Components.Schemas == nil {
		t.Fatal("Expected components.schemas to be set")
	}

	userSchema, ok := spec.Components.Schemas["User"]
	if !ok {
		t.Fatal("Expected User schema to exist")
	}

	if userSchema.Value == nil {
		t.Fatal("Expected User schema value to be set")
	}

	if len(userSchema.Value.Properties) != 3 {
		t.Errorf("Expected 3 properties in User schema, got %d", len(userSchema.Value.Properties))
	}
}
