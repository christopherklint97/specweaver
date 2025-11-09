package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// AuthGenerator generates authentication code from OpenAPI security schemes
type AuthGenerator struct {
	spec *openapi.Document
}

// NewAuthGenerator creates a new AuthGenerator instance
func NewAuthGenerator(spec *openapi.Document) *AuthGenerator {
	return &AuthGenerator{
		spec: spec,
	}
}

// Generate generates authentication code
func (g *AuthGenerator) Generate() (string, error) {
	var sb strings.Builder

	sb.WriteString("package api\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"encoding/base64\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString("\t\"strings\"\n")
	sb.WriteString(")\n\n")

	// Generate context key
	g.generateContextKey(&sb)

	// Generate credential types
	g.generateCredentialTypes(&sb)

	// Generate authenticator interface
	g.generateAuthenticatorInterface(&sb)

	// Generate authentication middleware
	g.generateAuthMiddleware(&sb)

	// Generate credential extraction helpers
	g.generateCredentialExtractors(&sb)

	return sb.String(), nil
}

// generateContextKey generates the context key for storing auth info
func (g *AuthGenerator) generateContextKey(sb *strings.Builder) {
	sb.WriteString("// contextKey is a private type for context keys to avoid collisions\n")
	sb.WriteString("type contextKey string\n\n")
	sb.WriteString("// securityContextKey is the context key for security information\n")
	sb.WriteString("const securityContextKey contextKey = \"security\"\n\n")
}

// generateCredentialTypes generates types for different credential types
func (g *AuthGenerator) generateCredentialTypes(sb *strings.Builder) {
	sb.WriteString("// Credential types for different authentication schemes\n\n")

	// BasicAuthCredentials
	sb.WriteString("// BasicAuthCredentials holds HTTP Basic authentication credentials\n")
	sb.WriteString("type BasicAuthCredentials struct {\n")
	sb.WriteString("\tUsername string\n")
	sb.WriteString("\tPassword string\n")
	sb.WriteString("}\n\n")

	// BearerTokenCredentials
	sb.WriteString("// BearerTokenCredentials holds HTTP Bearer token credentials\n")
	sb.WriteString("type BearerTokenCredentials struct {\n")
	sb.WriteString("\tToken string\n")
	sb.WriteString("}\n\n")

	// APIKeyCredentials
	sb.WriteString("// APIKeyCredentials holds API key credentials\n")
	sb.WriteString("type APIKeyCredentials struct {\n")
	sb.WriteString("\tKey      string\n")
	sb.WriteString("\tLocation string // \"header\", \"query\", or \"cookie\"\n")
	sb.WriteString("\tName     string // The name of the header, query param, or cookie\n")
	sb.WriteString("}\n\n")

	// OAuth2Credentials
	sb.WriteString("// OAuth2Credentials holds OAuth 2.0 credentials\n")
	sb.WriteString("type OAuth2Credentials struct {\n")
	sb.WriteString("\tToken  string\n")
	sb.WriteString("\tScopes []string\n")
	sb.WriteString("}\n\n")

	// OpenIDConnectCredentials
	sb.WriteString("// OpenIDConnectCredentials holds OpenID Connect credentials\n")
	sb.WriteString("type OpenIDConnectCredentials struct {\n")
	sb.WriteString("\tToken string\n")
	sb.WriteString("}\n\n")

	// SecurityContext
	sb.WriteString("// SecurityContext holds authentication information\n")
	sb.WriteString("type SecurityContext struct {\n")
	sb.WriteString("\t// Principal is the authenticated user/entity (type depends on your implementation)\n")
	sb.WriteString("\tPrincipal any\n")
	sb.WriteString("\t// SchemeName is the name of the security scheme that was used\n")
	sb.WriteString("\tSchemeName string\n")
	sb.WriteString("\t// Scopes are the OAuth2 scopes (if applicable)\n")
	sb.WriteString("\tScopes []string\n")
	sb.WriteString("}\n\n")

	// Helper to get security context from request context
	sb.WriteString("// GetSecurityContext retrieves the security context from the request context\n")
	sb.WriteString("// Returns nil if no authentication was performed\n")
	sb.WriteString("func GetSecurityContext(ctx context.Context) *SecurityContext {\n")
	sb.WriteString("\tif sc, ok := ctx.Value(securityContextKey).(*SecurityContext); ok {\n")
	sb.WriteString("\t\treturn sc\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
}

// generateAuthenticatorInterface generates the Authenticator interface
func (g *AuthGenerator) generateAuthenticatorInterface(sb *strings.Builder) {
	sb.WriteString("// Authenticator defines the interface for authentication handlers\n")
	sb.WriteString("// Implement this interface to provide authentication logic for your API\n")
	sb.WriteString("type Authenticator interface {\n")

	if g.spec.Components != nil && g.spec.Components.SecuritySchemes != nil {
		// Get security scheme names in sorted order
		schemes := make([]string, 0, len(g.spec.Components.SecuritySchemes))
		for name := range g.spec.Components.SecuritySchemes {
			schemes = append(schemes, name)
		}
		sort.Strings(schemes)

		// Generate method for each security scheme
		for _, name := range schemes {
			scheme := g.spec.Components.SecuritySchemes[name]
			if scheme == nil {
				continue
			}

			methodName := toPascalCase(name)

			switch scheme.Type {
			case "http":
				if scheme.Scheme == "basic" {
					sb.WriteString(fmt.Sprintf("\t// Authenticate%s authenticates using HTTP Basic Auth\n", methodName))
					sb.WriteString("\t// Returns the authenticated principal or an error\n")
					sb.WriteString(fmt.Sprintf("\tAuthenticate%s(ctx context.Context, credentials BasicAuthCredentials) (any, error)\n\n", methodName))
				} else if scheme.Scheme == "bearer" {
					sb.WriteString(fmt.Sprintf("\t// Authenticate%s authenticates using HTTP Bearer token\n", methodName))
					sb.WriteString("\t// Returns the authenticated principal or an error\n")
					sb.WriteString(fmt.Sprintf("\tAuthenticate%s(ctx context.Context, credentials BearerTokenCredentials) (any, error)\n\n", methodName))
				}
			case "apiKey":
				sb.WriteString(fmt.Sprintf("\t// Authenticate%s authenticates using API Key\n", methodName))
				sb.WriteString("\t// Returns the authenticated principal or an error\n")
				sb.WriteString(fmt.Sprintf("\tAuthenticate%s(ctx context.Context, credentials APIKeyCredentials) (any, error)\n\n", methodName))
			case "oauth2":
				sb.WriteString(fmt.Sprintf("\t// Authenticate%s authenticates using OAuth 2.0\n", methodName))
				sb.WriteString("\t// Returns the authenticated principal or an error\n")
				sb.WriteString(fmt.Sprintf("\tAuthenticate%s(ctx context.Context, credentials OAuth2Credentials) (any, error)\n\n", methodName))
			case "openIdConnect":
				sb.WriteString(fmt.Sprintf("\t// Authenticate%s authenticates using OpenID Connect\n", methodName))
				sb.WriteString("\t// Returns the authenticated principal or an error\n")
				sb.WriteString(fmt.Sprintf("\tAuthenticate%s(ctx context.Context, credentials OpenIDConnectCredentials) (any, error)\n\n", methodName))
			}
		}
	}

	sb.WriteString("}\n\n")
}

// generateAuthMiddleware generates the authentication middleware
func (g *AuthGenerator) generateAuthMiddleware(sb *strings.Builder) {
	sb.WriteString("// authMiddleware creates authentication middleware for an operation\n")
	sb.WriteString("func authMiddleware(authenticator Authenticator, securityReqs []map[string][]string, schemes map[string]*SecuritySchemeInfo) func(http.Handler) http.Handler {\n")
	sb.WriteString("\treturn func(next http.Handler) http.Handler {\n")
	sb.WriteString("\t\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n")
	sb.WriteString("\t\t\tctx := r.Context()\n\n")

	sb.WriteString("\t\t\t// If no security requirements, continue without authentication\n")
	sb.WriteString("\t\t\tif len(securityReqs) == 0 {\n")
	sb.WriteString("\t\t\t\tnext.ServeHTTP(w, r)\n")
	sb.WriteString("\t\t\t\treturn\n")
	sb.WriteString("\t\t\t}\n\n")

	sb.WriteString("\t\t\t// Try each security requirement (OR logic)\n")
	sb.WriteString("\t\t\tfor _, req := range securityReqs {\n")
	sb.WriteString("\t\t\t\t// All schemes in a requirement must be satisfied (AND logic)\n")
	sb.WriteString("\t\t\t\tvar secCtx *SecurityContext\n")
	sb.WriteString("\t\t\t\tvar authErr error\n")
	sb.WriteString("\t\t\t\tallSatisfied := true\n\n")

	sb.WriteString("\t\t\t\tfor schemeName, scopes := range req {\n")
	sb.WriteString("\t\t\t\t\tschemeInfo, exists := schemes[schemeName]\n")
	sb.WriteString("\t\t\t\t\tif !exists {\n")
	sb.WriteString("\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\tbreak\n")
	sb.WriteString("\t\t\t\t\t}\n\n")

	sb.WriteString("\t\t\t\t\t// Authenticate based on scheme type\n")
	sb.WriteString("\t\t\t\t\tvar principal any\n")
	sb.WriteString("\t\t\t\t\tswitch schemeInfo.Type {\n")
	sb.WriteString("\t\t\t\t\tcase \"http\":\n")
	sb.WriteString("\t\t\t\t\t\tif schemeInfo.Scheme == \"basic\" {\n")
	sb.WriteString("\t\t\t\t\t\t\tcreds, err := extractBasicAuth(r)\n")
	sb.WriteString("\t\t\t\t\t\t\tif err != nil {\n")
	sb.WriteString("\t\t\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\t\t\tauthErr = err\n")
	sb.WriteString("\t\t\t\t\t\t\t\tbreak\n")
	sb.WriteString("\t\t\t\t\t\t\t}\n")
	sb.WriteString("\t\t\t\t\t\t\tprincipal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)\n")
	sb.WriteString("\t\t\t\t\t\t} else if schemeInfo.Scheme == \"bearer\" {\n")
	sb.WriteString("\t\t\t\t\t\t\tcreds, err := extractBearerToken(r)\n")
	sb.WriteString("\t\t\t\t\t\t\tif err != nil {\n")
	sb.WriteString("\t\t\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\t\t\tauthErr = err\n")
	sb.WriteString("\t\t\t\t\t\t\t\tbreak\n")
	sb.WriteString("\t\t\t\t\t\t\t}\n")
	sb.WriteString("\t\t\t\t\t\t\tprincipal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)\n")
	sb.WriteString("\t\t\t\t\t\t}\n")
	sb.WriteString("\t\t\t\t\tcase \"apiKey\":\n")
	sb.WriteString("\t\t\t\t\t\tcreds, err := extractAPIKey(r, schemeInfo.In, schemeInfo.Name)\n")
	sb.WriteString("\t\t\t\t\t\tif err != nil {\n")
	sb.WriteString("\t\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\t\tauthErr = err\n")
	sb.WriteString("\t\t\t\t\t\t\tbreak\n")
	sb.WriteString("\t\t\t\t\t\t}\n")
	sb.WriteString("\t\t\t\t\t\tprincipal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)\n")
	sb.WriteString("\t\t\t\t\tcase \"oauth2\":\n")
	sb.WriteString("\t\t\t\t\t\tcreds, err := extractOAuth2Token(r, scopes)\n")
	sb.WriteString("\t\t\t\t\t\tif err != nil {\n")
	sb.WriteString("\t\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\t\tauthErr = err\n")
	sb.WriteString("\t\t\t\t\t\t\tbreak\n")
	sb.WriteString("\t\t\t\t\t\t}\n")
	sb.WriteString("\t\t\t\t\t\tprincipal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)\n")
	sb.WriteString("\t\t\t\t\tcase \"openIdConnect\":\n")
	sb.WriteString("\t\t\t\t\t\tcreds, err := extractOpenIDConnectToken(r)\n")
	sb.WriteString("\t\t\t\t\t\tif err != nil {\n")
	sb.WriteString("\t\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\t\tauthErr = err\n")
	sb.WriteString("\t\t\t\t\t\t\tbreak\n")
	sb.WriteString("\t\t\t\t\t\t}\n")
	sb.WriteString("\t\t\t\t\t\tprincipal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)\n")
	sb.WriteString("\t\t\t\t\tdefault:\n")
	sb.WriteString("\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\tauthErr = errors.New(\"unsupported security scheme type\")\n")
	sb.WriteString("\t\t\t\t\t}\n\n")

	sb.WriteString("\t\t\t\t\tif authErr != nil {\n")
	sb.WriteString("\t\t\t\t\t\tallSatisfied = false\n")
	sb.WriteString("\t\t\t\t\t\tbreak\n")
	sb.WriteString("\t\t\t\t\t}\n\n")

	sb.WriteString("\t\t\t\t\t// Create or update security context\n")
	sb.WriteString("\t\t\t\t\tsecCtx = &SecurityContext{\n")
	sb.WriteString("\t\t\t\t\t\tPrincipal:  principal,\n")
	sb.WriteString("\t\t\t\t\t\tSchemeName: schemeName,\n")
	sb.WriteString("\t\t\t\t\t\tScopes:     scopes,\n")
	sb.WriteString("\t\t\t\t\t}\n")
	sb.WriteString("\t\t\t\t}\n\n")

	sb.WriteString("\t\t\t\t// If all schemes in this requirement were satisfied, continue\n")
	sb.WriteString("\t\t\t\tif allSatisfied && secCtx != nil {\n")
	sb.WriteString("\t\t\t\t\tctx = context.WithValue(ctx, securityContextKey, secCtx)\n")
	sb.WriteString("\t\t\t\t\tr = r.WithContext(ctx)\n")
	sb.WriteString("\t\t\t\t\tnext.ServeHTTP(w, r)\n")
	sb.WriteString("\t\t\t\t\treturn\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t}\n\n")

	sb.WriteString("\t\t\t// None of the security requirements were satisfied\n")
	sb.WriteString("\t\t\tWriteError(w, http.StatusUnauthorized, errors.New(\"authentication required\"))\n")
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Helper to call the right authenticator method
	sb.WriteString("// callAuthenticator calls the appropriate authenticator method based on scheme name\n")
	sb.WriteString("func callAuthenticator(authenticator Authenticator, schemeName string, ctx context.Context, credentials any) (any, error) {\n")
	sb.WriteString("\tif authenticator == nil {\n")
	sb.WriteString("\t\treturn nil, errors.New(\"no authenticator provided\")\n")
	sb.WriteString("\t}\n\n")

	if g.spec.Components != nil && g.spec.Components.SecuritySchemes != nil {
		sb.WriteString("\tswitch schemeName {\n")

		// Get security scheme names in sorted order
		schemes := make([]string, 0, len(g.spec.Components.SecuritySchemes))
		for name := range g.spec.Components.SecuritySchemes {
			schemes = append(schemes, name)
		}
		sort.Strings(schemes)

		for _, name := range schemes {
			scheme := g.spec.Components.SecuritySchemes[name]
			if scheme == nil {
				continue
			}

			methodName := toPascalCase(name)
			sb.WriteString(fmt.Sprintf("\tcase \"%s\":\n", name))

			switch scheme.Type {
			case "http":
				if scheme.Scheme == "basic" {
					sb.WriteString("\t\tif creds, ok := credentials.(BasicAuthCredentials); ok {\n")
					sb.WriteString(fmt.Sprintf("\t\t\treturn authenticator.Authenticate%s(ctx, creds)\n", methodName))
					sb.WriteString("\t\t}\n")
				} else if scheme.Scheme == "bearer" {
					sb.WriteString("\t\tif creds, ok := credentials.(BearerTokenCredentials); ok {\n")
					sb.WriteString(fmt.Sprintf("\t\t\treturn authenticator.Authenticate%s(ctx, creds)\n", methodName))
					sb.WriteString("\t\t}\n")
				}
			case "apiKey":
				sb.WriteString("\t\tif creds, ok := credentials.(APIKeyCredentials); ok {\n")
				sb.WriteString(fmt.Sprintf("\t\t\treturn authenticator.Authenticate%s(ctx, creds)\n", methodName))
				sb.WriteString("\t\t}\n")
			case "oauth2":
				sb.WriteString("\t\tif creds, ok := credentials.(OAuth2Credentials); ok {\n")
				sb.WriteString(fmt.Sprintf("\t\t\treturn authenticator.Authenticate%s(ctx, creds)\n", methodName))
				sb.WriteString("\t\t}\n")
			case "openIdConnect":
				sb.WriteString("\t\tif creds, ok := credentials.(OpenIDConnectCredentials); ok {\n")
				sb.WriteString(fmt.Sprintf("\t\t\treturn authenticator.Authenticate%s(ctx, creds)\n", methodName))
				sb.WriteString("\t\t}\n")
			}
		}

		sb.WriteString("\t}\n\n")
	}

	sb.WriteString("\treturn nil, errors.New(\"unknown security scheme\")\n")
	sb.WriteString("}\n\n")

	// SecuritySchemeInfo type
	sb.WriteString("// SecuritySchemeInfo holds information about a security scheme\n")
	sb.WriteString("type SecuritySchemeInfo struct {\n")
	sb.WriteString("\tType   string\n")
	sb.WriteString("\tScheme string\n")
	sb.WriteString("\tIn     string\n")
	sb.WriteString("\tName   string\n")
	sb.WriteString("}\n\n")
}

// generateCredentialExtractors generates helper functions to extract credentials
func (g *AuthGenerator) generateCredentialExtractors(sb *strings.Builder) {
	sb.WriteString("// Credential extraction helpers\n\n")

	// extractBasicAuth
	sb.WriteString("// extractBasicAuth extracts HTTP Basic Auth credentials from request\n")
	sb.WriteString("func extractBasicAuth(r *http.Request) (BasicAuthCredentials, error) {\n")
	sb.WriteString("\tauth := r.Header.Get(\"Authorization\")\n")
	sb.WriteString("\tif auth == \"\" {\n")
	sb.WriteString("\t\treturn BasicAuthCredentials{}, errors.New(\"missing Authorization header\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tconst prefix = \"Basic \"\n")
	sb.WriteString("\tif !strings.HasPrefix(auth, prefix) {\n")
	sb.WriteString("\t\treturn BasicAuthCredentials{}, errors.New(\"invalid Authorization header format\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tdecoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn BasicAuthCredentials{}, errors.New(\"invalid base64 encoding\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tparts := strings.SplitN(string(decoded), \":\", 2)\n")
	sb.WriteString("\tif len(parts) != 2 {\n")
	sb.WriteString("\t\treturn BasicAuthCredentials{}, errors.New(\"invalid credentials format\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn BasicAuthCredentials{\n")
	sb.WriteString("\t\tUsername: parts[0],\n")
	sb.WriteString("\t\tPassword: parts[1],\n")
	sb.WriteString("\t}, nil\n")
	sb.WriteString("}\n\n")

	// extractBearerToken
	sb.WriteString("// extractBearerToken extracts HTTP Bearer token from request\n")
	sb.WriteString("func extractBearerToken(r *http.Request) (BearerTokenCredentials, error) {\n")
	sb.WriteString("\tauth := r.Header.Get(\"Authorization\")\n")
	sb.WriteString("\tif auth == \"\" {\n")
	sb.WriteString("\t\treturn BearerTokenCredentials{}, errors.New(\"missing Authorization header\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tconst prefix = \"Bearer \"\n")
	sb.WriteString("\tif !strings.HasPrefix(auth, prefix) {\n")
	sb.WriteString("\t\treturn BearerTokenCredentials{}, errors.New(\"invalid Authorization header format\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\ttoken := strings.TrimSpace(auth[len(prefix):])\n")
	sb.WriteString("\tif token == \"\" {\n")
	sb.WriteString("\t\treturn BearerTokenCredentials{}, errors.New(\"empty bearer token\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn BearerTokenCredentials{Token: token}, nil\n")
	sb.WriteString("}\n\n")

	// extractAPIKey
	sb.WriteString("// extractAPIKey extracts API key from request (header, query, or cookie)\n")
	sb.WriteString("func extractAPIKey(r *http.Request, location, name string) (APIKeyCredentials, error) {\n")
	sb.WriteString("\tvar key string\n\n")
	sb.WriteString("\tswitch location {\n")
	sb.WriteString("\tcase \"header\":\n")
	sb.WriteString("\t\tkey = r.Header.Get(name)\n")
	sb.WriteString("\tcase \"query\":\n")
	sb.WriteString("\t\tkey = r.URL.Query().Get(name)\n")
	sb.WriteString("\tcase \"cookie\":\n")
	sb.WriteString("\t\tif cookie, err := r.Cookie(name); err == nil {\n")
	sb.WriteString("\t\t\tkey = cookie.Value\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn APIKeyCredentials{}, errors.New(\"invalid API key location\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tif key == \"\" {\n")
	sb.WriteString("\t\treturn APIKeyCredentials{}, errors.New(\"missing API key\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn APIKeyCredentials{\n")
	sb.WriteString("\t\tKey:      key,\n")
	sb.WriteString("\t\tLocation: location,\n")
	sb.WriteString("\t\tName:     name,\n")
	sb.WriteString("\t}, nil\n")
	sb.WriteString("}\n\n")

	// extractOAuth2Token
	sb.WriteString("// extractOAuth2Token extracts OAuth 2.0 token from request\n")
	sb.WriteString("func extractOAuth2Token(r *http.Request, scopes []string) (OAuth2Credentials, error) {\n")
	sb.WriteString("\t// OAuth 2.0 typically uses Bearer token\n")
	sb.WriteString("\tbearer, err := extractBearerToken(r)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn OAuth2Credentials{}, err\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn OAuth2Credentials{\n")
	sb.WriteString("\t\tToken:  bearer.Token,\n")
	sb.WriteString("\t\tScopes: scopes,\n")
	sb.WriteString("\t}, nil\n")
	sb.WriteString("}\n\n")

	// extractOpenIDConnectToken
	sb.WriteString("// extractOpenIDConnectToken extracts OpenID Connect token from request\n")
	sb.WriteString("func extractOpenIDConnectToken(r *http.Request) (OpenIDConnectCredentials, error) {\n")
	sb.WriteString("\t// OpenID Connect typically uses Bearer token\n")
	sb.WriteString("\tbearer, err := extractBearerToken(r)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn OpenIDConnectCredentials{}, err\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn OpenIDConnectCredentials{Token: bearer.Token}, nil\n")
	sb.WriteString("}\n\n")
}
