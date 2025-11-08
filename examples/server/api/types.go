package api

import (
	"time"
)

type Owner struct {
	// Name of the owner
	Name string `json:"name"`
	// Phone number of the owner
	Phone string `json:"phone,omitempty"`
	// Email address of the owner
	Email string `json:"email,omitempty"`
}

// PetStatus Current status of the pet
type PetStatus string

const (
	PetStatusAvailable PetStatus = "available"
	PetStatusPending PetStatus = "pending"
	PetStatusSold PetStatus = "sold"
)

type Error struct {
	// Error type
	Error string `json:"error"`
	// Detailed error message
	Message string `json:"message,omitempty"`
	// Error code
	Code int `json:"code,omitempty"`
}

type Pet struct {
	// Name of the pet
	Name string `json:"name"`
	Owner *Owner `json:"owner,omitempty"`
	Status PetStatus `json:"status"`
	// Tag to categorize the pet
	Tag string `json:"tag,omitempty"`
	// Birth date of the pet
	BirthDate *time.Time `json:"birthDate,omitempty"`
	// Unique identifier for the pet
	Id int64 `json:"id"`
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

