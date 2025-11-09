package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/christopherklint97/specweaver/examples/webhooks-server/api"
)

// Server implements the generated api.Server interface
type Server struct {
	subscriptions map[string]*api.Subscription
	mu            sync.RWMutex
	webhookClient api.WebhookClient
}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{
		subscriptions: make(map[string]*api.Subscription),
		webhookClient: api.NewWebhookClient(),
	}
}

// CreateSubscription implements the subscription creation endpoint
func (s *Server) CreateSubscription(ctx context.Context, req api.CreateSubscriptionRequest) (api.CreateSubscriptionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate the subscription
	if req.Body.Url == "" {
		return nil, api.NewHTTPError(http.StatusBadRequest, "url is required")
	}

	if len(req.Body.Events) == 0 {
		return nil, api.NewHTTPError(http.StatusBadRequest, "at least one event is required")
	}

	// Store the subscription
	s.subscriptions[req.Body.Id] = &req.Body

	log.Printf("Created subscription %s for URL %s (events: %v)", req.Body.Id, req.Body.Url, req.Body.Events)

	return api.CreateSubscription201Response{Body: req.Body}, nil
}

// DeleteSubscription implements the subscription deletion endpoint
func (s *Server) DeleteSubscription(ctx context.Context, req api.DeleteSubscriptionRequest) (api.DeleteSubscriptionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subscriptions[req.Id]; !exists {
		return api.DeleteSubscription404Response{
			Body: api.Error{
				Error:   "Not Found",
				Message: "subscription not found",
			},
		}, nil
	}

	delete(s.subscriptions, req.Id)
	log.Printf("Deleted subscription %s", req.Id)

	return api.DeleteSubscription204Response{}, nil
}

// notifyNewPet sends the newPet webhook to all subscribers
func (s *Server) notifyNewPet(ctx context.Context, pet api.Pet) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, sub := range s.subscriptions {
		if !sub.Active {
			continue
		}

		// Check if this subscription includes the newPet event
		hasEvent := false
		for _, event := range sub.Events {
			if event == "newPet" {
				hasEvent = true
				break
			}
		}

		if !hasEvent {
			continue
		}

		// Send webhook asynchronously
		go func(subscriptionID string, url string) {
			resp, err := s.webhookClient.OnNewPet(ctx, api.OnNewPetRequest{
				URL:  url,
				Body: pet,
			})

			if err != nil {
				log.Printf("Failed to send newPet webhook to subscription %s: %v", subscriptionID, err)
				return
			}

			log.Printf("Sent newPet webhook to subscription %s (status: %d)", subscriptionID, resp.StatusCode())
		}(id, sub.Url)
	}
}

// notifyPetStatusChanged sends the petStatusChanged webhook to all subscribers
func (s *Server) notifyPetStatusChanged(ctx context.Context, event api.PetStatusEvent, eventID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, sub := range s.subscriptions {
		if !sub.Active {
			continue
		}

		// Check if this subscription includes the petStatusChanged event
		hasEvent := false
		for _, evt := range sub.Events {
			if evt == "petStatusChanged" {
				hasEvent = true
				break
			}
		}

		if !hasEvent {
			continue
		}

		// Send webhook asynchronously
		go func(subscriptionID string, url string) {
			resp, err := s.webhookClient.OnPetStatusChanged(ctx, api.OnPetStatusChangedRequest{
				URL:      url,
				XEventID: eventID,
				Body:     event,
			})

			if err != nil {
				log.Printf("Failed to send petStatusChanged webhook to subscription %s: %v", subscriptionID, err)
				return
			}

			log.Printf("Sent petStatusChanged webhook to subscription %s (status: %d)", subscriptionID, resp.StatusCode())
		}(id, sub.Url)
	}
}

func main() {
	// Create server
	server := NewServer()

	// Create router with generated routes
	router := api.NewRouter(server)

	// Demo: Simulate a new pet being added
	go func() {
		// Wait a bit for the server to start
		// In a real application, you would trigger this based on actual events
		ctx := context.Background()

		// Example: notify about a new pet
		pet := api.Pet{
			Id:      "pet-123",
			Name:    "Buddy",
			Status:  "available",
			Species: "dog",
		}

		log.Println("Demo: Notifying subscribers about new pet")
		server.notifyNewPet(ctx, pet)
	}()

	// Start server
	addr := ":8080"
	log.Printf("Server starting on %s", addr)
	log.Printf("Example endpoints:")
	log.Printf("  POST   http://localhost%s/subscriptions", addr)
	log.Printf("  DELETE http://localhost%s/subscriptions/{id}", addr)
	log.Printf("")
	log.Printf("Webhooks:")
	log.Printf("  - newPet: Triggered when a new pet is created")
	log.Printf("  - petStatusChanged: Triggered when a pet's status changes")
	log.Printf("")
	log.Printf("Try subscribing to webhooks:")
	log.Printf(`  curl -X POST http://localhost%s/subscriptions \`, addr)
	log.Printf(`    -H "Content-Type: application/json" \`)
	log.Printf(`    -d '{`)
	log.Printf(`      "id": "sub-1",`)
	log.Printf(`      "url": "http://your-webhook-receiver.com/webhook",`)
	log.Printf(`      "events": ["newPet", "petStatusChanged"],`)
	log.Printf(`      "active": true`)
	log.Printf(`    }'`)
	log.Printf("")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
