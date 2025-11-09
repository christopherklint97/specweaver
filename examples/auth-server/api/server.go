package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/christopherklint97/specweaver/pkg/router"
)

// HTTPError represents an HTTP error with a status code
type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

// NewHTTPErrorf creates a new HTTPError with formatted message
func NewHTTPErrorf(code int, format string, args ...any) *HTTPError {
	return &HTTPError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// WrapHTTPError wraps an existing error with an HTTP status code
func WrapHTTPError(code int, err error, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message, Err: err}
}

// ListUsersRequest represents the request for ListUsers
type ListUsersRequest struct {
}

// GetFlexibleRequest represents the request for GetFlexible
type GetFlexibleRequest struct {
}

// GetLegacyDataRequest represents the request for GetLegacyData
type GetLegacyDataRequest struct {
}

// GetProfileRequest represents the request for GetProfile
type GetProfileRequest struct {
}

// GetHealthRequest represents the request for GetHealth
type GetHealthRequest struct {
}

// ListResourcesRequest represents the request for ListResources
type ListResourcesRequest struct {
	Limit *int32 `json:"limit,omitempty"`
}

// CreateResourceRequest represents the request for CreateResource
type CreateResourceRequest struct {
	// Request body
	Body map[string]any `json:"body"`
}

// GetResourceRequest represents the request for GetResource
type GetResourceRequest struct {
	ResourceId int64 `json:"resourceId"`
}

// UpdateResourceRequest represents the request for UpdateResource
type UpdateResourceRequest struct {
	ResourceId int64 `json:"resourceId"`
	// Request body
	Body map[string]any `json:"body"`
}

// DeleteResourceRequest represents the request for DeleteResource
type DeleteResourceRequest struct {
	ResourceId int64 `json:"resourceId"`
}

// GetCurrentUserRequest represents the request for GetCurrentUser
type GetCurrentUserRequest struct {
}

// ListUsersResponse represents possible responses for ListUsers
type ListUsersResponse interface {
	isListUsersResponse()
	StatusCode() int
	ResponseBody() any
}

// ListUsers200Response represents a 200 response
type ListUsers200Response struct {
	Body []User `json:"body"`
}

func (r ListUsers200Response) isListUsersResponse() {}
func (r ListUsers200Response) StatusCode() int { return 200 }
func (r ListUsers200Response) ResponseBody() any { return r.Body }

// ListUsers401Response represents a 401 response
type ListUsers401Response struct {
	Body Error `json:"body"`
}

func (r ListUsers401Response) isListUsersResponse() {}
func (r ListUsers401Response) StatusCode() int { return 401 }
func (r ListUsers401Response) ResponseBody() any { return r.Body }

// GetFlexibleResponse represents possible responses for GetFlexible
type GetFlexibleResponse interface {
	isGetFlexibleResponse()
	StatusCode() int
	ResponseBody() any
}

// GetFlexible200Response represents a 200 response
type GetFlexible200Response struct {
	Body map[string]any `json:"body"`
}

func (r GetFlexible200Response) isGetFlexibleResponse() {}
func (r GetFlexible200Response) StatusCode() int { return 200 }
func (r GetFlexible200Response) ResponseBody() any { return r.Body }

// GetFlexible401Response represents a 401 response
type GetFlexible401Response struct {
	Body Error `json:"body"`
}

func (r GetFlexible401Response) isGetFlexibleResponse() {}
func (r GetFlexible401Response) StatusCode() int { return 401 }
func (r GetFlexible401Response) ResponseBody() any { return r.Body }

// GetLegacyDataResponse represents possible responses for GetLegacyData
type GetLegacyDataResponse interface {
	isGetLegacyDataResponse()
	StatusCode() int
	ResponseBody() any
}

// GetLegacyData200Response represents a 200 response
type GetLegacyData200Response struct {
	Body map[string]any `json:"body"`
}

func (r GetLegacyData200Response) isGetLegacyDataResponse() {}
func (r GetLegacyData200Response) StatusCode() int { return 200 }
func (r GetLegacyData200Response) ResponseBody() any { return r.Body }

// GetLegacyData401Response represents a 401 response
type GetLegacyData401Response struct {
	Body Error `json:"body"`
}

func (r GetLegacyData401Response) isGetLegacyDataResponse() {}
func (r GetLegacyData401Response) StatusCode() int { return 401 }
func (r GetLegacyData401Response) ResponseBody() any { return r.Body }

