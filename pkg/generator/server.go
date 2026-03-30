package generator

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// ServerGenerator generates Go server code from OpenAPI paths
type ServerGenerator struct {
	spec *openapi.Document
}

// NewServerGenerator creates a new ServerGenerator instance
func NewServerGenerator(spec *openapi.Document) *ServerGenerator {
	return &ServerGenerator{
		spec: spec,
	}
}

// Generate generates server code including handlers and router
func (g *ServerGenerator) Generate() (string, error) {
	// Build template data
	data := ServerTemplateData{
		NeedsStrconv:       g.needsStrconvImport(),
		HasSecuritySchemes: g.hasSecuritySchemes(),
		RequestTypes:       g.buildRequestTypes(),
		ResponseTypes:      g.buildResponseTypes(),
		Operations:         g.buildOperations(),
		SecuritySchemes:    g.buildSecuritySchemes(),
		Routes:             g.buildRoutes(),
	}

	// Execute template
	return executeTemplate("server.go.tmpl", data)
}

// hasSecuritySchemes checks if there are security schemes defined
func (g *ServerGenerator) hasSecuritySchemes() bool {
	return g.spec.Components != nil &&
		g.spec.Components.SecuritySchemes != nil &&
		len(g.spec.Components.SecuritySchemes) > 0
}

// buildRequestTypes builds request type data for all operations
func (g *ServerGenerator) buildRequestTypes() []RequestTypeData {
	if g.spec.Paths == nil {
		return nil
	}

	var result []RequestTypeData

	// Sort paths for deterministic output
	paths := g.getSortedPaths()

	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)
			requestTypeName := handlerName + "Request"

			reqType := RequestTypeData{
				Name:        requestTypeName,
				HandlerName: handlerName,
				Fields:      []FieldDefinition{},
			}

			// Add path parameters
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param != nil && param.In == "path" {
						fieldName := toPascalCase(param.Name)
						fieldType := getParamType(param)
						field := FieldDefinition{
							Name:        fieldName,
							Type:        fieldType,
							JSONTag:     param.Name,
							Description: param.Description,
						}
						reqType.Fields = append(reqType.Fields, field)
					}
				}
			}

			// Add query parameters
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param != nil && param.In == "query" {
						fieldName := toPascalCase(param.Name)
						fieldType := getParamType(param)

						// Query params are optional by default
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

// buildResponseTypes builds response type data for all operations
func (g *ServerGenerator) buildResponseTypes() []ResponseTypeData {
	if g.spec.Paths == nil {
		return nil
	}

	var result []ResponseTypeData

	paths := g.getSortedPaths()

	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)
			responseTypeName := handlerName + "Response"

			respType := ResponseTypeData{
				InterfaceName: responseTypeName,
				HandlerName:   handlerName,
				ConcreteTypes: []ConcreteResponseType{},
			}

			// Generate concrete response types for each status code
			if op.Responses != nil {
				statusCodes := g.getSortedStatusCodes(op.Responses)

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

// buildOperations builds operation data for all operations
func (g *ServerGenerator) buildOperations() []OperationData {
	if g.spec.Paths == nil {
		return nil
	}

	var result []OperationData

	paths := g.getSortedPaths()

	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)

			opData := OperationData{
				HandlerName:      handlerName,
				Summary:          op.Summary,
				RequestTypeName:  handlerName + "Request",
				ResponseTypeName: handlerName + "Response",
				PathParams:       []ParamData{},
				QueryParams:      []ParamData{},
				HasRequestBody:   false,
			}

			// Build path parameters
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param != nil && param.In == "path" {
						paramData := g.buildParamData(param, true)
						opData.PathParams = append(opData.PathParams, paramData)
					}
				}
			}

			// Build query parameters
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param != nil && param.In == "query" {
						paramData := g.buildParamData(param, false)
						opData.QueryParams = append(opData.QueryParams, paramData)
					}
				}
			}

			// Check for request body
			if op.RequestBody != nil {
				content := op.RequestBody.Content
				if _, ok := content["application/json"]; ok {
					opData.HasRequestBody = true
				}
			}

			result = append(result, opData)
		}
	}

	return result
}

// buildParamData builds parameter data for a parameter
func (g *ServerGenerator) buildParamData(param *openapi.Parameter, isPath bool) ParamData {
	paramType := getParamType(param)
	baseType := strings.TrimPrefix(paramType, "*")

	bitSize := "0"
	if baseType == "int32" {
		bitSize = "32"
	} else if baseType == "int64" {
		bitSize = "64"
	} else if baseType == "float32" {
		bitSize = "32"
	} else if baseType == "float64" {
		bitSize = "64"
	}

	return ParamData{
		ParamName: param.Name,
		FieldName: toPascalCase(param.Name),
		BaseType:  baseType,
		BitSize:   bitSize,
		Required:  param.Required || isPath,
	}
}

// buildSecuritySchemes builds security scheme data
func (g *ServerGenerator) buildSecuritySchemes() []SecuritySchemeData {
	if !g.hasSecuritySchemes() {
		return nil
	}

	var result []SecuritySchemeData

	// Get security scheme names in sorted order
	schemes := make([]string, 0, len(g.spec.Components.SecuritySchemes))
	for name := range g.spec.Components.SecuritySchemes {
		schemes = append(schemes, name)
	}
	sort.Strings(schemes)

	for _, name := range schemes {
		scheme := g.spec.Components.SecuritySchemes[name]
		if scheme == nil {
			continue
		}

		result = append(result, SecuritySchemeData{
			Name:      name,
			Type:      scheme.Type,
			Scheme:    scheme.Scheme,
			In:        scheme.In,
			ParamName: scheme.Name,
		})
	}

	return result
}

