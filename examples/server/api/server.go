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

// ListPetsRequest represents the request for ListPets
type ListPetsRequest struct {
	// Maximum number of pets to return
	Limit *int32 `json:"limit,omitempty"`
	// Filter pets by tag
	Tag *string `json:"tag,omitempty"`
}

// CreatePetRequest represents the request for CreatePet
type CreatePetRequest struct {
	// Request body
	Body NewPet `json:"body"`
}

// UpdatePetRequest represents the request for UpdatePet
type UpdatePetRequest struct {
	// The ID of the pet to update
	PetId int64 `json:"petId"`
	// Request body
	Body NewPet `json:"body"`
}

// DeletePetRequest represents the request for DeletePet
type DeletePetRequest struct {
	// The ID of the pet to delete
	PetId int64 `json:"petId"`
}

// GetPetByIdRequest represents the request for GetPetById
type GetPetByIdRequest struct {
	// The ID of the pet to retrieve
	PetId int64 `json:"petId"`
}

// ListPetsResponse represents possible responses for ListPets
type ListPetsResponse interface {
	isListPetsResponse()
	StatusCode() int
	ResponseBody() any
}

// ListPets200Response represents a 200 response
type ListPets200Response struct {
	Body []Pet `json:"body"`
}

func (r ListPets200Response) isListPetsResponse() {}
func (r ListPets200Response) StatusCode() int { return 200 }
func (r ListPets200Response) ResponseBody() any { return r.Body }

// ListPets500Response represents a 500 response
type ListPets500Response struct {
	Body Error `json:"body"`
}

func (r ListPets500Response) isListPetsResponse() {}
func (r ListPets500Response) StatusCode() int { return 500 }
func (r ListPets500Response) ResponseBody() any { return r.Body }

// CreatePetResponse represents possible responses for CreatePet
type CreatePetResponse interface {
	isCreatePetResponse()
	StatusCode() int
	ResponseBody() any
}

// CreatePet201Response represents a 201 response
type CreatePet201Response struct {
	Body Pet `json:"body"`
}

func (r CreatePet201Response) isCreatePetResponse() {}
func (r CreatePet201Response) StatusCode() int { return 201 }
func (r CreatePet201Response) ResponseBody() any { return r.Body }

// CreatePet400Response represents a 400 response
type CreatePet400Response struct {
	Body Error `json:"body"`
}

func (r CreatePet400Response) isCreatePetResponse() {}
func (r CreatePet400Response) StatusCode() int { return 400 }
func (r CreatePet400Response) ResponseBody() any { return r.Body }

// GetPetByIdResponse represents possible responses for GetPetById
type GetPetByIdResponse interface {
	isGetPetByIdResponse()
	StatusCode() int
	ResponseBody() any
}

// GetPetById200Response represents a 200 response
type GetPetById200Response struct {
	Body Pet `json:"body"`
}

func (r GetPetById200Response) isGetPetByIdResponse() {}
func (r GetPetById200Response) StatusCode() int { return 200 }
func (r GetPetById200Response) ResponseBody() any { return r.Body }

// GetPetById404Response represents a 404 response
type GetPetById404Response struct {
	Body Error `json:"body"`
}

func (r GetPetById404Response) isGetPetByIdResponse() {}
func (r GetPetById404Response) StatusCode() int { return 404 }
func (r GetPetById404Response) ResponseBody() any { return r.Body }

// UpdatePetResponse represents possible responses for UpdatePet
type UpdatePetResponse interface {
	isUpdatePetResponse()
	StatusCode() int
	ResponseBody() any
}

// UpdatePet200Response represents a 200 response
type UpdatePet200Response struct {
	Body Pet `json:"body"`
}

func (r UpdatePet200Response) isUpdatePetResponse() {}
func (r UpdatePet200Response) StatusCode() int { return 200 }
func (r UpdatePet200Response) ResponseBody() any { return r.Body }

// UpdatePet404Response represents a 404 response
type UpdatePet404Response struct {
	Body Error `json:"body"`
}

func (r UpdatePet404Response) isUpdatePetResponse() {}
func (r UpdatePet404Response) StatusCode() int { return 404 }
func (r UpdatePet404Response) ResponseBody() any { return r.Body }

// DeletePetResponse represents possible responses for DeletePet
type DeletePetResponse interface {
	isDeletePetResponse()
	StatusCode() int
	ResponseBody() any
}

// DeletePet404Response represents a 404 response
type DeletePet404Response struct {
	Body Error `json:"body"`
}

