package api

import (
	"time"
)

type Error struct {
	// Detailed error message
	Message string `json:"message,omitempty"`
	// Error code
	Code int `json:"code,omitempty"`
	// Error type
	Error string `json:"error"`
}

type NewPet struct {
	// Current status of the pet
	Status PetStatus `json:"status"`
	// Tag to categorize the pet
	Tag string `json:"tag,omitempty"`
	// Birth date of the pet
	BirthDate *time.Time `json:"birthDate,omitempty"`
	// Name of the pet
	Name string `json:"name"`
	Owner *Owner `json:"owner,omitempty"`
}

type Owner struct {
	// Phone number of the owner
	Phone string `json:"phone,omitempty"`
	// Email address of the owner
	Email string `json:"email,omitempty"`
	// Name of the owner
	Name string `json:"name"`
}

type Pet struct {
	// Birth date of the pet
	BirthDate *time.Time `json:"birthDate,omitempty"`
	// Unique identifier for the pet
	Id int64 `json:"id"`
	// Name of the pet
	Name string `json:"name"`
	Owner *Owner `json:"owner,omitempty"`
	// Current status of the pet
	Status PetStatus `json:"status"`
	// Tag to categorize the pet
	Tag string `json:"tag,omitempty"`
}

// PetStatus Current status of the pet
type PetStatus string

const (
	PetStatusAvailable PetStatus = "available"
	PetStatusPending PetStatus = "pending"
	PetStatusSold PetStatus = "sold"
)

