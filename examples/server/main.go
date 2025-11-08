package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/christopherklint97/specweaver/examples/server/api"
	"github.com/go-chi/chi/v5"
)

// PetStoreServer implements the generated ServerInterface
type PetStoreServer struct {
	mu      sync.RWMutex
	pets    map[int64]api.Pet
	nextID  int64
}

// NewPetStoreServer creates a new pet store server instance
func NewPetStoreServer() *PetStoreServer {
	return &PetStoreServer{
		pets:   make(map[int64]api.Pet),
		nextID: 1,
	}
}

// ListPets implements the GET /pets endpoint
func (s *PetStoreServer) ListPets(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	tag := r.URL.Query().Get("tag")

	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Collect pets
	pets := make([]api.Pet, 0)
	count := 0
	for _, pet := range s.pets {
		// Filter by tag if provided
		if tag != "" && pet.Tag != tag {
			continue
		}

		pets = append(pets, pet)
		count++
		if count >= limit {
			break
		}
	}

	api.WriteJSON(w, http.StatusOK, pets)
}

// CreatePet implements the POST /pets endpoint
func (s *PetStoreServer) CreatePet(w http.ResponseWriter, r *http.Request) {
	var newPet api.NewPet
	if err := api.ReadJSON(r, &newPet); err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate required fields
	if newPet.Name == "" {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("name is required"))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new pet
	pet := api.Pet{
		Id:        s.nextID,
		Name:      newPet.Name,
		Tag:       newPet.Tag,
		Status:    newPet.Status,
		BirthDate: newPet.BirthDate,
		Owner:     newPet.Owner,
	}

	s.pets[s.nextID] = pet
	s.nextID++

	api.WriteJSON(w, http.StatusCreated, pet)
}

// GetPetById implements the GET /pets/{petId} endpoint
func (s *PetStoreServer) GetPetById(w http.ResponseWriter, r *http.Request) {
	petIDStr := chi.URLParam(r, "petId")
	petID, err := strconv.ParseInt(petIDStr, 10, 64)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid pet ID"))
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	pet, exists := s.pets[petID]
	if !exists {
		api.WriteError(w, http.StatusNotFound, fmt.Errorf("pet not found"))
		return
	}

	api.WriteJSON(w, http.StatusOK, pet)
}

// UpdatePet implements the PUT /pets/{petId} endpoint
func (s *PetStoreServer) UpdatePet(w http.ResponseWriter, r *http.Request) {
	petIDStr := chi.URLParam(r, "petId")
	petID, err := strconv.ParseInt(petIDStr, 10, 64)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid pet ID"))
		return
	}

	var updatePet api.NewPet
	if err := api.ReadJSON(r, &updatePet); err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pet, exists := s.pets[petID]
	if !exists {
		api.WriteError(w, http.StatusNotFound, fmt.Errorf("pet not found"))
		return
	}

	// Update fields
	pet.Name = updatePet.Name
	pet.Tag = updatePet.Tag
	pet.Status = updatePet.Status
	pet.BirthDate = updatePet.BirthDate
	pet.Owner = updatePet.Owner

	s.pets[petID] = pet

	api.WriteJSON(w, http.StatusOK, pet)
}

// DeletePet implements the DELETE /pets/{petId} endpoint
func (s *PetStoreServer) DeletePet(w http.ResponseWriter, r *http.Request) {
	petIDStr := chi.URLParam(r, "petId")
	petID, err := strconv.ParseInt(petIDStr, 10, 64)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid pet ID"))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pets[petID]; !exists {
		api.WriteError(w, http.StatusNotFound, fmt.Errorf("pet not found"))
		return
	}

	delete(s.pets, petID)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// Create server implementation
	server := NewPetStoreServer()

	// Seed with some sample data
	birthDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)
	server.pets[1] = api.Pet{
		Id:        1,
		Name:      "Fluffy",
		Tag:       "cat",
		Status:    api.PetStatusAvailable,
		BirthDate: &birthDate,
		Owner: &api.Owner{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "555-1234",
		},
	}
	server.nextID = 2

	// Create router with generated code
	router := api.NewRouter(server)

	// Start server
	port := ":8080"
	log.Printf("Starting Pet Store API server on http://localhost%s", port)
	log.Printf("Try: curl http://localhost%s/pets", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal(err)
	}
}
