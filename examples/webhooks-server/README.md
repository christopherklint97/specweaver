# Webhooks Example

This example demonstrates how to use SpecWeaver's webhook support to send outbound HTTP notifications to subscribers.

## Overview

Webhooks allow your API to notify external systems when certain events occur. This example shows:

1. **Server endpoints** for managing webhook subscriptions
2. **Webhook client** for sending notifications to subscribers
3. **Type-safe webhook interfaces** generated from OpenAPI webhooks

## Generated Code

From `webhooks-example.yaml`, SpecWeaver generates:

- **types.go**: Type definitions for subscriptions, pets, and webhook payloads
- **server.go**: Server interface and HTTP handlers for managing subscriptions
- **webhooks.go**: Webhook client interface and sender functions

### Webhook Client

The generated `WebhookClient` interface provides type-safe methods for sending webhooks:

```go
type WebhookClient interface {
    // OnNewPet Triggered when a new pet is created
    OnNewPet(ctx context.Context, req OnNewPetRequest) (OnNewPetResponse, error)

    // OnPetStatusChanged Triggered when a pet's status changes
    OnPetStatusChanged(ctx context.Context, req OnPetStatusChangedRequest) (OnPetStatusChangedResponse, error)
}
```

### Default HTTP Client

A default implementation `DefaultWebhookClient` is generated automatically:

```go
client := api.NewWebhookClient()

// Send a webhook
resp, err := client.OnNewPet(ctx, api.OnNewPetRequest{
    URL: "https://subscriber.example.com/webhook",
    Body: pet,
})
```

## Running the Example

1. **Generate the API code** (already done in this directory):
   ```bash
   specweaver -spec ../webhooks-example.yaml -output ./api
   ```

2. **Run the server**:
   ```bash
   go run main.go
   ```

3. **Subscribe to webhooks**:
   ```bash
   curl -X POST http://localhost:8080/subscriptions \
     -H "Content-Type: application/json" \
     -d '{
       "id": "sub-1",
       "url": "http://your-webhook-receiver.com/webhook",
       "events": ["newPet", "petStatusChanged"],
       "active": true
     }'
   ```

4. **Delete a subscription**:
   ```bash
   curl -X DELETE http://localhost:8080/subscriptions/sub-1
   ```

## Implementation Details

### Server Implementation

The `main.go` file demonstrates:

1. **Subscription Management**: Store and manage webhook subscriptions
2. **Event Triggering**: Call webhook client when events occur
3. **Asynchronous Delivery**: Send webhooks in background goroutines

### Webhook Request Structure

Each webhook has a request type that includes:

- **URL**: Destination URL for the webhook (required)
- **Headers**: Custom headers defined in the OpenAPI spec
- **Body**: The webhook payload (typed)

Example:
```go
type OnPetStatusChangedRequest struct {
    URL        string           `json:"url"`
    XEventID   string           `json:"X-Event-ID,omitempty"`
    XEventTime *string          `json:"X-Event-Time,omitempty"`
    Body       PetStatusEvent   `json:"body"`
}
```

### Webhook Response Handling

Responses are type-safe interfaces with concrete implementations for each status code:

```go
type OnNewPetResponse interface {
    isOnNewPetResponse()
    StatusCode() int
    ResponseBody() any
}

// Concrete implementations:
// - OnNewPet200Response: Success
// - OnNewPet500Response: Error
```

## Testing Webhooks

### Using a Webhook Testing Service

1. Use [webhook.site](https://webhook.site) to get a test URL
2. Subscribe using that URL:
   ```bash
   curl -X POST http://localhost:8080/subscriptions \
     -H "Content-Type: application/json" \
     -d '{
       "id": "test-1",
       "url": "https://webhook.site/your-unique-id",
       "events": ["newPet", "petStatusChanged"],
       "active": true
     }'
   ```
3. Watch for webhook deliveries on webhook.site

### Local Testing

Create a simple webhook receiver:

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    fmt.Printf("Received webhook: %s\n", string(body))

    // Return success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"received": true})
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    log.Println("Webhook receiver listening on :9000")
    http.ListenAndServe(":9000", nil)
}
```

Then subscribe to `http://localhost:9000/webhook`.

## Custom Webhook Client

You can implement your own webhook client with custom behavior:

```go
type CustomWebhookClient struct {
    httpClient *http.Client
    logger     *log.Logger
    retries    int
}

func (c *CustomWebhookClient) OnNewPet(ctx context.Context, req api.OnNewPetRequest) (api.OnNewPetResponse, error) {
    // Custom implementation with retries, logging, etc.
}
```

## Key Features

- **Type Safety**: All webhook payloads and responses are typed
- **Context Support**: All webhook senders accept `context.Context`
- **Custom Headers**: Support for required and optional headers
- **Error Handling**: Proper error types and status code handling
- **Customizable**: Replace default client with your own implementation
- **Async Delivery**: Easy to send webhooks asynchronously

## OpenAPI Webhooks Specification

The webhooks are defined in the OpenAPI spec under the `webhooks` section:

```yaml
webhooks:
  newPet:
    post:
      operationId: onNewPet
      summary: Triggered when a new pet is created
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Pet'
      responses:
        '200':
          description: Webhook received successfully
```

SpecWeaver generates client code to send these webhooks to subscriber URLs.
