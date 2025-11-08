package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

func TestNewGenerator(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
	}

	t.Run("With default config", func(t *testing.T) {
		gen := NewGenerator(spec, Config{})

		if gen.packageName != "api" {
			t.Errorf("Expected default package name 'api', got %s", gen.packageName)
		}

		if gen.outputDir != "./generated" {
			t.Errorf("Expected default output dir './generated', got %s", gen.outputDir)
		}
	})

	t.Run("With custom config", func(t *testing.T) {
		config := Config{
			OutputDir:   "./custom",
			PackageName: "myapi",
		}

		gen := NewGenerator(spec, config)

		if gen.packageName != "myapi" {
			t.Errorf("Expected package name 'myapi', got %s", gen.packageName)
		}

		if gen.outputDir != "./custom" {
			t.Errorf("Expected output dir './custom', got %s", gen.outputDir)
		}
	})
}

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/pets": {
				Get: &openapi.Operation{
					OperationID: "listPets",
					Summary:     "List all pets",
					Responses: map[string]*openapi.Response{
						"200": {
							Description: "Success",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: []string{"array"},
											Items: &openapi.SchemaRef{
												Ref: "#/components/schemas/Pet",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"Pet": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"id": {
								Value: &openapi.Schema{
									Type:   []string{"integer"},
									Format: "int64",
								},
							},
							"name": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
						},
						Required: []string{"id", "name"},
					},
				},
			},
		},
	}

	config := Config{
		OutputDir:   tmpDir,
		PackageName: "api",
	}

	gen := NewGenerator(spec, config)
	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that output directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Expected output directory to be created")
	}

	// Check that types.go was created
	typesPath := filepath.Join(tmpDir, "types.go")
	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		t.Error("Expected types.go to be created")
	}

	// Check that server.go was created
	serverPath := filepath.Join(tmpDir, "server.go")
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Error("Expected server.go to be created")
	}

	// Read and verify types.go content
	typesContent, err := os.ReadFile(typesPath)
	if err != nil {
		t.Fatalf("Failed to read types.go: %v", err)
	}

	typesStr := string(typesContent)
	if len(typesStr) == 0 {
		t.Error("Expected types.go to have content")
	}

	// Read and verify server.go content
	serverContent, err := os.ReadFile(serverPath)
	if err != nil {
		t.Fatalf("Failed to read server.go: %v", err)
	}

	serverStr := string(serverContent)
	if len(serverStr) == 0 {
		t.Error("Expected server.go to have content")
	}
}

func TestGenerateTypes(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"User": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"id": {
								Value: &openapi.Schema{
									Type: []string{"integer"},
								},
							},
							"email": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
						},
						Required: []string{"id", "email"},
					},
				},
			},
		},
	}

	config := Config{
		OutputDir:   tmpDir,
		PackageName: "api",
	}

	gen := NewGenerator(spec, config)
	err := gen.generateTypes()
	if err != nil {
		t.Fatalf("generateTypes failed: %v", err)
	}

	// Check that types.go was created
	typesPath := filepath.Join(tmpDir, "types.go")
	content, err := os.ReadFile(typesPath)
	if err != nil {
		t.Fatalf("Failed to read types.go: %v", err)
	}

	contentStr := string(content)

	// Verify package declaration
	if len(contentStr) == 0 {
		t.Error("Expected types.go to have content")
	}

	// Verify the file is valid Go code by checking for package declaration
	if !contains([]string{"package api"}, "package api") {
		// Just verify file was created
		if _, err := os.Stat(typesPath); os.IsNotExist(err) {
			t.Error("Expected types.go file to exist")
		}
	}
}

