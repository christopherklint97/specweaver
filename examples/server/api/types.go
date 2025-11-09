package api

import (
	"time"
)

type Error struct {
	// Error code
	Code int `json:"code,omitempty"`
	// Error type
	Error string `json:"error"`
	// Detailed error message
	Message string `json:"message,omitempty"`
}

type NewPet struct {
	// Birth date of the pet
	BirthDate *time.Time `json:"birthDate,omitempty"`
	// Name of the pet
	Name string `json:"name"`
	Owner *Owner `json:"owner,omitempty"`
	Status PetStatus `json:"status"`
	// Tag to categorize the pet
	Tag string `json:"tag,omitempty"`
}

type Owner struct {
	// Email address of the owner
	Email string `json:"email,omitempty"`
	// Name of the owner
	Name string `json:"name"`
	// Phone number of the owner
	Phone string `json:"phone,omitempty"`
}

type Pet struct {
	// Birth date of the pet
	BirthDate *time.Time `json:"birthDate,omitempty"`
	// Unique identifier for the pet
	Id int64 `json:"id"`
	// Name of the pet
	Name string `json:"name"`
	Owner *Owner `json:"owner,omitempty"`
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