// GetProfileResponse represents possible responses for GetProfile
type GetProfileResponse interface {
	isGetProfileResponse()
	StatusCode() int
	ResponseBody() any
}

// GetProfile200Response represents a 200 response
type GetProfile200Response struct {
	Body User `json:"body"`
}

func (r GetProfile200Response) isGetProfileResponse() {}
func (r GetProfile200Response) StatusCode() int { return 200 }
func (r GetProfile200Response) ResponseBody() any { return r.Body }

// GetProfile401Response represents a 401 response
type GetProfile401Response struct {
	Body Error `json:"body"`
}

func (r GetProfile401Response) isGetProfileResponse() {}
func (r GetProfile401Response) StatusCode() int { return 401 }
func (r GetProfile401Response) ResponseBody() any { return r.Body }

// GetHealthResponse represents possible responses for GetHealth
type GetHealthResponse interface {
	isGetHealthResponse()
	StatusCode() int
	ResponseBody() any
}

// GetHealth200Response represents a 200 response
type GetHealth200Response struct {
	Body map[string]any `json:"body"`
}

func (r GetHealth200Response) isGetHealthResponse() {}
func (r GetHealth200Response) StatusCode() int { return 200 }
func (r GetHealth200Response) ResponseBody() any { return r.Body }

// ListResourcesResponse represents possible responses for ListResources
type ListResourcesResponse interface {
	isListResourcesResponse()
	StatusCode() int
	ResponseBody() any
}

// ListResources200Response represents a 200 response
type ListResources200Response struct {
	Body []Resource `json:"body"`
}

func (r ListResources200Response) isListResourcesResponse() {}
func (r ListResources200Response) StatusCode() int { return 200 }
func (r ListResources200Response) ResponseBody() any { return r.Body }

// ListResources401Response represents a 401 response
type ListResources401Response struct {
	Body Error `json:"body"`
}

func (r ListResources401Response) isListResourcesResponse() {}
func (r ListResources401Response) StatusCode() int { return 401 }
func (r ListResources401Response) ResponseBody() any { return r.Body }

// CreateResourceResponse represents possible responses for CreateResource
type CreateResourceResponse interface {
	isCreateResourceResponse()
	StatusCode() int
	ResponseBody() any
}

// CreateResource201Response represents a 201 response
type CreateResource201Response struct {
	Body Resource `json:"body"`
}

func (r CreateResource201Response) isCreateResourceResponse() {}
func (r CreateResource201Response) StatusCode() int { return 201 }
func (r CreateResource201Response) ResponseBody() any { return r.Body }

// CreateResource401Response represents a 401 response
type CreateResource401Response struct {
	Body Error `json:"body"`
}

func (r CreateResource401Response) isCreateResourceResponse() {}
func (r CreateResource401Response) StatusCode() int { return 401 }
func (r CreateResource401Response) ResponseBody() any { return r.Body }

// GetResourceResponse represents possible responses for GetResource
type GetResourceResponse interface {
	isGetResourceResponse()
	StatusCode() int
	ResponseBody() any
}

// GetResource200Response represents a 200 response
type GetResource200Response struct {
	Body Resource `json:"body"`
}

func (r GetResource200Response) isGetResourceResponse() {}
func (r GetResource200Response) StatusCode() int { return 200 }
func (r GetResource200Response) ResponseBody() any { return r.Body }

// GetResource401Response represents a 401 response
type GetResource401Response struct {
	Body Error `json:"body"`
}

func (r GetResource401Response) isGetResourceResponse() {}
func (r GetResource401Response) StatusCode() int { return 401 }
func (r GetResource401Response) ResponseBody() any { return r.Body }

// GetResource404Response represents a 404 response
type GetResource404Response struct {
	Body Error `json:"body"`
}

func (r GetResource404Response) isGetResourceResponse() {}
func (r GetResource404Response) StatusCode() int { return 404 }
func (r GetResource404Response) ResponseBody() any { return r.Body }

// UpdateResourceResponse represents possible responses for UpdateResource
type UpdateResourceResponse interface {
	isUpdateResourceResponse()
	StatusCode() int
	ResponseBody() any
}

// UpdateResource200Response represents a 200 response
type UpdateResource200Response struct {
	Body Resource `json:"body"`
}

func (r UpdateResource200Response) isUpdateResourceResponse() {}
func (r UpdateResource200Response) StatusCode() int { return 200 }
func (r UpdateResource200Response) ResponseBody() any { return r.Body }

