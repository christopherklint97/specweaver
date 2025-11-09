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
	if g.spec.Webhooks == nil || len(g.spec.Webhooks) == 0 {
		return "", nil
	}

	var sb strings.Builder

	sb.WriteString("package api\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"bytes\"\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"encoding/json\"\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"io\"\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString(")\n\n")

	// Generate request types for each webhook
	if err := g.generateWebhookRequestTypes(&sb); err != nil {
		return "", err
	}

	// Generate response types for each webhook
	if err := g.generateWebhookResponseTypes(&sb); err != nil {
		return "", err
	}

	// Generate the webhook client interface
	if err := g.generateWebhookClientInterface(&sb); err != nil {
		return "", err
	}

	// Generate the default HTTP webhook client
	g.generateDefaultWebhookClient(&sb)

	// Generate webhook sender methods
	if err := g.generateWebhookSenders(&sb); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// generateWebhookRequestTypes generates request structs for each webhook
func (g *WebhookGenerator) generateWebhookRequestTypes(sb *strings.Builder) error {
	// Sort webhook names for deterministic output
	webhookNames := make([]string, 0, len(g.spec.Webhooks))
	for name := range g.spec.Webhooks {
		webhookNames = append(webhookNames, name)
	}
	sort.Strings(webhookNames)

	for _, webhookName := range webhookNames {
		pathItem := g.spec.Webhooks[webhookName]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateWebhookName(webhookName, method, op.OperationID)
			requestTypeName := handlerName + "Request"

			sb.WriteString(fmt.Sprintf("// %s represents the request for %s webhook\n", requestTypeName, webhookName))
			sb.WriteString(fmt.Sprintf("type %s struct {\n", requestTypeName))

			// Webhooks always need a destination URL
			sb.WriteString("\t// URL is the webhook destination URL\n")
			sb.WriteString("\tURL string `json:\"url\"`\n")

			// Add headers if needed
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param == nil {
						continue
					}

					if param.In == "header" {
						fieldName := toPascalCase(param.Name)
						fieldType := g.getParamType(param)

						// Headers are optional by default
						if !param.Required && !strings.HasPrefix(fieldType, "*") {
							fieldType = "*" + fieldType
						}

						jsonTag := param.Name + ",omitempty"
						if param.Description != "" {
							sb.WriteString(fmt.Sprintf("\t// %s\n", param.Description))
						}
						sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, fieldType, jsonTag))
					}
				}
			}

			// Add request body if present
			if op.RequestBody != nil {
				content := op.RequestBody.Content
				if jsonContent, ok := content["application/json"]; ok && jsonContent.Schema != nil {
					bodyType := g.resolveSchemaType(jsonContent.Schema)
					sb.WriteString("\t// Request body\n")
					sb.WriteString(fmt.Sprintf("\tBody %s `json:\"body\"`\n", bodyType))
				}
			}

			sb.WriteString("}\n\n")
		}
	}

	return nil
}

