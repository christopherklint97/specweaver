package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	p := New()
	assert.NotNil(t, p, "Expected parser to be created")
	assert.Nil(t, p.spec, "Expected spec to be nil for new parser")
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
		require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

		p := New()
		err := p.ParseFile(specPath)
		assert.NoError(t, err)
		assert.NotNil(t, p.spec, "Expected spec to be set after parsing")
		assert.Equal(t, "3.1.0", p.spec.OpenAPI)
	})

	t.Run("Parse non-existent file", func(t *testing.T) {
		p := New()
		err := p.ParseFile("/nonexistent/file.yaml")
		assert.Error(t, err, "Expected error for non-existent file")
		assert.Nil(t, p.spec, "Expected spec to remain nil after failed parse")
	})

	t.Run("Parse invalid spec", func(t *testing.T) {
		invalidSpec := `openapi: 2.0.0
info:
  title: Invalid
  version: 1.0.0
`
		specPath := filepath.Join(tmpDir, "invalid.yaml")
		require.NoError(t, os.WriteFile(specPath, []byte(invalidSpec), 0644))

		p := New()
		err := p.ParseFile(specPath)
		assert.Error(t, err, "Expected error for unsupported OpenAPI version")
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
		require.NoError(t, os.WriteFile(specPath, []byte(jsonSpec), 0644))

		p := New()
		err := p.ParseFile(specPath)
		require.NoError(t, err)
		assert.Equal(t, "JSON API", p.spec.Info.Title)
	})
}

func TestGetSpec(t *testing.T) {
	t.Run("Get spec before parsing", func(t *testing.T) {
		p := New()
		spec := p.GetSpec()
		assert.Nil(t, spec, "Expected nil spec before parsing")
	})

	t.Run("Get spec after parsing", func(t *testing.T) {
		tmpDir := t.TempDir()
		specPath := filepath.Join(tmpDir, "test.yaml")

		validSpec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths: {}`

		require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

		p := New()
		require.NoError(t, p.ParseFile(specPath))

		spec := p.GetSpec()
		require.NotNil(t, spec, "Expected spec to be available after parsing")
		assert.Equal(t, "Test API", spec.Info.Title)
	})
}

func TestGetVersion(t *testing.T) {
	t.Run("Get version before parsing", func(t *testing.T) {
		p := New()
		version := p.GetVersion()
		assert.Empty(t, version, "Expected empty version before parsing")
	})

	t.Run("Get version after parsing", func(t *testing.T) {
		tmpDir := t.TempDir()
		specPath := filepath.Join(tmpDir, "test.yaml")

		validSpec := `openapi: 3.2.0
info:
  title: Test API
  version: 1.0.0
paths: {}`

		require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

		p := New()
		require.NoError(t, p.ParseFile(specPath))

		version := p.GetVersion()
		assert.Equal(t, "3.2.0", version)
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

			require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))

			p := New()
			require.NoError(t, p.ParseFile(specPath))

			version := p.GetVersion()
			assert.Equal(t, expectedVersion, version)
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

	require.NoError(t, os.WriteFile(specPath, []byte(complexSpec), 0644))

	p := New()
	require.NoError(t, p.ParseFile(specPath))

	spec := p.GetSpec()

	// Verify document structure
	assert.Equal(t, "Complex API", spec.Info.Title)
	assert.Equal(t, "A complex API with multiple features", spec.Info.Description)
	require.NotNil(t, spec.Info.Contact)
	assert.Equal(t, "team@example.com", spec.Info.Contact.Email)
	assert.Len(t, spec.Servers, 1)
	assert.Len(t, spec.Paths, 1)

	// Verify path operation
	userPath, ok := spec.Paths["/users/{userId}"]
	require.True(t, ok, "Expected /users/{userId} path to exist")
	require.NotNil(t, userPath.Get, "Expected GET operation to exist")
	assert.Equal(t, "getUser", userPath.Get.OperationID)

	// Verify parameters
	assert.Len(t, userPath.Get.Parameters, 1)

	// Verify components
	require.NotNil(t, spec.Components)
	require.NotNil(t, spec.Components.Schemas)

	userSchema, ok := spec.Components.Schemas["User"]
	require.True(t, ok, "Expected User schema to exist")
	require.NotNil(t, userSchema.Value, "Expected User schema value to be set")
	assert.Len(t, userSchema.Value.Properties, 3)
}
