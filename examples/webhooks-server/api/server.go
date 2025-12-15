package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

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

// CreateSubscriptionRequest represents the request for CreateSubscription
type CreateSubscriptionRequest struct {
	// Request body
	Body Subscription `json:"body"`
}

// DeleteSubscriptionRequest represents the request for DeleteSubscription
type DeleteSubscriptionRequest struct {
	Id string `json:"id"`
}

// CreateSubscriptionResponse represents possible responses for CreateSubscription
type CreateSubscriptionResponse interface {
	isCreateSubscriptionResponse()
	StatusCode() int
	ResponseBody() any
}

// CreateSubscription201Response represents a 201 response
type CreateSubscription201Response struct {
	Body Subscription `json:"body"`
}

func (r CreateSubscription201Response) isCreateSubscriptionResponse() {}
func (r CreateSubscription201Response) StatusCode() int { return 201 }
func (r CreateSubscription201Response) ResponseBody() any { return r.Body }

// CreateSubscription400Response represents a 400 response
type CreateSubscription400Response struct {
	Body Error `json:"body"`
}

func (r CreateSubscription400Response) isCreateSubscriptionResponse() {}
func (r CreateSubscription400Response) StatusCode() int { return 400 }
func (r CreateSubscription400Response) ResponseBody() any { return r.Body }

// DeleteSubscriptionResponse represents possible responses for DeleteSubscription
type DeleteSubscriptionResponse interface {
	isDeleteSubscriptionResponse()
	StatusCode() int
	ResponseBody() any
}

// DeleteSubscription204Response represents a 204 response
type DeleteSubscription204Response struct {
}

func (r DeleteSubscription204Response) isDeleteSubscriptionResponse() {}
func (r DeleteSubscription204Response) StatusCode() int { return 204 }
func (r DeleteSubscription204Response) ResponseBody() any { return nil }

// DeleteSubscription404Response represents a 404 response
type DeleteSubscription404Response struct {
	Body Error `json:"body"`
}

func (r DeleteSubscription404Response) isDeleteSubscriptionResponse() {}
func (r DeleteSubscription404Response) StatusCode() int { return 404 }
func (r DeleteSubscription404Response) ResponseBody() any { return r.Body }

// Server represents all server handlers
type Server interface {
	// CreateSubscription Subscribe to webhooks
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (CreateSubscriptionResponse, error)
	// DeleteSubscription Unsubscribe from webhooks
	DeleteSubscription(ctx context.Context, req DeleteSubscriptionRequest) (DeleteSubscriptionResponse, error)
}

// ServerWrapper wraps the Server with HTTP handler logic
type ServerWrapper struct {
	Handler Server
}

// handleCreateSubscription adapts HTTP request to CreateSubscription handler
func (w *ServerWrapper) handleCreateSubscription(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := CreateSubscriptionRequest{}

	// Parse request body
	if err := ReadJSON(r, &req.Body); err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid request body"))
		return
	}

	// Call handler
	resp, err := w.Handler.CreateSubscription(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleDeleteSubscription adapts HTTP request to DeleteSubscription handler
func (w *ServerWrapper) handleDeleteSubscription(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := DeleteSubscriptionRequest{}

	// Parse path parameter: id
	idStr := router.URLParam(r, "id")
	req.Id = idStr

	// Call handler
	resp, err := w.Handler.DeleteSubscription(ctx, req)
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
func ConfigureRouter(r router.Router, si Server) {
	wrapper := &ServerWrapper{Handler: si}

	r.Post("/subscriptions", wrapper.handleCreateSubscription)
	r.Delete("/subscriptions/{id}", wrapper.handleDeleteSubscription)
}

// NewRouter creates a new router with all routes configured using the built-in router.
// For using a custom router, use ConfigureRouter instead.
func NewRouter(si Server) *router.Mux {
	r := router.NewRouter()

	// Default middleware
	r.Use(router.Logger)
	r.Use(router.Recoverer)
	r.Use(router.RequestID)
	r.Use(router.RealIP)

	ConfigureRouter(r, si)
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