// UpdateResource401Response represents a 401 response
type UpdateResource401Response struct {
	Body Error `json:"body"`
}

func (r UpdateResource401Response) isUpdateResourceResponse() {}
func (r UpdateResource401Response) StatusCode() int { return 401 }
func (r UpdateResource401Response) ResponseBody() any { return r.Body }

// UpdateResource404Response represents a 404 response
type UpdateResource404Response struct {
	Body Error `json:"body"`
}

func (r UpdateResource404Response) isUpdateResourceResponse() {}
func (r UpdateResource404Response) StatusCode() int { return 404 }
func (r UpdateResource404Response) ResponseBody() any { return r.Body }

// DeleteResourceResponse represents possible responses for DeleteResource
type DeleteResourceResponse interface {
	isDeleteResourceResponse()
	StatusCode() int
	ResponseBody() any
}

// DeleteResource204Response represents a 204 response
type DeleteResource204Response struct {
}

func (r DeleteResource204Response) isDeleteResourceResponse() {}
func (r DeleteResource204Response) StatusCode() int { return 204 }
func (r DeleteResource204Response) ResponseBody() any { return nil }

// DeleteResource401Response represents a 401 response
type DeleteResource401Response struct {
	Body Error `json:"body"`
}

func (r DeleteResource401Response) isDeleteResourceResponse() {}
func (r DeleteResource401Response) StatusCode() int { return 401 }
func (r DeleteResource401Response) ResponseBody() any { return r.Body }

// DeleteResource404Response represents a 404 response
type DeleteResource404Response struct {
	Body Error `json:"body"`
}

func (r DeleteResource404Response) isDeleteResourceResponse() {}
func (r DeleteResource404Response) StatusCode() int { return 404 }
func (r DeleteResource404Response) ResponseBody() any { return r.Body }

// GetCurrentUserResponse represents possible responses for GetCurrentUser
type GetCurrentUserResponse interface {
	isGetCurrentUserResponse()
	StatusCode() int
	ResponseBody() any
}

// GetCurrentUser200Response represents a 200 response
type GetCurrentUser200Response struct {
	Body User `json:"body"`
}

func (r GetCurrentUser200Response) isGetCurrentUserResponse() {}
func (r GetCurrentUser200Response) StatusCode() int { return 200 }
func (r GetCurrentUser200Response) ResponseBody() any { return r.Body }

// GetCurrentUser401Response represents a 401 response
type GetCurrentUser401Response struct {
	Body Error `json:"body"`
}

func (r GetCurrentUser401Response) isGetCurrentUserResponse() {}
func (r GetCurrentUser401Response) StatusCode() int { return 401 }
func (r GetCurrentUser401Response) ResponseBody() any { return r.Body }

// Server represents all server handlers
type Server interface {
	// ListUsers List all users
	ListUsers(ctx context.Context, req ListUsersRequest) (ListUsersResponse, error)
	// GetFlexible Flexible authentication
	GetFlexible(ctx context.Context, req GetFlexibleRequest) (GetFlexibleResponse, error)
	// GetLegacyData Get legacy data
	GetLegacyData(ctx context.Context, req GetLegacyDataRequest) (GetLegacyDataResponse, error)
	// GetProfile Get user profile
	GetProfile(ctx context.Context, req GetProfileRequest) (GetProfileResponse, error)
	// GetHealth Health check
	GetHealth(ctx context.Context, req GetHealthRequest) (GetHealthResponse, error)
	// ListResources List resources
	ListResources(ctx context.Context, req ListResourcesRequest) (ListResourcesResponse, error)
	// CreateResource Create resource
	CreateResource(ctx context.Context, req CreateResourceRequest) (CreateResourceResponse, error)
	// GetResource Get resource
	GetResource(ctx context.Context, req GetResourceRequest) (GetResourceResponse, error)
	// UpdateResource Update resource
	UpdateResource(ctx context.Context, req UpdateResourceRequest) (UpdateResourceResponse, error)
	// DeleteResource Delete resource
	DeleteResource(ctx context.Context, req DeleteResourceRequest) (DeleteResourceResponse, error)
	// GetCurrentUser Get current user
	GetCurrentUser(ctx context.Context, req GetCurrentUserRequest) (GetCurrentUserResponse, error)
}