// generateWebhookResponseTypes generates response types for each webhook
func (g *WebhookGenerator) generateWebhookResponseTypes(sb *strings.Builder) error {
	// Sort webhook names for deterministic output
	webhookNames := make([]string, 0, len(g.spec.Webhooks))
	for name := range g.spec.Webhooks {
		webhookNames = append(webhookNames, name)
	}
	sort.Strings(webhookNames)

	for _, webhookName := range webhookNames {
		pathItem := g.spec.Webhooks[webhookName]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateWebhookName(webhookName, method, op.OperationID)
			responseTypeName := handlerName + "Response"

			// Generate response interface
			sb.WriteString(fmt.Sprintf("// %s represents possible responses for %s webhook\n", responseTypeName, webhookName))
			sb.WriteString(fmt.Sprintf("type %s interface {\n", responseTypeName))
			sb.WriteString(fmt.Sprintf("\tis%s()\n", responseTypeName))
			sb.WriteString("\tStatusCode() int\n")
			sb.WriteString("\tResponseBody() any\n")
			sb.WriteString("}\n\n")

			// Generate concrete response types for each status code (in sorted order)
			if op.Responses != nil {
				statusCodes := make([]string, 0, len(op.Responses))
				for statusCode := range op.Responses {
					statusCodes = append(statusCodes, statusCode)
				}
				sort.Strings(statusCodes)

				for _, statusCode := range statusCodes {
					response := op.Responses[statusCode]
					if response == nil {
						continue
					}

					// Skip "default" responses
					if statusCode == "default" {
						continue
					}

					// Parse status code
					statusCodeInt := parseStatusCode(statusCode)
					if statusCodeInt == 0 {
						continue
					}
					concreteTypeName := fmt.Sprintf("%s%dResponse", handlerName, statusCodeInt)

					sb.WriteString(fmt.Sprintf("// %s represents a %d response from webhook\n", concreteTypeName, statusCodeInt))
					sb.WriteString(fmt.Sprintf("type %s struct {\n", concreteTypeName))

					// Check if response has content
					hasBody := false
					if response.Content != nil {
						if jsonContent, ok := response.Content["application/json"]; ok && jsonContent.Schema != nil {
							bodyType := g.resolveSchemaType(jsonContent.Schema)
							sb.WriteString(fmt.Sprintf("\tBody %s `json:\"body\"`\n", bodyType))
							hasBody = true
						}
					}

					sb.WriteString("}\n\n")

					// Generate interface implementation methods
					sb.WriteString(fmt.Sprintf("func (r %s) is%s() {}\n", concreteTypeName, responseTypeName))
					sb.WriteString(fmt.Sprintf("func (r %s) StatusCode() int { return %d }\n", concreteTypeName, statusCodeInt))

					// Generate ResponseBody method
					if hasBody {
						sb.WriteString(fmt.Sprintf("func (r %s) ResponseBody() any { return r.Body }\n\n", concreteTypeName))
					} else {
						sb.WriteString(fmt.Sprintf("func (r %s) ResponseBody() any { return nil }\n\n", concreteTypeName))
					}
				}
			}
		}
	}

	return nil
}

// generateWebhookClientInterface generates the interface for webhook clients
func (g *WebhookGenerator) generateWebhookClientInterface(sb *strings.Builder) error {
	sb.WriteString("// WebhookClient represents all webhook senders\n")
	sb.WriteString("type WebhookClient interface {\n")

	// Sort webhook names for deterministic output
	webhookNames := make([]string, 0, len(g.spec.Webhooks))
	for name := range g.spec.Webhooks {
		webhookNames = append(webhookNames, name)
	}
	sort.Strings(webhookNames)

	for _, webhookName := range webhookNames {
		pathItem := g.spec.Webhooks[webhookName]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateWebhookName(webhookName, method, op.OperationID)
			requestTypeName := handlerName + "Request"
			responseTypeName := handlerName + "Response"

			// Add comment with operation summary
			if op.Summary != "" {
				sb.WriteString(fmt.Sprintf("\t// %s %s\n", handlerName, op.Summary))
			} else {
				sb.WriteString(fmt.Sprintf("\t// %s sends the %s webhook\n", handlerName, webhookName))
			}

			sb.WriteString(fmt.Sprintf("\t%s(ctx context.Context, req %s) (%s, error)\n", handlerName, requestTypeName, responseTypeName))
		}
	}

	sb.WriteString("}\n\n")
	return nil
}

// generateDefaultWebhookClient generates the default HTTP webhook client implementation
func (g *WebhookGenerator) generateDefaultWebhookClient(sb *strings.Builder) {
	sb.WriteString("// DefaultWebhookClient is the default HTTP implementation of WebhookClient\n")
	sb.WriteString("type DefaultWebhookClient struct {\n")
	sb.WriteString("\tHTTPClient *http.Client\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// NewWebhookClient creates a new default webhook client\n")
	sb.WriteString("func NewWebhookClient() *DefaultWebhookClient {\n")
	sb.WriteString("\treturn &DefaultWebhookClient{\n")
	sb.WriteString("\t\tHTTPClient: http.DefaultClient,\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")
}

