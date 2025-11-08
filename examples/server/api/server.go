package api

import (
	"encoding/json"
	"net/http"
	"io"

	"github.com/christopherklint97/specweaver/pkg/router"
)

// ServerInterface represents all server handlers
type ServerInterface interface {
	// ListPets List all pets
	ListPets(w http.ResponseWriter, r *http.Request)
	// CreatePet Create a pet
	CreatePet(w http.ResponseWriter, r *http.Request)
	// UpdatePet Update a pet
	UpdatePet(w http.ResponseWriter, r *http.Request)
	// DeletePet Delete a pet
	DeletePet(w http.ResponseWriter, r *http.Request)
	// GetPetById Get a pet by ID
	GetPetById(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper wraps the ServerInterface with HTTP handler logic
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// NewRouter creates a new router with all routes configured
func NewRouter(si ServerInterface) *router.Mux {
	r := router.NewRouter()

	// Middleware
	r.Use(router.Logger)
	r.Use(router.Recoverer)
	r.Use(router.RequestID)
	r.Use(router.RealIP)

	wrapper := &ServerInterfaceWrapper{Handler: si}

	r.Get("/pets", wrapper.Handler.ListPets)
	r.Post("/pets", wrapper.Handler.CreatePet)
	r.Get("/pets/{petId}", wrapper.Handler.GetPetById)
	r.Put("/pets/{petId}", wrapper.Handler.UpdatePet)
	r.Delete("/pets/{petId}", wrapper.Handler.DeletePet)

	return r
}

// Helper functions for request/response handling

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
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