// buildRoutes builds route data for all operations
func (g *ServerGenerator) buildRoutes() []RouteData {
	if g.spec.Paths == nil {
		return nil
	}

	var result []RouteData

	paths := g.getSortedPaths()
	hasSecuritySchemes := g.hasSecuritySchemes()

	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		routerPath := convertToRouterPath(path)
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)

			route := RouteData{
				Method:      getRouterMethodName(method),
				Path:        routerPath,
				HandlerName: handlerName,
				HasSecurity: hasSecuritySchemes && g.hasSecurityRequirements(op),
			}

			if route.HasSecurity {
				route.SecurityReqs = g.generateSecurityRequirementsLiteral(op)
			}

			result = append(result, route)
		}
	}

	return result
}

// getSortedPaths returns paths in sorted order
func (g *ServerGenerator) getSortedPaths() []string {
	paths := make([]string, 0, len(g.spec.Paths))
	for path := range g.spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// getSortedStatusCodes returns status codes in sorted order
func (g *ServerGenerator) getSortedStatusCodes(responses map[string]*openapi.Response) []string {
	statusCodes := make([]string, 0, len(responses))
	for statusCode := range responses {
		statusCodes = append(statusCodes, statusCode)
	}
	sort.Strings(statusCodes)
	return statusCodes
}

// hasSecurityRequirements checks if an operation has security requirements
func (g *ServerGenerator) hasSecurityRequirements(op *openapi.Operation) bool {
	// Check operation-level security
	if len(op.Security) > 0 {
		return true
	}

	// Check global security (if operation doesn't override)
	if op.Security == nil && len(g.spec.Security) > 0 {
		return true
	}

	return false
}

// generateSecurityRequirementsLiteral generates a Go literal for security requirements
func (g *ServerGenerator) generateSecurityRequirementsLiteral(op *openapi.Operation) string {
	var sb strings.Builder

	// Use operation-level security if present, otherwise use global
	securityReqs := op.Security
	if securityReqs == nil {
		securityReqs = g.spec.Security
	}

	sb.WriteString("[]map[string][]string{\n")
	for _, req := range securityReqs {
		sb.WriteString("\t\t{\n")

		// Get keys in sorted order for determinism
		keys := make([]string, 0, len(req))
		for k := range req {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, schemeName := range keys {
			scopes := req[schemeName]
			sb.WriteString(fmt.Sprintf("\t\t\t\"%s\": ", schemeName))
			if len(scopes) == 0 {
				sb.WriteString("[]string{},\n")
			} else {
				sb.WriteString("[]string{")
				for i, scope := range scopes {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("\"%s\"", scope))
				}
				sb.WriteString("},\n")
			}
		}
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t}")

	return sb.String()
}

// parseStatusCode parses a status code string to int
func parseStatusCode(code string) int {
	statusCode, err := strconv.Atoi(code)
	if err != nil {
		return 0
	}
	return statusCode
}

// generateHandlerName creates a handler function name from method, path and operationID
func generateHandlerName(method, path, operationID string) string {
	if operationID != "" {
		return toPascalCase(operationID)
	}

	// Generate from method and path
	name := strings.ToLower(method)
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	for _, part := range pathParts {
		// Skip path parameters
		if !strings.HasPrefix(part, "{") {
			name += "_" + part
		}
	}

	return toPascalCase(name)
}

// convertToRouterPath converts OpenAPI path to router path format
func convertToRouterPath(path string) string {
	// Both OpenAPI and our router use {param} format
	return path
}

// getRouterMethodName returns the router method name for an HTTP method
func getRouterMethodName(method string) string {
	switch method {
	case http.MethodGet:
		return "Get"
	case http.MethodPost:
		return "Post"
	case http.MethodPut:
		return "Put"
	case http.MethodPatch:
		return "Patch"
	case http.MethodDelete:
		return "Delete"
	case http.MethodOptions:
		return "Options"
	case http.MethodHead:
		return "Head"
	default:
		return "Get"
	}
}

// methodOperation represents an HTTP method and its operation
type methodOperation struct {
	Method    string
	Operation *openapi.Operation
}

// getOperationsInOrder returns operations for a path item in deterministic order
func getOperationsInOrder(pathItem *openapi.PathItem) []methodOperation {
	// Define the order of HTTP methods for determinism
	methodOrder := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodHead,
	}

	var result []methodOperation
	for _, method := range methodOrder {
		var op *openapi.Operation
		switch method {
		case http.MethodGet:
			op = pathItem.Get
		case http.MethodPost:
			op = pathItem.Post
		case http.MethodPut:
			op = pathItem.Put
		case http.MethodPatch:
			op = pathItem.Patch
		case http.MethodDelete:
			op = pathItem.Delete
		case http.MethodOptions:
			op = pathItem.Options
		case http.MethodHead:
			op = pathItem.Head
		}

		if op != nil {
			result = append(result, methodOperation{
				Method:    method,
				Operation: op,
			})
		}
	}

	return result
}

// needsStrconvImport checks if any parameters require strconv for parsing
func (g *ServerGenerator) needsStrconvImport() bool {
	if g.spec.Paths == nil {
		return false
	}

	for _, pathItem := range g.spec.Paths {
		operations := getOperationsInOrder(pathItem)
		for _, methodOp := range operations {
			op := methodOp.Operation
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param == nil || param.Schema == nil || param.Schema.Value == nil {
						continue
					}

					// Check if parameter is in path or query (these get parsed)
					if param.In != "path" && param.In != "query" {
						continue
					}

					schemaType := param.Schema.Value.GetSchemaType()
					// strconv is needed for integer, number, and boolean types
					if schemaType == "integer" || schemaType == "number" || schemaType == "boolean" {
						return true
					}
				}
			}
		}
	}

	return false
}