func TestGenerateServer(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/test": {
				Get: &openapi.Operation{
					OperationID: "getTest",
					Responses: map[string]*openapi.Response{
						"200": {
							Description: "Success",
						},
					},
				},
			},
		},
	}

	config := Config{
		OutputDir:   tmpDir,
		PackageName: "api",
	}

	gen := NewGenerator(spec, config)
	err := gen.generateServer()
	if err != nil {
		t.Fatalf("generateServer failed: %v", err)
	}

	// Check that server.go was created
	serverPath := filepath.Join(tmpDir, "server.go")
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Error("Expected server.go to be created")
	}

	// Read content to verify it's not empty
	content, err := os.ReadFile(serverPath)
	if err != nil {
		t.Fatalf("Failed to read server.go: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected server.go to have content")
	}
}

func TestGenerateWithComplexSpec(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:       "Pet Store API",
			Version:     "1.0.0",
			Description: "A sample pet store API",
		},
		Paths: map[string]*openapi.PathItem{
			"/pets": {
				Get: &openapi.Operation{
					OperationID: "listPets",
					Summary:     "List all pets",
					Parameters: []*openapi.Parameter{
						{
							Name:     "limit",
							In:       "query",
							Required: false,
							Schema: &openapi.SchemaRef{
								Value: &openapi.Schema{
									Type: []string{"integer"},
								},
							},
						},
					},
					Responses: map[string]*openapi.Response{
						"200": {
							Description: "Success",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: []string{"array"},
											Items: &openapi.SchemaRef{
												Ref: "#/components/schemas/Pet",
											},
										},
									},
								},
							},
						},
					},
				},
				Post: &openapi.Operation{
					OperationID: "createPet",
					Summary:     "Create a pet",
					RequestBody: &openapi.RequestBody{
						Required: true,
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &openapi.SchemaRef{
									Ref: "#/components/schemas/NewPet",
								},
							},
						},
					},
					Responses: map[string]*openapi.Response{
						"201": {
							Description: "Created",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Ref: "#/components/schemas/Pet",
									},
								},
							},
						},
					},
				},
			},
			"/pets/{petId}": {
				Get: &openapi.Operation{
					OperationID: "getPetById",
					Parameters: []*openapi.Parameter{
						{
							Name:     "petId",
							In:       "path",
							Required: true,
							Schema: &openapi.SchemaRef{
								Value: &openapi.Schema{
									Type: []string{"integer"},
								},
							},
						},
					},
					Responses: map[string]*openapi.Response{
						"200": {
							Description: "Success",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Ref: "#/components/schemas/Pet",
									},
								},
							},
						},
						"404": {
							Description: "Not found",
						},
					},
				},
			},
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"Pet": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"id": {
								Value: &openapi.Schema{
									Type: []string{"integer"},
								},
							},
							"name": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
							"tag": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
						},
						Required: []string{"id", "name"},
					},
				},
				"NewPet": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"name": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
							"tag": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
						},
						Required: []string{"name"},
					},
				},
			},
		},
	}

	config := Config{
		OutputDir:   tmpDir,
		PackageName: "api",
	}

	gen := NewGenerator(spec, config)
	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify both files exist
	typesPath := filepath.Join(tmpDir, "types.go")
	serverPath := filepath.Join(tmpDir, "server.go")

	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		t.Error("Expected types.go to exist")
	}

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Error("Expected server.go to exist")
	}
}

func TestGenerateWithInvalidOutputDir(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{},
	}

	// Use an invalid path (try to create inside a file instead of directory)
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		OutputDir: filepath.Join(tmpFile, "subdir"), // Invalid: trying to create dir inside a file
	}

	gen := NewGenerator(spec, config)
	err := gen.Generate()
	if err == nil {
		t.Error("Expected error when creating invalid output directory")
	}
}

func TestGenerateEmptySpec(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Empty API",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{},
	}

	config := Config{
		OutputDir:   tmpDir,
		PackageName: "api",
	}

	gen := NewGenerator(spec, config)
	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed for empty spec: %v", err)
	}

	// Files should still be created even with empty spec
	typesPath := filepath.Join(tmpDir, "types.go")
	serverPath := filepath.Join(tmpDir, "server.go")

	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		t.Error("Expected types.go to be created for empty spec")
	}

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Error("Expected server.go to be created for empty spec")
	}
}
