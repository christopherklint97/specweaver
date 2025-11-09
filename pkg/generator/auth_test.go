package generator

import (
	"strings"
	"testing"

	"github.com/christopherklint97/specweaver/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthGenerator(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	gen := NewAuthGenerator(spec)
	assert.NotNil(t, gen, "AuthGenerator should not be nil")
	assert.Equal(t, spec, gen.spec, "AuthGenerator should store spec")
}

func TestAuthGeneratorWithBasicAuth(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"basicAuth": {
					Type:   "http",
					Scheme: "basic",
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")
	assert.NotEmpty(t, code, "Generated code should not be empty")

	// Verify generated code contains expected components
	assert.Contains(t, code, "package api", "Should have package declaration")
	assert.Contains(t, code, "type BasicAuthCredentials struct", "Should have BasicAuthCredentials type")
	assert.Contains(t, code, "type SecurityContext struct", "Should have SecurityContext type")
	assert.Contains(t, code, "type Authenticator interface", "Should have Authenticator interface")
	assert.Contains(t, code, "AuthenticateBasicAuth", "Should have AuthenticateBasicAuth method")
	assert.Contains(t, code, "func authMiddleware", "Should have auth middleware")
	assert.Contains(t, code, "func extractBasicAuth", "Should have extractBasicAuth helper")
}

func TestAuthGeneratorWithBearerAuth(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	assert.Contains(t, code, "type BearerTokenCredentials struct", "Should have BearerTokenCredentials type")
	assert.Contains(t, code, "AuthenticateBearerAuth", "Should have AuthenticateBearerAuth method")
	assert.Contains(t, code, "func extractBearerToken", "Should have extractBearerToken helper")
}

func TestAuthGeneratorWithAPIKey(t *testing.T) {
	testCases := []struct {
		name     string
		location string
		keyName  string
	}{
		{"API Key in Header", "header", "X-API-Key"},
		{"API Key in Query", "query", "api_key"},
		{"API Key in Cookie", "cookie", "session_id"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec := &openapi.Document{
				OpenAPI: "3.1.0",
				Info: &openapi.Info{
					Title:   "Test API",
					Version: "1.0.0",
				},
				Components: &openapi.Components{
					SecuritySchemes: map[string]*openapi.SecurityScheme{
						"apiKey": {
							Type: "apiKey",
							In:   tc.location,
							Name: tc.keyName,
						},
					},
				},
			}

			gen := NewAuthGenerator(spec)
			code, err := gen.Generate()
			require.NoError(t, err, "Generate should not fail")

			assert.Contains(t, code, "type APIKeyCredentials struct", "Should have APIKeyCredentials type")
			assert.Contains(t, code, "AuthenticateApiKey", "Should have AuthenticateApiKey method")
			assert.Contains(t, code, "func extractAPIKey", "Should have extractAPIKey helper")
		})
	}
}

