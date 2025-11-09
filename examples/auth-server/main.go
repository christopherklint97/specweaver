package main

import (
	"context"
	"log"
	"net/http"

	api "github.com/christopherklint97/specweaver/examples/auth-server/api"
)

// MyAuthenticator implements the authentication interface
type MyAuthenticator struct {
	// In a real application, you would have database connections, etc.
}

// AuthenticateBasicAuth validates HTTP Basic Auth credentials
func (a *MyAuthenticator) AuthenticateBasicAuth(ctx context.Context, credentials api.BasicAuthCredentials) (any, error) {
	// In a real app, check against database
	if credentials.Username == "admin" && credentials.Password == "secret" {
		return &api.User{
			Id:       1,
			Username: "admin",
			Email:    "admin@example.com",
			Role:     "admin",
		}, nil
	}
	return nil, api.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
}

// AuthenticateBearerAuth validates Bearer token
func (a *MyAuthenticator) AuthenticateBearerAuth(ctx context.Context, credentials api.BearerTokenCredentials) (any, error) {
	// In a real app, validate JWT token
	if credentials.Token == "valid-token-123" {
		return &api.User{
			Id:       2,
			Username: "user1",
			Email:    "user1@example.com",
			Role:     "user",
		}, nil
	}
	return nil, api.NewHTTPError(http.StatusUnauthorized, "invalid token")
}

// AuthenticateApiKeyHeader validates API key from header
func (a *MyAuthenticator) AuthenticateApiKeyHeader(ctx context.Context, credentials api.APIKeyCredentials) (any, error) {
	// In a real app, check against database
	if credentials.Key == "valid-api-key" {
		return &api.User{
			Id:       3,
			Username: "api-user",
			Email:    "api@example.com",
			Role:     "user",
		}, nil
	}
	return nil, api.NewHTTPError(http.StatusUnauthorized, "invalid API key")
}

// AuthenticateApiKeyQuery validates API key from query
func (a *MyAuthenticator) AuthenticateApiKeyQuery(ctx context.Context, credentials api.APIKeyCredentials) (any, error) {
	// Similar to header validation
	if credentials.Key == "legacy-key-456" {
		return &api.User{
			Id:       4,
			Username: "legacy-user",
			Email:    "legacy@example.com",
			Role:     "user",
		}, nil
	}
	return nil, api.NewHTTPError(http.StatusUnauthorized, "invalid API key")
}

// AuthenticateApiKeyCookie validates API key from cookie
func (a *MyAuthenticator) AuthenticateApiKeyCookie(ctx context.Context, credentials api.APIKeyCredentials) (any, error) {
	// Check session ID
	if credentials.Key == "valid-session" {
		return &api.User{
			Id:       5,
			Username: "session-user",
			Email:    "session@example.com",
			Role:     "user",
		}, nil
	}
	return nil, api.NewHTTPError(http.StatusUnauthorized, "invalid session")
}

// AuthenticateOauth2Auth validates OAuth2 token and scopes
func (a *MyAuthenticator) AuthenticateOauth2Auth(ctx context.Context, credentials api.OAuth2Credentials) (any, error) {
	// In a real app, validate OAuth2 token and check scopes
	if credentials.Token == "oauth-token-789" {
		// Check if user has required scopes
		return &api.User{
			Id:       6,
			Username: "oauth-user",
			Email:    "oauth@example.com",
			Role:     "user",
		}, nil
	}
	return nil, api.NewHTTPError(http.StatusUnauthorized, "invalid OAuth2 token")
}

// AuthenticateOpenIdAuth validates OpenID Connect token
func (a *MyAuthenticator) AuthenticateOpenIdAuth(ctx context.Context, credentials api.OpenIDConnectCredentials) (any, error) {
	// In a real app, validate OIDC token
	if credentials.Token == "oidc-token-abc" {
		return &api.User{
			Id:       7,
			Username: "oidc-user",
			Email:    "oidc@example.com",
			Role:     "user",
		}, nil
	}
	return nil, api.NewHTTPError(http.StatusUnauthorized, "invalid OIDC token")
}

// MyServer implements the Server interface
type MyServer struct {
	// Your application state
}

func (s *MyServer) GetHealth(ctx context.Context, req api.GetHealthRequest) (api.GetHealthResponse, error) {
	return api.GetHealth200Response{
		Body: map[string]any{
			"status": "ok",
		},
	}, nil
}

