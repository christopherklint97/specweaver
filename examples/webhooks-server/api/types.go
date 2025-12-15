package api

import (
	"time"
)

type Error struct {
	Error string `json:"error"`
	Message string `json:"message,omitempty"`
}

type Pet struct {
	// Unique pet identifier
	Id string `json:"id"`
	// Pet name
	Name string `json:"name"`
	// Species of the pet
	Species string `json:"species,omitempty"`
	// Pet status
	Status string `json:"status,omitempty"`
}

type PetStatusEvent struct {
	// New status
	NewStatus string `json:"newStatus"`
	// Previous status
	OldStatus string `json:"oldStatus"`
	// Pet identifier
	PetId string `json:"petId"`
	// When the status changed
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type Subscription struct {
	// Whether the subscription is active
	Active bool `json:"active,omitempty"`
	// Events to subscribe to
	Events []string `json:"events"`
	// Unique subscription identifier
	Id string `json:"id"`
	// Webhook destination URL
	Url string `json:"url"`
}

type WebhookAck struct {
	Message string `json:"message,omitempty"`
	Received bool `json:"received,omitempty"`
}