func TestAuthGeneratorWithOAuth2(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"oauth2": {
					Type: "oauth2",
					Flows: &openapi.OAuthFlows{
						AuthorizationCode: &openapi.OAuthFlow{
							AuthorizationURL: "https://example.com/oauth/authorize",
							TokenURL:         "https://example.com/oauth/token",
							Scopes: map[string]string{
								"read":  "Read access",
								"write": "Write access",
							},
						},
					},
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	assert.Contains(t, code, "type OAuth2Credentials struct", "Should have OAuth2Credentials type")
	assert.Contains(t, code, "AuthenticateOauth2", "Should have AuthenticateOauth2 method")
	assert.Contains(t, code, "func extractOAuth2Token", "Should have extractOAuth2Token helper")
}

func TestAuthGeneratorWithOpenIDConnect(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"oidc": {
					Type:             "openIdConnect",
					OpenIDConnectURL: "https://example.com/.well-known/openid-configuration",
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	assert.Contains(t, code, "type OpenIDConnectCredentials struct", "Should have OpenIDConnectCredentials type")
	assert.Contains(t, code, "AuthenticateOidc", "Should have AuthenticateOidc method")
	assert.Contains(t, code, "func extractOpenIDConnectToken", "Should have extractOpenIDConnectToken helper")
}

func TestAuthGeneratorWithMultipleSchemes(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"basicAuth": {
					Type:   "http",
					Scheme: "basic",
				},
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
				},
				"apiKey": {
					Type: "apiKey",
					In:   "header",
					Name: "X-API-Key",
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// All credential types should be present
	assert.Contains(t, code, "type BasicAuthCredentials struct")
	assert.Contains(t, code, "type BearerTokenCredentials struct")
	assert.Contains(t, code, "type APIKeyCredentials struct")

	// All authenticator methods should be present
	assert.Contains(t, code, "AuthenticateBasicAuth")
	assert.Contains(t, code, "AuthenticateBearerAuth")
	assert.Contains(t, code, "AuthenticateApiKey")

	// All extractors should be present
	assert.Contains(t, code, "func extractBasicAuth")
	assert.Contains(t, code, "func extractBearerToken")
	assert.Contains(t, code, "func extractAPIKey")
}

func TestAuthGeneratorContextHelpers(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"basic": {
					Type:   "http",
					Scheme: "basic",
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	assert.Contains(t, code, "func GetSecurityContext(ctx context.Context) *SecurityContext",
		"Should have GetSecurityContext helper")
	assert.Contains(t, code, "securityContextKey", "Should have context key")
}

func TestAuthGeneratorMiddleware(t *testing.T) {
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
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Verify middleware function exists
	assert.Contains(t, code, "func authMiddleware(authenticator Authenticator",
		"Should have authMiddleware function")

	// Verify middleware handles no security requirements
	assert.Contains(t, code, "If no security requirements, continue without authentication")

	// Verify middleware handles OR logic
	assert.Contains(t, code, "Try each security requirement (OR logic)")

	// Verify middleware handles AND logic
	assert.Contains(t, code, "All schemes in a requirement must be satisfied (AND logic)")
}

func TestAuthGeneratorDeterministicOutput(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"zAuth": {Type: "http", Scheme: "bearer"},
				"aAuth": {Type: "http", Scheme: "basic"},
				"mAuth": {Type: "apiKey", In: "header", Name: "X-API-Key"},
			},
		},
	}

	gen := NewAuthGenerator(spec)

	// Generate twice to verify deterministic output
	code1, err1 := gen.Generate()
	require.NoError(t, err1)

	code2, err2 := gen.Generate()
	require.NoError(t, err2)

	assert.Equal(t, code1, code2, "Generate should produce deterministic output")

	// Verify alphabetical ordering in authenticator interface
	// Find the position of each method in the interface
	aAuthPos := strings.Index(code1, "AuthenticateAAuth")
	mAuthPos := strings.Index(code1, "AuthenticateMAuth")
	zAuthPos := strings.Index(code1, "AuthenticateZAuth")

	assert.True(t, aAuthPos < mAuthPos, "AuthenticateAAuth should come before AuthenticateMAuth")
	assert.True(t, mAuthPos < zAuthPos, "AuthenticateMAuth should come before AuthenticateZAuth")
}

func TestAuthGeneratorImports(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"basic": {Type: "http", Scheme: "basic"},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Verify required imports
	assert.Contains(t, code, "\"context\"", "Should import context")
	assert.Contains(t, code, "\"encoding/base64\"", "Should import base64 for Basic Auth")
	assert.Contains(t, code, "\"errors\"", "Should import errors")
	assert.Contains(t, code, "\"net/http\"", "Should import net/http")
	assert.Contains(t, code, "\"strings\"", "Should import strings")
}

func TestAuthGeneratorCallAuthenticator(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"testAuth": {Type: "http", Scheme: "basic"},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Verify callAuthenticator helper exists
	assert.Contains(t, code, "func callAuthenticator(authenticator Authenticator, schemeName string",
		"Should have callAuthenticator helper")

	// Verify it handles nil authenticator
	assert.Contains(t, code, "if authenticator == nil", "Should check for nil authenticator")

	// Verify it has switch on schemeName
	assert.Contains(t, code, "switch schemeName", "Should switch on scheme name")
}

func TestAuthGeneratorSecuritySchemeInfo(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"apiKey": {
					Type: "apiKey",
					In:   "header",
					Name: "X-API-Key",
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Verify SecuritySchemeInfo type is generated
	assert.Contains(t, code, "type SecuritySchemeInfo struct",
		"Should have SecuritySchemeInfo type")
	assert.Contains(t, code, "Type   string", "SecuritySchemeInfo should have Type field")
	assert.Contains(t, code, "Scheme string", "SecuritySchemeInfo should have Scheme field")
	assert.Contains(t, code, "In     string", "SecuritySchemeInfo should have In field")
	assert.Contains(t, code, "Name   string", "SecuritySchemeInfo should have Name field")
}

func TestAuthMiddlewareSkipsWhenAuthenticatorIsNil(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"basicAuth": {
					Type:   "http",
					Scheme: "basic",
				},
			},
		},
	}

	gen := NewAuthGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Verify middleware checks for nil authenticator
	assert.Contains(t, code, "if authenticator == nil {",
		"Middleware should check for nil authenticator")
	assert.Contains(t, code, "// If no authenticator provided, skip authentication",
		"Middleware should have comment about nil authenticator")

	// Verify the nil check comes before security requirements processing
	nilCheckPos := strings.Index(code, "if authenticator == nil {")
	secReqsPos := strings.Index(code, "// Try each security requirement")
	assert.Greater(t, secReqsPos, nilCheckPos,
		"Nil authenticator check should come before security requirements processing")
}
