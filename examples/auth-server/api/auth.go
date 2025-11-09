package api

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

// securityContextKey is the context key for security information
const securityContextKey contextKey = "security"

// Credential types for different authentication schemes

// BasicAuthCredentials holds HTTP Basic authentication credentials
type BasicAuthCredentials struct {
	Username string
	Password string
}

// BearerTokenCredentials holds HTTP Bearer token credentials
type BearerTokenCredentials struct {
	Token string
}

// APIKeyCredentials holds API key credentials
type APIKeyCredentials struct {
	Key      string
	Location string // "header", "query", or "cookie"
	Name     string // The name of the header, query param, or cookie
}

// OAuth2Credentials holds OAuth 2.0 credentials
type OAuth2Credentials struct {
	Token  string
	Scopes []string
}

// OpenIDConnectCredentials holds OpenID Connect credentials
type OpenIDConnectCredentials struct {
	Token string
}

// SecurityContext holds authentication information
type SecurityContext struct {
	// Principal is the authenticated user/entity (type depends on your implementation)
	Principal any
	// SchemeName is the name of the security scheme that was used
	SchemeName string
	// Scopes are the OAuth2 scopes (if applicable)
	Scopes []string
}

// GetSecurityContext retrieves the security context from the request context
// Returns nil if no authentication was performed
func GetSecurityContext(ctx context.Context) *SecurityContext {
	if sc, ok := ctx.Value(securityContextKey).(*SecurityContext); ok {
		return sc
	}
	return nil
}

// Authenticator defines the interface for authentication handlers
// Implement this interface to provide authentication logic for your API
type Authenticator interface {
	// AuthenticateApiKeyCookie authenticates using API Key
	// Returns the authenticated principal or an error
	AuthenticateApiKeyCookie(ctx context.Context, credentials APIKeyCredentials) (any, error)

	// AuthenticateApiKeyHeader authenticates using API Key
	// Returns the authenticated principal or an error
	AuthenticateApiKeyHeader(ctx context.Context, credentials APIKeyCredentials) (any, error)

	// AuthenticateApiKeyQuery authenticates using API Key
	// Returns the authenticated principal or an error
	AuthenticateApiKeyQuery(ctx context.Context, credentials APIKeyCredentials) (any, error)

	// AuthenticateBasicAuth authenticates using HTTP Basic Auth
	// Returns the authenticated principal or an error
	AuthenticateBasicAuth(ctx context.Context, credentials BasicAuthCredentials) (any, error)

	// AuthenticateBearerAuth authenticates using HTTP Bearer token
	// Returns the authenticated principal or an error
	AuthenticateBearerAuth(ctx context.Context, credentials BearerTokenCredentials) (any, error)

	// AuthenticateOauth2Auth authenticates using OAuth 2.0
	// Returns the authenticated principal or an error
	AuthenticateOauth2Auth(ctx context.Context, credentials OAuth2Credentials) (any, error)

	// AuthenticateOpenIdAuth authenticates using OpenID Connect
	// Returns the authenticated principal or an error
	AuthenticateOpenIdAuth(ctx context.Context, credentials OpenIDConnectCredentials) (any, error)

}

// authMiddleware creates authentication middleware for an operation
func authMiddleware(authenticator Authenticator, securityReqs []map[string][]string, schemes map[string]*SecuritySchemeInfo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// If no authenticator provided, skip authentication
			if authenticator == nil {
				next.ServeHTTP(w, r)
				return
			}

			// If no security requirements, continue without authentication
			if len(securityReqs) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Try each security requirement (OR logic)
			for _, req := range securityReqs {
				// All schemes in a requirement must be satisfied (AND logic)
				var secCtx *SecurityContext
				var authErr error
				allSatisfied := true

				for schemeName, scopes := range req {
					schemeInfo, exists := schemes[schemeName]
					if !exists {
						allSatisfied = false
						break
					}

					// Authenticate based on scheme type
					var principal any
					switch schemeInfo.Type {
					case "http":
						if schemeInfo.Scheme == "basic" {
							creds, err := extractBasicAuth(r)
							if err != nil {
								allSatisfied = false
								authErr = err
								break
							}
							principal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)
						} else if schemeInfo.Scheme == "bearer" {
							creds, err := extractBearerToken(r)
							if err != nil {
								allSatisfied = false
								authErr = err
								break
							}
							principal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)
						}
					case "apiKey":
						creds, err := extractAPIKey(r, schemeInfo.In, schemeInfo.Name)
						if err != nil {
							allSatisfied = false
							authErr = err
							break
						}
						principal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)
					case "oauth2":
						creds, err := extractOAuth2Token(r, scopes)
						if err != nil {
							allSatisfied = false
							authErr = err
							break
						}
						principal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)
					case "openIdConnect":
						creds, err := extractOpenIDConnectToken(r)
						if err != nil {
							allSatisfied = false
							authErr = err
							break
						}
						principal, authErr = callAuthenticator(authenticator, schemeName, ctx, creds)
					default:
						allSatisfied = false
						authErr = errors.New("unsupported security scheme type")
					}

					if authErr != nil {
						allSatisfied = false
						break
					}

					// Create or update security context
					secCtx = &SecurityContext{
						Principal:  principal,
						SchemeName: schemeName,
						Scopes:     scopes,
					}
				}

				// If all schemes in this requirement were satisfied, continue
				if allSatisfied && secCtx != nil {
					ctx = context.WithValue(ctx, securityContextKey, secCtx)
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
					return
				}
			}

			// None of the security requirements were satisfied
			WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		})
	}
}