// generateWebhookSenders generates the webhook sender methods
func (g *WebhookGenerator) generateWebhookSenders(sb *strings.Builder) error {
	// Sort webhook names for deterministic output
	webhookNames := make([]string, 0, len(g.spec.Webhooks))
	for name := range g.spec.Webhooks {
		webhookNames = append(webhookNames, name)
	}
	sort.Strings(webhookNames)

	for _, webhookName := range webhookNames {
		pathItem := g.spec.Webhooks[webhookName]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateWebhookName(webhookName, method, op.OperationID)
			g.generateWebhookSenderMethod(sb, handlerName, webhookName, method, op)
		}
	}

	return nil
}

// generateWebhookSenderMethod generates a single webhook sender method
func (g *WebhookGenerator) generateWebhookSenderMethod(sb *strings.Builder, handlerName, webhookName, method string, op *openapi.Operation) {
	requestTypeName := handlerName + "Request"
	responseTypeName := handlerName + "Response"

	sb.WriteString(fmt.Sprintf("// %s sends the %s webhook\n", handlerName, webhookName))
	sb.WriteString(fmt.Sprintf("func (c *DefaultWebhookClient) %s(ctx context.Context, req %s) (%s, error) {\n", handlerName, requestTypeName, responseTypeName))

	// Prepare request body
	if op.RequestBody != nil {
		content := op.RequestBody.Content
		if _, ok := content["application/json"]; ok {
			sb.WriteString("\t// Serialize request body\n")
			sb.WriteString("\tbody, err := json.Marshal(req.Body)\n")
			sb.WriteString("\tif err != nil {\n")
			sb.WriteString("\t\treturn nil, fmt.Errorf(\"failed to marshal request body: %w\", err)\n")
			sb.WriteString("\t}\n\n")
		}
	}

	// Create HTTP request
	sb.WriteString("\t// Create HTTP request\n")
	if op.RequestBody != nil {
		sb.WriteString(fmt.Sprintf("\thttpReq, err := http.NewRequestWithContext(ctx, %q, req.URL, bytes.NewReader(body))\n", strings.ToUpper(method)))
	} else {
		sb.WriteString(fmt.Sprintf("\thttpReq, err := http.NewRequestWithContext(ctx, %q, req.URL, nil)\n", strings.ToUpper(method)))
	}
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, fmt.Errorf(\"failed to create request: %w\", err)\n")
	sb.WriteString("\t}\n\n")

	// Set content type if needed
	if op.RequestBody != nil {
		sb.WriteString("\t// Set content type\n")
		sb.WriteString("\thttpReq.Header.Set(\"Content-Type\", \"application/json\")\n\n")
	}

	// Set custom headers
	if op.Parameters != nil {
		hasHeaders := false
		for _, param := range op.Parameters {
			if param != nil && param.In == "header" {
				hasHeaders = true
				break
			}
		}

		if hasHeaders {
			sb.WriteString("\t// Set custom headers\n")
			for _, param := range op.Parameters {
				if param == nil || param.In != "header" {
					continue
				}

				fieldName := toPascalCase(param.Name)
				if param.Required {
					sb.WriteString(fmt.Sprintf("\thttpReq.Header.Set(\"%s\", fmt.Sprintf(\"%%v\", req.%s))\n", param.Name, fieldName))
				} else {
					sb.WriteString(fmt.Sprintf("\tif req.%s != nil {\n", fieldName))
					sb.WriteString(fmt.Sprintf("\t\thttpReq.Header.Set(\"%s\", fmt.Sprintf(\"%%v\", *req.%s))\n", param.Name, fieldName))
					sb.WriteString("\t}\n")
				}
			}
			sb.WriteString("\n")
		}
	}

	// Send request
	sb.WriteString("\t// Send request\n")
	sb.WriteString("\tresp, err := c.HTTPClient.Do(httpReq)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, fmt.Errorf(\"failed to send webhook: %w\", err)\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tdefer resp.Body.Close()\n\n")

	// Parse response
	sb.WriteString("\t// Parse response based on status code\n")
	sb.WriteString("\tswitch resp.StatusCode {\n")

	if op.Responses != nil {
		// Get status codes in sorted order
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

			concreteTypeName := fmt.Sprintf("%s%dResponse", handlerName, statusCodeInt)

			sb.WriteString(fmt.Sprintf("\tcase %d:\n", statusCodeInt))

			// Check if response has content
			hasBody := false
			if response.Content != nil {
				if jsonContent, ok := response.Content["application/json"]; ok && jsonContent.Schema != nil {
					hasBody = true
				}
			}

			if hasBody {
				sb.WriteString(fmt.Sprintf("\t\tvar result %s\n", concreteTypeName))
				sb.WriteString("\t\tbody, err := io.ReadAll(resp.Body)\n")
				sb.WriteString("\t\tif err != nil {\n")
				sb.WriteString("\t\t\treturn nil, fmt.Errorf(\"failed to read response body: %w\", err)\n")
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t\tif err := json.Unmarshal(body, &result.Body); err != nil {\n")
				sb.WriteString("\t\t\treturn nil, fmt.Errorf(\"failed to parse response body: %w\", err)\n")
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t\treturn result, nil\n")
			} else {
				sb.WriteString(fmt.Sprintf("\t\treturn %s{}, nil\n", concreteTypeName))
			}
		}
	}

	// Default case
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\tbody, _ := io.ReadAll(resp.Body)\n")
	sb.WriteString("\t\treturn nil, fmt.Errorf(\"unexpected status code %d: %s\", resp.StatusCode, string(body))\n")
	sb.WriteString("\t}\n")

	sb.WriteString("}\n\n")
}

