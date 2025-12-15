package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// WebhookGenerator generates Go webhook client code from OpenAPI webhooks
type WebhookGenerator struct {
	spec *openapi.Document
}

// NewWebhookGenerator creates a new WebhookGenerator instance
func NewWebhookGenerator(spec *openapi.Document) *WebhookGenerator {
	return &WebhookGenerator{
		spec: spec,
	}
}

// Generate generates webhook client code
func (g *WebhookGenerator) Generate() (string, error) {
	// If no webhooks, return empty
	if len(g.spec.Webhooks) == 0 {
		return "", nil
	}

	// Build template data
	data := WebhooksTemplateData{
		RequestTypes:  g.buildRequestTypes(),
		ResponseTypes: g.buildResponseTypes(),
		Operations:    g.buildOperations(),
	}

	// Execute template
	return executeTemplate("webhooks.go.tmpl", data)
}

// buildRequestTypes builds webhook request type data
func (g *WebhookGenerator) buildRequestTypes() []WebhookRequestTypeData {
	var result []WebhookRequestTypeData

	webhookNames := g.getSortedWebhookNames()

	for _, webhookName := range webhookNames {
		pathItem := g.spec.Webhooks[webhookName]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateWebhookName(webhookName, method, op.OperationID)
			requestTypeName := handlerName + "Request"

			reqType := WebhookRequestTypeData{
				Name:        requestTypeName,
				WebhookName: webhookName,
				Fields:      []FieldDefinition{},
			}

			// Add header parameters
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param != nil && param.In == "header" {
						fieldName := toPascalCase(param.Name)
						fieldType := getParamType(param)

						// Headers are optional by default
						if !param.Required && !strings.HasPrefix(fieldType, "*") {
							fieldType = "*" + fieldType
						}

						field := FieldDefinition{
							Name:        fieldName,
							Type:        fieldType,
							JSONTag:     param.Name + ",omitempty",
							Description: param.Description,
						}
						reqType.Fields = append(reqType.Fields, field)
					}
				}
			}

			// Add request body if present
			if op.RequestBody != nil {
				content := op.RequestBody.Content
				if jsonContent, ok := content["application/json"]; ok && jsonContent.Schema != nil {
					bodyType := resolveSchemaType(jsonContent.Schema)
					field := FieldDefinition{
						Name:        "Body",
						Type:        bodyType,
						JSONTag:     "body",
						Description: "Request body",
					}
					reqType.Fields = append(reqType.Fields, field)
				}
			}

			result = append(result, reqType)
		}
	}

	return result
}

// buildResponseTypes builds webhook response type data
func (g *WebhookGenerator) buildResponseTypes() []WebhookResponseTypeData {
	var result []WebhookResponseTypeData

	webhookNames := g.getSortedWebhookNames()

	for _, webhookName := range webhookNames {
		pathItem := g.spec.Webhooks[webhookName]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateWebhookName(webhookName, method, op.OperationID)
			responseTypeName := handlerName + "Response"

			respType := WebhookResponseTypeData{
				InterfaceName: responseTypeName,
				WebhookName:   webhookName,
				ConcreteTypes: []ConcreteResponseType{},
			}

			// Generate concrete response types for each status code
			if op.Responses != nil {
				statusCodes := make([]string, 0, len(op.Responses))
				for statusCode := range op.Responses {
					statusCodes = append(statusCodes, statusCode)
				}
				sort.Strings(statusCodes)

				for _, statusCode := range statusCodes {
					response := op.Responses[statusCode]
					if response == nil || statusCode == "default" {
						continue
					}

					statusCodeInt := parseStatusCode(statusCode)
					if statusCodeInt == 0 {
						continue
					}

					concreteType := ConcreteResponseType{
						Name:       fmt.Sprintf("%s%dResponse", handlerName, statusCodeInt),
						StatusCode: statusCodeInt,
						HasBody:    false,
					}

					// Check if response has content
					if response.Content != nil {
						if jsonContent, ok := response.Content["application/json"]; ok && jsonContent.Schema != nil {
							concreteType.HasBody = true
							concreteType.BodyType = resolveSchemaType(jsonContent.Schema)
						}
					}

					respType.ConcreteTypes = append(respType.ConcreteTypes, concreteType)
				}
			}

			result = append(result, respType)
		}
	}

	return result
}

// buildOperations builds webhook operation data
func (g *WebhookGenerator) buildOperations() []WebhookOperationData {
	var result []WebhookOperationData

	webhookNames := g.getSortedWebhookNames()

	for _, webhookName := range webhookNames {
		pathItem := g.spec.Webhooks[webhookName]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateWebhookName(webhookName, method, op.OperationID)

			opData := WebhookOperationData{
				HandlerName:      handlerName,
				WebhookName:      webhookName,
				Summary:          op.Summary,
				RequestTypeName:  handlerName + "Request",
				ResponseTypeName: handlerName + "Response",
				Method:           strings.ToUpper(method),
				HasRequestBody:   false,
				HasHeaders:       false,
				Headers:          []HeaderData{},
				Responses:        []WebhookResponseData{},
			}

			// Check for request body
			if op.RequestBody != nil {
				content := op.RequestBody.Content
				if _, ok := content["application/json"]; ok {
					opData.HasRequestBody = true
				}
			}

			// Build headers
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param != nil && param.In == "header" {
						opData.HasHeaders = true
						opData.Headers = append(opData.Headers, HeaderData{
							HeaderName: param.Name,
							FieldName:  toPascalCase(param.Name),
							Required:   param.Required,
						})
					}
				}
			}

			// Build responses
			if op.Responses != nil {
				statusCodes := make([]string, 0, len(op.Responses))
				for statusCode := range op.Responses {
					if statusCode != "default" {
						statusCodes = append(statusCodes, statusCode)
					}
				}
				sort.Strings(statusCodes)

				for _, statusCode := range statusCodes {
					response := op.Responses[statusCode]
					if response == nil {
						continue
					}

					statusCodeInt := parseStatusCode(statusCode)
					if statusCodeInt == 0 {
						continue
					}

					respData := WebhookResponseData{
						StatusCode: statusCodeInt,
						TypeName:   fmt.Sprintf("%s%dResponse", handlerName, statusCodeInt),
						HasBody:    false,
					}

					// Check if response has content
					if response.Content != nil {
						if jsonContent, ok := response.Content["application/json"]; ok && jsonContent.Schema != nil {
							respData.HasBody = true
						}
					}

					opData.Responses = append(opData.Responses, respData)
				}
			}

			result = append(result, opData)
		}
	}

	return result
}

// getSortedWebhookNames returns webhook names in sorted order
func (g *WebhookGenerator) getSortedWebhookNames() []string {
	webhookNames := make([]string, 0, len(g.spec.Webhooks))
	for name := range g.spec.Webhooks {
		webhookNames = append(webhookNames, name)
	}
	sort.Strings(webhookNames)
	return webhookNames
}

// generateWebhookName creates a webhook function name
func generateWebhookName(webhookName, method, operationID string) string {
	if operationID != "" {
		return toPascalCase(operationID)
	}

	// Generate from webhook name and method
	name := "send_" + webhookName + "_" + strings.ToLower(method)
	return toPascalCase(name)
}
