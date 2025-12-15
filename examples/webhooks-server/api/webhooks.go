package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OnNewPetRequest represents the request for newPet webhook
type OnNewPetRequest struct {
	// URL is the webhook destination URL
	URL string `json:"url"`
	// Request body
	Body Pet `json:"body"`
}

// OnPetStatusChangedRequest represents the request for petStatusChanged webhook
type OnPetStatusChangedRequest struct {
	// URL is the webhook destination URL
	URL string `json:"url"`
	// Unique event identifier
	XEventID string `json:"X-Event-ID,omitempty"`
	// Event timestamp
	XEventTime *string `json:"X-Event-Time,omitempty"`
	// Request body
	Body PetStatusEvent `json:"body"`
}

// OnNewPetResponse represents possible responses for newPet webhook
type OnNewPetResponse interface {
	isOnNewPetResponse()
	StatusCode() int
	ResponseBody() any
}

// OnNewPet200Response represents a 200 response from webhook
type OnNewPet200Response struct {
	Body WebhookAck `json:"body"`
}

func (r OnNewPet200Response) isOnNewPetResponse() {}
func (r OnNewPet200Response) StatusCode() int { return 200 }
func (r OnNewPet200Response) ResponseBody() any { return r.Body }

// OnNewPet500Response represents a 500 response from webhook
type OnNewPet500Response struct {
	Body Error `json:"body"`
}

func (r OnNewPet500Response) isOnNewPetResponse() {}
func (r OnNewPet500Response) StatusCode() int { return 500 }
func (r OnNewPet500Response) ResponseBody() any { return r.Body }

// OnPetStatusChangedResponse represents possible responses for petStatusChanged webhook
type OnPetStatusChangedResponse interface {
	isOnPetStatusChangedResponse()
	StatusCode() int
	ResponseBody() any
}

// OnPetStatusChanged200Response represents a 200 response from webhook
type OnPetStatusChanged200Response struct {
	Body WebhookAck `json:"body"`
}

func (r OnPetStatusChanged200Response) isOnPetStatusChangedResponse() {}
func (r OnPetStatusChanged200Response) StatusCode() int { return 200 }
func (r OnPetStatusChanged200Response) ResponseBody() any { return r.Body }

// OnPetStatusChanged400Response represents a 400 response from webhook
type OnPetStatusChanged400Response struct {
	Body Error `json:"body"`
}

func (r OnPetStatusChanged400Response) isOnPetStatusChangedResponse() {}
func (r OnPetStatusChanged400Response) StatusCode() int { return 400 }
func (r OnPetStatusChanged400Response) ResponseBody() any { return r.Body }

// WebhookClient represents all webhook senders
type WebhookClient interface {
	// OnNewPet Triggered when a new pet is created
	OnNewPet(ctx context.Context, req OnNewPetRequest) (OnNewPetResponse, error)
	// OnPetStatusChanged Triggered when a pet's status changes
	OnPetStatusChanged(ctx context.Context, req OnPetStatusChangedRequest) (OnPetStatusChangedResponse, error)
}

// DefaultWebhookClient is the default HTTP implementation of WebhookClient
type DefaultWebhookClient struct {
	HTTPClient *http.Client
}

// NewWebhookClient creates a new default webhook client
func NewWebhookClient() *DefaultWebhookClient {
	return &DefaultWebhookClient{
		HTTPClient: http.DefaultClient,
	}
}

// OnNewPet sends the newPet webhook
func (c *DefaultWebhookClient) OnNewPet(ctx context.Context, req OnNewPetRequest) (OnNewPetResponse, error) {
	// Serialize request body
	body, err := json.Marshal(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", req.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Parse response based on status code
	switch resp.StatusCode {
	case 200:
		var result OnNewPet200Response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if err := json.Unmarshal(body, &result.Body); err != nil {
			return nil, fmt.Errorf("failed to parse response body: %w", err)
		}
		return result, nil
	case 500:
		var result OnNewPet500Response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if err := json.Unmarshal(body, &result.Body); err != nil {
			return nil, fmt.Errorf("failed to parse response body: %w", err)
		}
		return result, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

// OnPetStatusChanged sends the petStatusChanged webhook
func (c *DefaultWebhookClient) OnPetStatusChanged(ctx context.Context, req OnPetStatusChangedRequest) (OnPetStatusChangedResponse, error) {
	// Serialize request body
	body, err := json.Marshal(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", req.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type
	httpReq.Header.Set("Content-Type", "application/json")

	// Set custom headers
	httpReq.Header.Set("X-Event-ID", fmt.Sprintf("%v", req.XEventID))
	if req.XEventTime != nil {
		httpReq.Header.Set("X-Event-Time", fmt.Sprintf("%v", *req.XEventTime))
	}

	// Send request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Parse response based on status code
	switch resp.StatusCode {
	case 200:
		var result OnPetStatusChanged200Response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if err := json.Unmarshal(body, &result.Body); err != nil {
			return nil, fmt.Errorf("failed to parse response body: %w", err)
		}
		return result, nil
	case 400:
		var result OnPetStatusChanged400Response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if err := json.Unmarshal(body, &result.Body); err != nil {
			return nil, fmt.Errorf("failed to parse response body: %w", err)
		}
		return result, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