// Helper functions

// generateWebhookName creates a webhook function name
func generateWebhookName(webhookName, method, operationID string) string {
	if operationID != "" {
		return toPascalCase(operationID)
	}

	// Generate from webhook name and method
	name := "send_" + webhookName + "_" + strings.ToLower(method)
	return toPascalCase(name)
}

// getParamType returns the Go type for a parameter (reused from server.go)
func (g *WebhookGenerator) getParamType(param *openapi.Parameter) string {
	if param.Schema == nil || param.Schema.Value == nil {
		return "string"
	}

	schema := param.Schema.Value
	schemaType := schema.GetSchemaType()

	switch schemaType {
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		} else if schema.Format == "int32" {
			return "int32"
		}
		return "int"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	case "string":
		return "string"
	default:
		return "string"
	}
}

// resolveSchemaType resolves a schema reference to a Go type (reused from server.go)
func (g *WebhookGenerator) resolveSchemaType(schemaRef *openapi.SchemaRef) string {
	if schemaRef == nil {
		return "any"
	}

	// If this is a reference, extract the type name
	if schemaRef.Ref != "" {
		parts := strings.Split(schemaRef.Ref, "/")
		if len(parts) > 0 {
			typeName := parts[len(parts)-1]
			return toPascalCase(typeName)
		}
	}

	// Otherwise resolve from schema
	if schemaRef.Value != nil {
		return g.resolveSchemaTypeFromValue(schemaRef.Value)
	}

	return "any"
}

// resolveSchemaTypeFromValue resolves the Go type from a schema value (reused from server.go)
func (g *WebhookGenerator) resolveSchemaTypeFromValue(schema *openapi.Schema) string {
	if schema == nil {
		return "any"
	}

	schemaType := schema.GetSchemaType()

	switch schemaType {
	case "array":
		if schema.Items != nil {
			itemType := g.resolveSchemaType(schema.Items)
			return "[]" + itemType
		}
		return "[]any"
	case "object":
		return "map[string]any"
	case "string":
		return "string"
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}
