package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/christopherklint97/specweaver/examples/server/api"
)

// PetStoreServer implements the generated Server interface
type PetStoreServer struct {
	mu     sync.RWMutex
	pets   map[int64]api.Pet
	nextID int64
}

// NewPetStoreServer creates a new pet store server instance
func NewPetStoreServer() *PetStoreServer {
	return &PetStoreServer{
		pets:   make(map[int64]api.Pet),
		nextID: 1,
	}
}

// ListPets implements the ListPets handler
func (s *PetStoreServer) ListPets(ctx context.Context, req api.ListPetsRequest) (api.ListPetsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Default limit
	limit := int32(20)
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Collect pets
	pets := make([]api.Pet, 0)
	count := int32(0)
	for _, pet := range s.pets {
		// Filter by tag if provided
		if req.Tag != nil && pet.Tag != *req.Tag {
			continue
		}

		pets = append(pets, pet)
		count++
		if count >= limit {
			break
		}
	}

	return api.ListPets200Response{Body: pets}, nil
}

// CreatePet implements the CreatePet handler
func (s *PetStoreServer) CreatePet(ctx context.Context, req api.CreatePetRequest) (api.CreatePetResponse, error) {
	// Validate required fields
	if req.Body.Name == "" {
		return nil, api.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new pet
	pet := api.Pet{
		Id:        s.nextID,
		Name:      req.Body.Name,
		Tag:       req.Body.Tag,
		Status:    req.Body.Status,
		BirthDate: req.Body.BirthDate,
		Owner:     req.Body.Owner,
	}

	s.pets[s.nextID] = pet
	s.nextID++

	return api.CreatePet201Response{Body: pet}, nil
}

// GetPetById implements the GetPetById handler
func (s *PetStoreServer) GetPetById(ctx context.Context, req api.GetPetByIdRequest) (api.GetPetByIdResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pet, exists := s.pets[req.PetId]
	if !exists {
		return api.GetPetById404Response{
			Body: api.Error{
				Error:   "Not Found",
				Message: "pet not found",
			},
		}, nil
	}

	return api.GetPetById200Response{Body: pet}, nil
}

// UpdatePet implements the UpdatePet handler
func (s *PetStoreServer) UpdatePet(ctx context.Context, req api.UpdatePetRequest) (api.UpdatePetResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pet, exists := s.pets[req.PetId]
	if !exists {
		return api.UpdatePet404Response{
			Body: api.Error{
				Error:   "Not Found",
				Message: "pet not found",
			},
		}, nil
	}

	// Update fields
	pet.Name = req.Body.Name
	pet.Tag = req.Body.Tag
	pet.Status = req.Body.Status
	pet.BirthDate = req.Body.BirthDate
	pet.Owner = req.Body.Owner

	s.pets[req.PetId] = pet

	return api.UpdatePet200Response{Body: pet}, nil
}

// DeletePet implements the DeletePet handler
func (s *PetStoreServer) DeletePet(ctx context.Context, req api.DeletePetRequest) (api.DeletePetResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pets[req.PetId]; !exists {
		return api.DeletePet404Response{
			Body: api.Error{
				Error:   "Not Found",
				Message: "pet not found",
			},
		}, nil
	}

	delete(s.pets, req.PetId)
	return api.DeletePet204Response{}, nil
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