// callAuthenticator calls the appropriate authenticator method based on scheme name
func callAuthenticator(authenticator Authenticator, schemeName string, ctx context.Context, credentials any) (any, error) {
	if authenticator == nil {
		return nil, errors.New("no authenticator provided")
	}

	switch schemeName {
	case "apiKeyCookie":
		if creds, ok := credentials.(APIKeyCredentials); ok {
			return authenticator.AuthenticateApiKeyCookie(ctx, creds)
		}
	case "apiKeyHeader":
		if creds, ok := credentials.(APIKeyCredentials); ok {
			return authenticator.AuthenticateApiKeyHeader(ctx, creds)
		}
	case "apiKeyQuery":
		if creds, ok := credentials.(APIKeyCredentials); ok {
			return authenticator.AuthenticateApiKeyQuery(ctx, creds)
		}
	case "basicAuth":
		if creds, ok := credentials.(BasicAuthCredentials); ok {
			return authenticator.AuthenticateBasicAuth(ctx, creds)
		}
	case "bearerAuth":
		if creds, ok := credentials.(BearerTokenCredentials); ok {
			return authenticator.AuthenticateBearerAuth(ctx, creds)
		}
	case "oauth2Auth":
		if creds, ok := credentials.(OAuth2Credentials); ok {
			return authenticator.AuthenticateOauth2Auth(ctx, creds)
		}
	case "openIdAuth":
		if creds, ok := credentials.(OpenIDConnectCredentials); ok {
			return authenticator.AuthenticateOpenIdAuth(ctx, creds)
		}
	}

	return nil, errors.New("unknown security scheme")
}

// SecuritySchemeInfo holds information about a security scheme
type SecuritySchemeInfo struct {
	Type   string
	Scheme string
	In     string
	Name   string
}

// Credential extraction helpers

// extractBasicAuth extracts HTTP Basic Auth credentials from request
func extractBasicAuth(r *http.Request) (BasicAuthCredentials, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return BasicAuthCredentials{}, errors.New("missing Authorization header")
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return BasicAuthCredentials{}, errors.New("invalid Authorization header format")
	}

	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return BasicAuthCredentials{}, errors.New("invalid base64 encoding")
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return BasicAuthCredentials{}, errors.New("invalid credentials format")
	}

	return BasicAuthCredentials{
		Username: parts[0],
		Password: parts[1],
	}, nil
}

// extractBearerToken extracts HTTP Bearer token from request
func extractBearerToken(r *http.Request) (BearerTokenCredentials, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return BearerTokenCredentials{}, errors.New("missing Authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return BearerTokenCredentials{}, errors.New("invalid Authorization header format")
	}

	token := strings.TrimSpace(auth[len(prefix):])
	if token == "" {
		return BearerTokenCredentials{}, errors.New("empty bearer token")
	}

	return BearerTokenCredentials{Token: token}, nil
}

// extractAPIKey extracts API key from request (header, query, or cookie)
func extractAPIKey(r *http.Request, location, name string) (APIKeyCredentials, error) {
	var key string

	switch location {
	case "header":
		key = r.Header.Get(name)
	case "query":
		key = r.URL.Query().Get(name)
	case "cookie":
		if cookie, err := r.Cookie(name); err == nil {
			key = cookie.Value
		}
	default:
		return APIKeyCredentials{}, errors.New("invalid API key location")
	}

	if key == "" {
		return APIKeyCredentials{}, errors.New("missing API key")
	}

	return APIKeyCredentials{
		Key:      key,
		Location: location,
		Name:     name,
	}, nil
}

// extractOAuth2Token extracts OAuth 2.0 token from request
func extractOAuth2Token(r *http.Request, scopes []string) (OAuth2Credentials, error) {
	// OAuth 2.0 typically uses Bearer token
	bearer, err := extractBearerToken(r)
	if err != nil {
		return OAuth2Credentials{}, err
	}

	return OAuth2Credentials{
		Token:  bearer.Token,
		Scopes: scopes,
	}, nil
}

// extractOpenIDConnectToken extracts OpenID Connect token from request
func extractOpenIDConnectToken(r *http.Request) (OpenIDConnectCredentials, error) {
	// OpenID Connect typically uses Bearer token
	bearer, err := extractBearerToken(r)
	if err != nil {
		return OpenIDConnectCredentials{}, err
	}

	return OpenIDConnectCredentials{Token: bearer.Token}, nil
}