// ServerWrapper wraps the Server with HTTP handler logic
type ServerWrapper struct {
	Handler Server
}

// handleListUsers adapts HTTP request to ListUsers handler
func (w *ServerWrapper) handleListUsers(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := ListUsersRequest{}

	// Call handler
	resp, err := w.Handler.ListUsers(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleGetFlexible adapts HTTP request to GetFlexible handler
func (w *ServerWrapper) handleGetFlexible(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := GetFlexibleRequest{}

	// Call handler
	resp, err := w.Handler.GetFlexible(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleGetLegacyData adapts HTTP request to GetLegacyData handler
func (w *ServerWrapper) handleGetLegacyData(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := GetLegacyDataRequest{}

	// Call handler
	resp, err := w.Handler.GetLegacyData(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleGetProfile adapts HTTP request to GetProfile handler
func (w *ServerWrapper) handleGetProfile(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := GetProfileRequest{}

	// Call handler
	resp, err := w.Handler.GetProfile(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleGetHealth adapts HTTP request to GetHealth handler
func (w *ServerWrapper) handleGetHealth(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := GetHealthRequest{}

	// Call handler
	resp, err := w.Handler.GetHealth(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleListResources adapts HTTP request to ListResources handler
func (w *ServerWrapper) handleListResources(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := ListResourcesRequest{}

	// Parse query parameter: limit
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limitVal, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil {
			limitTyped := int32(limitVal)
			req.Limit = &limitTyped
		}
	}

	// Call handler
	resp, err := w.Handler.ListResources(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleCreateResource adapts HTTP request to CreateResource handler
func (w *ServerWrapper) handleCreateResource(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := CreateResourceRequest{}

	// Parse request body
	if err := ReadJSON(r, &req.Body); err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid request body"))
		return
	}

	// Call handler
	resp, err := w.Handler.CreateResource(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleGetResource adapts HTTP request to GetResource handler
func (w *ServerWrapper) handleGetResource(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := GetResourceRequest{}

	// Parse path parameter: resourceId
	resourceIdStr := router.URLParam(r, "resourceId")
	resourceIdVal, err := strconv.ParseInt(resourceIdStr, 10, 64)
	if err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid resourceId parameter"))
		return
	}
	req.ResourceId = int64(resourceIdVal)

	// Call handler
	resp, err := w.Handler.GetResource(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleUpdateResource adapts HTTP request to UpdateResource handler
func (w *ServerWrapper) handleUpdateResource(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := UpdateResourceRequest{}

	// Parse path parameter: resourceId
	resourceIdStr := router.URLParam(r, "resourceId")
	resourceIdVal, err := strconv.ParseInt(resourceIdStr, 10, 64)
	if err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid resourceId parameter"))
		return
	}
	req.ResourceId = int64(resourceIdVal)

	// Parse request body
	if err := ReadJSON(r, &req.Body); err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid request body"))
		return
	}

	// Call handler
	resp, err := w.Handler.UpdateResource(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleDeleteResource adapts HTTP request to DeleteResource handler
func (w *ServerWrapper) handleDeleteResource(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := DeleteResourceRequest{}

	// Parse path parameter: resourceId
	resourceIdStr := router.URLParam(r, "resourceId")
	resourceIdVal, err := strconv.ParseInt(resourceIdStr, 10, 64)
	if err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid resourceId parameter"))
		return
	}
	req.ResourceId = int64(resourceIdVal)

	// Call handler
	resp, err := w.Handler.DeleteResource(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleGetCurrentUser adapts HTTP request to GetCurrentUser handler
func (w *ServerWrapper) handleGetCurrentUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := GetCurrentUserRequest{}

	// Call handler
	resp, err := w.Handler.GetCurrentUser(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleError handles errors and writes appropriate HTTP responses
func (w *ServerWrapper) handleError(rw http.ResponseWriter, err error) {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		WriteError(rw, httpErr.Code, httpErr)
		return
	}
	// Default to 500 Internal Server Error
	WriteError(rw, http.StatusInternalServerError, err)
}

// securitySchemeInfoMap contains information about all security schemes
var securitySchemeInfoMap = map[string]*SecuritySchemeInfo{
	"apiKeyCookie": {
		Type:   "apiKey",
		In:     "cookie",
		Name:   "session_id",
	},
	"apiKeyHeader": {
		Type:   "apiKey",
		In:     "header",
		Name:   "X-API-Key",
	},
	"apiKeyQuery": {
		Type:   "apiKey",
		In:     "query",
		Name:   "api_key",
	},
	"basicAuth": {
		Type:   "http",
		Scheme: "basic",
	},
	"bearerAuth": {
		Type:   "http",
		Scheme: "bearer",
	},
	"oauth2Auth": {
		Type:   "oauth2",
	},
	"openIdAuth": {
		Type:   "openIdConnect",
	},
}

// ConfigureRouter configures the given router with all routes.
// This function allows you to use any router that implements the router.Router interface.
//
// The authenticator parameter is optional. If nil, no authentication will be performed.
// If provided, authentication will be enforced for routes that require it.
//
// Example with built-in router:
//
//	r := router.NewRouter()
//	ConfigureRouter(r, myServer, myAuthenticator)
//
// Example with custom router:
//
//	r := myCustomRouter.New() // Must implement router.Router interface
//	ConfigureRouter(r, myServer, myAuthenticator)
func ConfigureRouter(r router.Router, si Server, authenticator Authenticator) {
	wrapper := &ServerWrapper{Handler: si}

	r.Get("/admin/users", authMiddleware(authenticator, []map[string][]string{
		{
			"basicAuth": []string{},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleListUsers)).ServeHTTP)
	r.Get("/flexible", authMiddleware(authenticator, []map[string][]string{
		{
			"bearerAuth": []string{},
		},
		{
			"apiKeyHeader": []string{},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleGetFlexible)).ServeHTTP)
	r.Get("/legacy/data", authMiddleware(authenticator, []map[string][]string{
		{
			"apiKeyQuery": []string{},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleGetLegacyData)).ServeHTTP)
	r.Get("/profile", authMiddleware(authenticator, []map[string][]string{
		{
			"openIdAuth": []string{},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleGetProfile)).ServeHTTP)
	r.Get("/public/health", wrapper.handleGetHealth)
	r.Get("/resources", authMiddleware(authenticator, []map[string][]string{
		{
			"apiKeyHeader": []string{},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleListResources)).ServeHTTP)
	r.Post("/resources", authMiddleware(authenticator, []map[string][]string{
		{
			"apiKeyHeader": []string{},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleCreateResource)).ServeHTTP)
	r.Get("/resources/{resourceId}", authMiddleware(authenticator, []map[string][]string{
		{
			"oauth2Auth": []string{"read"},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleGetResource)).ServeHTTP)
	r.Put("/resources/{resourceId}", authMiddleware(authenticator, []map[string][]string{
		{
			"oauth2Auth": []string{"write"},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleUpdateResource)).ServeHTTP)
	r.Delete("/resources/{resourceId}", authMiddleware(authenticator, []map[string][]string{
		{
			"oauth2Auth": []string{"admin"},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleDeleteResource)).ServeHTTP)
	r.Get("/users/me", authMiddleware(authenticator, []map[string][]string{
		{
			"bearerAuth": []string{},
		},
	}, securitySchemeInfoMap)(http.HandlerFunc(wrapper.handleGetCurrentUser)).ServeHTTP)
}

// NewRouter creates a new router with all routes configured using the built-in router.
// For using a custom router, use ConfigureRouter instead.
//
// The authenticator parameter is optional. If nil, no authentication will be performed.
func NewRouter(si Server, authenticator Authenticator) *router.Mux {
	r := router.NewRouter()

	// Default middleware
	r.Use(router.Logger)
	r.Use(router.Recoverer)
	r.Use(router.RequestID)
	r.Use(router.RealIP)

	ConfigureRouter(r, si, authenticator)
	return r
}

// Helper functions for request/response handling

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// WriteResponse writes a response based on its type
func WriteResponse(w http.ResponseWriter, resp any) error {
	// Extract status code and body using type assertion
	type responseWriter interface {
		StatusCode() int
		ResponseBody() any
	}

	if rw, ok := resp.(responseWriter); ok {
		statusCode := rw.StatusCode()
		body := rw.ResponseBody()
		// For 204 No Content or nil body, don't write a body
		if statusCode == http.StatusNoContent || body == nil {
			w.WriteHeader(statusCode)
			return nil
		}
		return WriteJSON(w, statusCode, body)
	}
	// Fallback to 200 OK
	return WriteJSON(w, http.StatusOK, resp)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: err.Error(),
	})
}

// ReadJSON reads and decodes JSON from request body
func ReadJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