func (r DeletePet404Response) isDeletePetResponse() {}
func (r DeletePet404Response) StatusCode() int { return 404 }
func (r DeletePet404Response) ResponseBody() any { return r.Body }

// DeletePet204Response represents a 204 response
type DeletePet204Response struct {
}

func (r DeletePet204Response) isDeletePetResponse() {}
func (r DeletePet204Response) StatusCode() int { return 204 }
func (r DeletePet204Response) ResponseBody() any { return nil }

// Server represents all server handlers
type Server interface {
	// ListPets List all pets
	ListPets(ctx context.Context, req ListPetsRequest) (ListPetsResponse, error)
	// CreatePet Create a pet
	CreatePet(ctx context.Context, req CreatePetRequest) (CreatePetResponse, error)
	// DeletePet Delete a pet
	DeletePet(ctx context.Context, req DeletePetRequest) (DeletePetResponse, error)
	// GetPetById Get a pet by ID
	GetPetById(ctx context.Context, req GetPetByIdRequest) (GetPetByIdResponse, error)
	// UpdatePet Update a pet
	UpdatePet(ctx context.Context, req UpdatePetRequest) (UpdatePetResponse, error)
}

// ServerWrapper wraps the Server with HTTP handler logic
type ServerWrapper struct {
	Handler Server
}

// handleCreatePet adapts HTTP request to CreatePet handler
func (w *ServerWrapper) handleCreatePet(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := CreatePetRequest{}

	// Parse request body
	if err := ReadJSON(r, &req.Body); err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid request body"))
		return
	}

	// Call handler
	resp, err := w.Handler.CreatePet(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleListPets adapts HTTP request to ListPets handler
func (w *ServerWrapper) handleListPets(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := ListPetsRequest{}

	// Parse query parameter: limit
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limitVal, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil {
			limitTyped := int32(limitVal)
			req.Limit = &limitTyped
		}
	}

	// Parse query parameter: tag
	tagStr := r.URL.Query().Get("tag")
	if tagStr != "" {
		req.Tag = &tagStr
	}

	// Call handler
	resp, err := w.Handler.ListPets(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleGetPetById adapts HTTP request to GetPetById handler
func (w *ServerWrapper) handleGetPetById(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := GetPetByIdRequest{}

	// Parse path parameter: petId
	petIdStr := router.URLParam(r, "petId")
	petIdVal, err := strconv.ParseInt(petIdStr, 10, 64)
	if err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid petId parameter"))
		return
	}
	req.PetId = int64(petIdVal)

	// Call handler
	resp, err := w.Handler.GetPetById(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleUpdatePet adapts HTTP request to UpdatePet handler
func (w *ServerWrapper) handleUpdatePet(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := UpdatePetRequest{}

	// Parse path parameter: petId
	petIdStr := router.URLParam(r, "petId")
	petIdVal, err := strconv.ParseInt(petIdStr, 10, 64)
	if err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid petId parameter"))
		return
	}
	req.PetId = int64(petIdVal)

	// Parse request body
	if err := ReadJSON(r, &req.Body); err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid request body"))
		return
	}

	// Call handler
	resp, err := w.Handler.UpdatePet(ctx, req)
	if err != nil {
		w.handleError(rw, err)
		return
	}

	// Write response
	WriteResponse(rw, resp)
}

// handleDeletePet adapts HTTP request to DeletePet handler
func (w *ServerWrapper) handleDeletePet(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := DeletePetRequest{}

	// Parse path parameter: petId
	petIdStr := router.URLParam(r, "petId")
	petIdVal, err := strconv.ParseInt(petIdStr, 10, 64)
	if err != nil {
		w.handleError(rw, NewHTTPError(http.StatusBadRequest, "invalid petId parameter"))
		return
	}
	req.PetId = int64(petIdVal)

	// Call handler
	resp, err := w.Handler.DeletePet(ctx, req)
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
// Example with built-in router:
//
//	r := router.NewRouter()
//	ConfigureRouter(r, myServer)
//
// Example with custom router:
//
//	r := myCustomRouter.New() // Must implement router.Router interface
//	ConfigureRouter(r, myServer)
func ConfigureRouter(r router.Router, si Server) {
	wrapper := &ServerWrapper{Handler: si}

	r.Post("/pets", wrapper.handleCreatePet)
	r.Get("/pets", wrapper.handleListPets)
	r.Get("/pets/{petId}", wrapper.handleGetPetById)
	r.Put("/pets/{petId}", wrapper.handleUpdatePet)
	r.Delete("/pets/{petId}", wrapper.handleDeletePet)
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

