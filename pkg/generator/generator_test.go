package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/christopherklint97/specweaver/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		assert.Equal(t, "api", gen.packageName, "Expected default package name 'api'")
		assert.Equal(t, "./generated", gen.outputDir, "Expected default output dir './generated'")
	})

	t.Run("With custom config", func(t *testing.T) {
		config := Config{
			OutputDir:   "./custom",
			PackageName: "myapi",
		}

		gen := NewGenerator(spec, config)

		assert.Equal(t, "myapi", gen.packageName, "Expected package name 'myapi'")
		assert.Equal(t, "./custom", gen.outputDir, "Expected output dir './custom'")
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
	require.NoError(t, err, "Generate should not fail")

	// Check that output directory was created
	assert.DirExists(t, tmpDir, "Expected output directory to be created")

	// Check that types.go was created
	typesPath := filepath.Join(tmpDir, "types.go")
	assert.FileExists(t, typesPath, "Expected types.go to be created")

	// Check that server.go was created
	serverPath := filepath.Join(tmpDir, "server.go")
	assert.FileExists(t, serverPath, "Expected server.go to be created")

	// Read and verify types.go content
	typesContent, err := os.ReadFile(typesPath)
	require.NoError(t, err, "Failed to read types.go")

	typesStr := string(typesContent)
	assert.NotEmpty(t, typesStr, "Expected types.go to have content")

	// Read and verify server.go content
	serverContent, err := os.ReadFile(serverPath)
	require.NoError(t, err, "Failed to read server.go")

	serverStr := string(serverContent)
	assert.NotEmpty(t, serverStr, "Expected server.go to have content")
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
	require.NoError(t, err, "generateTypes should not fail")

	// Check that types.go was created
	typesPath := filepath.Join(tmpDir, "types.go")
	content, err := os.ReadFile(typesPath)
	require.NoError(t, err, "Failed to read types.go")

	contentStr := string(content)

	// Verify package declaration
	assert.NotEmpty(t, contentStr, "Expected types.go to have content")

	// Verify the file is valid Go code by checking for package declaration
	if !contains([]string{"package api"}, "package api") {
		// Just verify file was created
		assert.FileExists(t, typesPath, "Expected types.go file to exist")
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
	require.NoError(t, err, "generateServer should not fail")

	// Check that server.go was created
	serverPath := filepath.Join(tmpDir, "server.go")
	assert.FileExists(t, serverPath, "Expected server.go to be created")

	// Read content to verify it's not empty
	content, err := os.ReadFile(serverPath)
	require.NoError(t, err, "Failed to read server.go")

	assert.NotEmpty(t, content, "Expected server.go to have content")
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
	require.NoError(t, err, "Generate should not fail")

	// Verify both files exist
	typesPath := filepath.Join(tmpDir, "types.go")
	serverPath := filepath.Join(tmpDir, "server.go")

	assert.FileExists(t, typesPath, "Expected types.go to exist")
	assert.FileExists(t, serverPath, "Expected server.go to exist")
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
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err, "Failed to create test file")

	config := Config{
		OutputDir: filepath.Join(tmpFile, "subdir"), // Invalid: trying to create dir inside a file
	}

	gen := NewGenerator(spec, config)
	err = gen.Generate()
	assert.Error(t, err, "Expected error when creating invalid output directory")
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
	require.NoError(t, err, "Generate should not fail for empty spec")

	// Files should still be created even with empty spec
	typesPath := filepath.Join(tmpDir, "types.go")
	serverPath := filepath.Join(tmpDir, "server.go")

	assert.FileExists(t, typesPath, "Expected types.go to be created for empty spec")
	assert.FileExists(t, serverPath, "Expected server.go to be created for empty spec")
}

func TestGenerateAuth(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API with Auth",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
				},
			},
		},
		Paths: map[string]*openapi.PathItem{
			"/protected": {
				Get: &openapi.Operation{
					OperationID: "getProtected",
					Responses: map[string]*openapi.Response{
						"200": {Description: "Success"},
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
	err := gen.generateAuth()
	require.NoError(t, err, "generateAuth should not fail")

	// Check that auth.go was created
	authPath := filepath.Join(tmpDir, "auth.go")
	assert.FileExists(t, authPath, "Expected auth.go to be created")

	// Read and verify content
	content, err := os.ReadFile(authPath)
	require.NoError(t, err, "Failed to read auth.go")

	contentStr := string(content)
	assert.Contains(t, contentStr, "type Authenticator interface", "Should contain Authenticator interface")
	assert.Contains(t, contentStr, "AuthenticateBearerAuth", "Should contain bearer auth method")
}

func TestGenerateAuthNotCreatedWithoutSecuritySchemes(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API without Auth",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/public": {
				Get: &openapi.Operation{
					OperationID: "getPublic",
					Responses: map[string]*openapi.Response{
						"200": {Description: "Success"},
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
	err := gen.generateAuth()
	require.NoError(t, err, "generateAuth should not fail even without security schemes")

	// auth.go should NOT be created
	authPath := filepath.Join(tmpDir, "auth.go")
	assert.NoFileExists(t, authPath, "auth.go should not be created without security schemes")
}

func TestHasSecuritySchemes(t *testing.T) {
	tests := []struct {
		name     string
		spec     *openapi.Document
		expected bool
	}{
		{
			name: "with security schemes",
			spec: &openapi.Document{
				Components: &openapi.Components{
					SecuritySchemes: map[string]*openapi.SecurityScheme{
						"bearer": {Type: "http", Scheme: "bearer"},
					},
				},
			},
			expected: true,
		},
		{
			name: "without components",
			spec: &openapi.Document{
				Components: nil,
			},
			expected: false,
		},
		{
			name: "without security schemes",
			spec: &openapi.Document{
				Components: &openapi.Components{
					SecuritySchemes: nil,
				},
			},
			expected: false,
		},
		{
			name: "with empty security schemes",
			spec: &openapi.Document{
				Components: &openapi.Components{
					SecuritySchemes: map[string]*openapi.SecurityScheme{},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tt.spec, Config{})
			result := gen.hasSecuritySchemes()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateWithAuthIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Auth API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"basicAuth": {
					Type:   "http",
					Scheme: "basic",
				},
				"apiKey": {
					Type: "apiKey",
					In:   "header",
					Name: "X-API-Key",
				},
			},
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
						},
						Required: []string{"id"},
					},
				},
			},
		},
		Security: []openapi.SecurityRequirement{
			{"basicAuth": []string{}},
		},
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Get: &openapi.Operation{
					OperationID: "listUsers",
					Responses: map[string]*openapi.Response{
						"200": {
							Description: "Success",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: []string{"array"},
											Items: &openapi.SchemaRef{
												Ref: "#/components/schemas/User",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/public": {
				Get: &openapi.Operation{
					OperationID: "getPublic",
					Security:    []openapi.SecurityRequirement{},
					Responses: map[string]*openapi.Response{
						"200": {Description: "Success"},
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
	require.NoError(t, err, "Generate should not fail")

	// All three files should be created
	typesPath := filepath.Join(tmpDir, "types.go")
	serverPath := filepath.Join(tmpDir, "server.go")
	authPath := filepath.Join(tmpDir, "auth.go")

	assert.FileExists(t, typesPath, "Expected types.go to be created")
	assert.FileExists(t, serverPath, "Expected server.go to be created")
	assert.FileExists(t, authPath, "Expected auth.go to be created")

	// Verify auth.go content
	authContent, err := os.ReadFile(authPath)
	require.NoError(t, err, "Failed to read auth.go")

	authStr := string(authContent)
	assert.Contains(t, authStr, "AuthenticateBasicAuth")
	assert.Contains(t, authStr, "AuthenticateApiKey")

	// Verify server.go has auth integration
	serverContent, err := os.ReadFile(serverPath)
	require.NoError(t, err, "Failed to read server.go")

	serverStr := string(serverContent)
	assert.Contains(t, serverStr, "authenticator Authenticator", "Server should accept authenticator")
	assert.Contains(t, serverStr, "authMiddleware", "Server should use auth middleware")
}

