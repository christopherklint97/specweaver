package generator

import (
	"testing"

	"github.com/christopherklint97/specweaver/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookGenerator_NoWebhooks(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi.Paths{},
	}

	gen := NewWebhookGenerator(spec)
	code, err := gen.Generate()

	require.NoError(t, err)
	assert.Empty(t, code, "Should return empty string when no webhooks defined")
}

func TestWebhookGenerator_SimpleWebhook(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Webhooks: openapi.Webhooks{
			"newPet": &openapi.PathItem{
				Post: &openapi.Operation{
					OperationID: "onNewPet",
					Summary:     "Triggered when a new pet is created",
					RequestBody: &openapi.RequestBody{
						Required: true,
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &openapi.SchemaRef{
									Ref: "#/components/schemas/Pet",
								},
							},
						},
					},
					Responses: openapi.Responses{
						"200": &openapi.Response{
							Description: "Success",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: []string{"object"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"Pet": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"id": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
							"name": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
						},
					},
				},
			},
		},
	}

	gen := NewWebhookGenerator(spec)
	code, err := gen.Generate()

	require.NoError(t, err)
	assert.NotEmpty(t, code)

	// Verify generated code contains expected elements
	assert.Contains(t, code, "package api")
	assert.Contains(t, code, "OnNewPetRequest")
	assert.Contains(t, code, "OnNewPetResponse")
	assert.Contains(t, code, "WebhookClient interface")
	assert.Contains(t, code, "DefaultWebhookClient")
	assert.Contains(t, code, "NewWebhookClient()")
	assert.Contains(t, code, "func (c *DefaultWebhookClient) OnNewPet")
	assert.Contains(t, code, "URL string")
}

func TestWebhookGenerator_WebhookWithHeaders(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Webhooks: openapi.Webhooks{
			"statusChanged": &openapi.PathItem{
				Post: &openapi.Operation{
					OperationID: "onStatusChanged",
					Parameters: []*openapi.Parameter{
						{
							Name:     "X-Event-ID",
							In:       "header",
							Required: true,
							Schema: &openapi.SchemaRef{
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
						},
						{
							Name:     "X-Event-Time",
							In:       "header",
							Required: false,
							Schema: &openapi.SchemaRef{
								Value: &openapi.Schema{
									Type:   []string{"string"},
									Format: "date-time",
								},
							},
						},
					},
					RequestBody: &openapi.RequestBody{
						Required: true,
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &openapi.SchemaRef{
									Value: &openapi.Schema{
										Type: []string{"object"},
									},
								},
							},
						},
					},
					Responses: openapi.Responses{
						"200": {
							Description: "Success",
						},
					},
				},
			},
		},
	}

	gen := NewWebhookGenerator(spec)
	code, err := gen.Generate()

	require.NoError(t, err)
	assert.NotEmpty(t, code)

	// Verify headers are included in request type
	assert.Contains(t, code, "XEventID")
	assert.Contains(t, code, "XEventTime")
	assert.Contains(t, code, "httpReq.Header.Set(\"X-Event-ID\"")
	assert.Contains(t, code, "httpReq.Header.Set(\"X-Event-Time\"")
}

func TestWebhookGenerator_MultipleStatusCodes(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Webhooks: openapi.Webhooks{
			"notification": &openapi.PathItem{
				Post: &openapi.Operation{
					OperationID: "onNotification",
					RequestBody: &openapi.RequestBody{
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &openapi.SchemaRef{
									Value: &openapi.Schema{
										Type: []string{"object"},
									},
								},
							},
						},
					},
					Responses: openapi.Responses{
						"200": {
							Description: "Success",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: []string{"object"},
										},
									},
								},
							},
						},
						"400": {
							Description: "Bad request",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: []string{"object"},
										},
									},
								},
							},
						},
						"500": {
							Description: "Server error",
							Content: map[string]*openapi.MediaType{
								"application/json": {
									Schema: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: []string{"object"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	gen := NewWebhookGenerator(spec)
	code, err := gen.Generate()

	require.NoError(t, err)
	assert.NotEmpty(t, code)

	// Verify all response types are generated
	assert.Contains(t, code, "OnNotification200Response")
	assert.Contains(t, code, "OnNotification400Response")
	assert.Contains(t, code, "OnNotification500Response")

	// Verify switch case for all status codes
	assert.Contains(t, code, "case 200:")
	assert.Contains(t, code, "case 400:")
	assert.Contains(t, code, "case 500:")
}

func TestWebhookGenerator_MultipleWebhooks(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Webhooks: openapi.Webhooks{
			"event1": &openapi.PathItem{
				Post: &openapi.Operation{
					OperationID: "onEvent1",
					RequestBody: &openapi.RequestBody{
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &openapi.SchemaRef{
									Value: &openapi.Schema{
										Type: []string{"object"},
									},
								},
							},
						},
					},
					Responses: openapi.Responses{
						"200": {Description: "Success"},
					},
				},
			},
			"event2": &openapi.PathItem{
				Post: &openapi.Operation{
					OperationID: "onEvent2",
					RequestBody: &openapi.RequestBody{
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &openapi.SchemaRef{
									Value: &openapi.Schema{
										Type: []string{"object"},
									},
								},
							},
						},
					},
					Responses: openapi.Responses{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	gen := NewWebhookGenerator(spec)
	code, err := gen.Generate()

	require.NoError(t, err)
	assert.NotEmpty(t, code)

	// Verify both webhooks are in the interface
	assert.Contains(t, code, "OnEvent1(")
	assert.Contains(t, code, "OnEvent2(")

	// Verify both sender methods are generated
	assert.Contains(t, code, "func (c *DefaultWebhookClient) OnEvent1")
	assert.Contains(t, code, "func (c *DefaultWebhookClient) OnEvent2")
}
