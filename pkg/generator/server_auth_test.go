package generator

import (
	"strings"
	"testing"

	"github.com/christopherklint97/specweaver/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasSecurityRequirements(t *testing.T) {
	tests := []struct {
		name     string
		spec     *openapi.Document
		op       *openapi.Operation
		expected bool
	}{
		{
			name: "operation with security",
			spec: &openapi.Document{},
			op: &openapi.Operation{
				Security: []openapi.SecurityRequirement{
					{"bearer": []string{}},
				},
			},
			expected: true,
		},
		{
			name: "operation with empty security array (public override)",
			spec: &openapi.Document{
				Security: []openapi.SecurityRequirement{
					{"bearer": []string{}},
				},
			},
			op: &openapi.Operation{
				Security: []openapi.SecurityRequirement{},
			},
			expected: false,
		},
		{
			name: "operation inherits global security",
			spec: &openapi.Document{
				Security: []openapi.SecurityRequirement{
					{"bearer": []string{}},
				},
			},
			op:       &openapi.Operation{},
			expected: true,
		},
		{
			name:     "no security requirements",
			spec:     &openapi.Document{},
			op:       &openapi.Operation{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewServerGenerator(tt.spec)
			result := gen.hasSecurityRequirements(tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSecurityRequirementsLiteral(t *testing.T) {
	tests := []struct {
		name     string
		spec     *openapi.Document
		op       *openapi.Operation
		expected []string
	}{
		{
			name: "single scheme no scopes",
			spec: &openapi.Document{},
			op: &openapi.Operation{
				Security: []openapi.SecurityRequirement{
					{"bearer": []string{}},
				},
			},
			expected: []string{
				"[]map[string][]string{",
				`"bearer": []string{}`,
			},
		},
		{
			name: "single scheme with scopes",
			spec: &openapi.Document{},
			op: &openapi.Operation{
				Security: []openapi.SecurityRequirement{
					{"oauth2": []string{"read", "write"}},
				},
			},
			expected: []string{
				"[]map[string][]string{",
				`"oauth2": []string{"read", "write"}`,
			},
		},
		{
			name: "multiple schemes (OR logic)",
			spec: &openapi.Document{},
			op: &openapi.Operation{
				Security: []openapi.SecurityRequirement{
					{"bearer": []string{}},
					{"apiKey": []string{}},
				},
			},
			expected: []string{
				"[]map[string][]string{",
				`"bearer": []string{}`,
				`"apiKey": []string{}`,
			},
		},
		{
			name: "combined schemes (AND logic)",
			spec: &openapi.Document{},
			op: &openapi.Operation{
				Security: []openapi.SecurityRequirement{
					{
						"oauth2": []string{"read"},
						"apiKey": []string{},
					},
				},
			},
			expected: []string{
				"[]map[string][]string{",
				`"apiKey": []string{}`,
				`"oauth2": []string{"read"}`,
			},
		},
		{
			name: "inherits from global",
			spec: &openapi.Document{
				Security: []openapi.SecurityRequirement{
					{"bearer": []string{}},
				},
			},
			op: &openapi.Operation{},
			expected: []string{
				"[]map[string][]string{",
				`"bearer": []string{}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewServerGenerator(tt.spec)
			result := gen.generateSecurityRequirementsLiteral(tt.op)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestGenerateSecuritySchemeInfoMap(t *testing.T) {
	spec := &openapi.Document{
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
		},
	}

	gen := NewServerGenerator(spec)
	var sb strings.Builder
	gen.generateSecuritySchemeInfoMap(&sb)
	result := sb.String()

	// Verify map declaration
	assert.Contains(t, result, "var securitySchemeInfoMap = map[string]*SecuritySchemeInfo{")

	// Verify basicAuth scheme
	assert.Contains(t, result, `"basicAuth": {`)
	assert.Contains(t, result, `Type:   "http"`)
	assert.Contains(t, result, `Scheme: "basic"`)

	// Verify apiKey scheme
	assert.Contains(t, result, `"apiKey": {`)
	assert.Contains(t, result, `Type:   "apiKey"`)
	assert.Contains(t, result, `In:     "header"`)
	assert.Contains(t, result, `Name:   "X-API-Key"`)
}

func TestGenerateRouterWithAuth(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"bearer": {
					Type:   "http",
					Scheme: "bearer",
				},
			},
		},
		Security: []openapi.SecurityRequirement{
			{"bearer": []string{}},
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

	gen := NewServerGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Verify security scheme info map is generated
	assert.Contains(t, code, "var securitySchemeInfoMap")

	// Verify ConfigureRouter accepts authenticator
	assert.Contains(t, code, "func ConfigureRouter(r router.Router, si Server, authenticator Authenticator)")

	// Verify NewRouter accepts authenticator
	assert.Contains(t, code, "func NewRouter(si Server, authenticator Authenticator)")

	// Verify protected endpoint uses auth middleware
	assert.Contains(t, code, "authMiddleware(authenticator,")

	// Verify public endpoint doesn't use auth middleware (no authMiddleware call for getPublic)
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "wrapper.handleGetPublic") {
			// Public endpoint should not have authMiddleware in the same line
			if strings.Contains(line, "authMiddleware") {
				t.Error("Public endpoint should not use auth middleware")
			}
		}
	}
}

func TestGenerateRouterWithoutAuth(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/test": {
				Get: &openapi.Operation{
					OperationID: "getTest",
					Responses: map[string]*openapi.Response{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	gen := NewServerGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Verify security scheme info map is NOT generated
	assert.NotContains(t, code, "var securitySchemeInfoMap")

	// Verify ConfigureRouter doesn't accept authenticator
	assert.Contains(t, code, "func ConfigureRouter(r router.Router, si Server)")
	assert.NotContains(t, code, "func ConfigureRouter(r router.Router, si Server, authenticator Authenticator)")

	// Verify NewRouter doesn't accept authenticator
	assert.Contains(t, code, "func NewRouter(si Server)")
	assert.NotContains(t, code, "func NewRouter(si Server, authenticator Authenticator)")

	// Verify no auth middleware is used
	assert.NotContains(t, code, "authMiddleware")
}

func TestGenerateRouterMultipleSecurityOptions(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"bearer": {Type: "http", Scheme: "bearer"},
				"apiKey": {Type: "apiKey", In: "header", Name: "X-API-Key"},
			},
		},
		Paths: map[string]*openapi.PathItem{
			"/flexible": {
				Get: &openapi.Operation{
					OperationID: "getFlexible",
					Security: []openapi.SecurityRequirement{
						{"bearer": []string{}},
						{"apiKey": []string{}},
					},
					Responses: map[string]*openapi.Response{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	gen := NewServerGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Should have OR logic for multiple security options
	assert.Contains(t, code, "[]map[string][]string{")

	// Find the router configuration section (where authMiddleware is applied)
	routerConfigStart := strings.Index(code, "func ConfigureRouter")
	assert.True(t, routerConfigStart > 0, "Should find ConfigureRouter")

	// Extract router configuration section
	routerConfigEnd := strings.Index(code[routerConfigStart:], "func NewRouter")
	assert.True(t, routerConfigEnd > 0, "Should find NewRouter")

	routerCode := code[routerConfigStart : routerConfigStart+routerConfigEnd]

	// Verify both schemes are present in the router configuration
	assert.Contains(t, routerCode, `"bearer"`)
	assert.Contains(t, routerCode, `"apiKey"`)
	assert.Contains(t, routerCode, "authMiddleware")
}

func TestGenerateRouterCombinedSecurityRequirements(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"oauth2": {Type: "oauth2"},
				"apiKey": {Type: "apiKey", In: "header", Name: "X-API-Key"},
			},
		},
		Paths: map[string]*openapi.PathItem{
			"/combined": {
				Get: &openapi.Operation{
					OperationID: "getCombined",
					Security: []openapi.SecurityRequirement{
						{
							"oauth2": []string{"read"},
							"apiKey": []string{},
						},
					},
					Responses: map[string]*openapi.Response{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	gen := NewServerGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Find the router configuration section
	routerConfigStart := strings.Index(code, "func ConfigureRouter")
	assert.True(t, routerConfigStart > 0, "Should find ConfigureRouter")

	routerConfigEnd := strings.Index(code[routerConfigStart:], "func NewRouter")
	assert.True(t, routerConfigEnd > 0, "Should find NewRouter")

	routerCode := code[routerConfigStart : routerConfigStart+routerConfigEnd]

	// Both schemes should be in the same requirement (AND logic)
	assert.Contains(t, routerCode, `"apiKey"`)
	assert.Contains(t, routerCode, `"oauth2"`)
	assert.Contains(t, routerCode, "authMiddleware")
}

func TestSecurityRequirementsLiteralDeterministic(t *testing.T) {
	spec := &openapi.Document{}
	op := &openapi.Operation{
		Security: []openapi.SecurityRequirement{
			{
				"zAuth":  []string{"scope1"},
				"aAuth":  []string{"scope2"},
				"mAuth":  []string{},
			},
		},
	}

	gen := NewServerGenerator(spec)

	// Generate twice
	result1 := gen.generateSecurityRequirementsLiteral(op)
	result2 := gen.generateSecurityRequirementsLiteral(op)

	assert.Equal(t, result1, result2, "Should produce deterministic output")

	// Verify alphabetical ordering
	aAuthPos := strings.Index(result1, `"aAuth"`)
	mAuthPos := strings.Index(result1, `"mAuth"`)
	zAuthPos := strings.Index(result1, `"zAuth"`)

	assert.True(t, aAuthPos < mAuthPos, "aAuth should come before mAuth")
	assert.True(t, mAuthPos < zAuthPos, "mAuth should come before zAuth")
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