func (s *MyServer) GetCurrentUser(ctx context.Context, req api.GetCurrentUserRequest) (api.GetCurrentUserResponse, error) {
	// Get the authenticated user from context
	secCtx := api.GetSecurityContext(ctx)
	if secCtx == nil {
		return nil, api.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}

	user, ok := secCtx.Principal.(*api.User)
	if !ok {
		return nil, api.NewHTTPError(http.StatusInternalServerError, "invalid principal type")
	}

	return api.GetCurrentUser200Response{Body: *user}, nil
}

func (s *MyServer) ListUsers(ctx context.Context, req api.ListUsersRequest) (api.ListUsersResponse, error) {
	// Check if user is admin
	secCtx := api.GetSecurityContext(ctx)
	if secCtx == nil {
		return nil, api.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}

	user, ok := secCtx.Principal.(*api.User)
	if !ok || user.Role != "admin" {
		return nil, api.NewHTTPError(http.StatusForbidden, "admin access required")
	}

	users := []api.User{
		{Id: 1, Username: "admin", Email: "admin@example.com", Role: "admin"},
		{Id: 2, Username: "user1", Email: "user1@example.com", Role: "user"},
	}

	return api.ListUsers200Response{Body: users}, nil
}

func (s *MyServer) ListResources(ctx context.Context, req api.ListResourcesRequest) (api.ListResourcesResponse, error) {
	resources := []api.Resource{
		{Id: 1, Name: "Resource 1", OwnerId: 1},
		{Id: 2, Name: "Resource 2", OwnerId: 2},
	}

	return api.ListResources200Response{Body: resources}, nil
}

func (s *MyServer) CreateResource(ctx context.Context, req api.CreateResourceRequest) (api.CreateResourceResponse, error) {
	name, _ := req.Body["name"].(string)
	resource := api.Resource{
		Id:      3,
		Name:    name,
		OwnerId: 1,
	}

	return api.CreateResource201Response{Body: resource}, nil
}

func (s *MyServer) GetLegacyData(ctx context.Context, req api.GetLegacyDataRequest) (api.GetLegacyDataResponse, error) {
	return api.GetLegacyData200Response{
		Body: map[string]any{
			"data": "legacy data",
		},
	}, nil
}

func (s *MyServer) GetResource(ctx context.Context, req api.GetResourceRequest) (api.GetResourceResponse, error) {
	if req.ResourceId == 1 {
		return api.GetResource200Response{
			Body: api.Resource{
				Id:      1,
				Name:    "Resource 1",
				OwnerId: 1,
			},
		}, nil
	}

	return api.GetResource404Response{
		Body: api.Error{
			Error:   "Not Found",
			Message: "resource not found",
		},
	}, nil
}

func (s *MyServer) UpdateResource(ctx context.Context, req api.UpdateResourceRequest) (api.UpdateResourceResponse, error) {
	if req.ResourceId == 1 {
		name, _ := req.Body["name"].(string)
		return api.UpdateResource200Response{
			Body: api.Resource{
				Id:      1,
				Name:    name,
				OwnerId: 1,
			},
		}, nil
	}

	return api.UpdateResource404Response{
		Body: api.Error{
			Error:   "Not Found",
			Message: "resource not found",
		},
	}, nil
}

func (s *MyServer) DeleteResource(ctx context.Context, req api.DeleteResourceRequest) (api.DeleteResourceResponse, error) {
	if req.ResourceId == 1 {
		return api.DeleteResource204Response{}, nil
	}

	return api.DeleteResource404Response{
		Body: api.Error{
			Error:   "Not Found",
			Message: "resource not found",
		},
	}, nil
}

func (s *MyServer) GetProfile(ctx context.Context, req api.GetProfileRequest) (api.GetProfileResponse, error) {
	secCtx := api.GetSecurityContext(ctx)
	if secCtx == nil {
		return nil, api.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}

	user, ok := secCtx.Principal.(*api.User)
	if !ok {
		return nil, api.NewHTTPError(http.StatusInternalServerError, "invalid principal type")
	}

	return api.GetProfile200Response{Body: *user}, nil
}

func (s *MyServer) GetFlexible(ctx context.Context, req api.GetFlexibleRequest) (api.GetFlexibleResponse, error) {
	return api.GetFlexible200Response{
		Body: map[string]any{
			"message": "flexible auth works",
		},
	}, nil
}

func main() {
	server := &MyServer{}
	authenticator := &MyAuthenticator{}

	router := api.NewRouter(server, authenticator)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
