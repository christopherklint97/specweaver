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
	var sb strings.Builder

	sb.WriteString("package api\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"encoding/json\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"io\"\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString("\t\"strconv\"\n")
	sb.WriteString("\n")
	sb.WriteString("\t\"github.com/christopherklint97/specweaver/pkg/router\"\n")
	sb.WriteString(")\n\n")

	// Generate HTTPError type
	g.generateHTTPError(&sb)

	// Generate request types for each operation
	if err := g.generateRequestTypes(&sb); err != nil {
		return "", err
	}

	// Generate response types for each operation
	if err := g.generateResponseTypes(&sb); err != nil {
		return "", err
	}

	// Generate the main server interface
	if err := g.generateServerInterface(&sb); err != nil {
		return "", err
	}

	// Generate the handler wrapper
	g.generateHandlerWrapper(&sb)

	// Generate the router setup
	g.generateRouter(&sb)

	// Generate helper functions
	g.generateHelpers(&sb)

	return sb.String(), nil
}

// generateHTTPError generates the HTTPError type for error handling
func (g *ServerGenerator) generateHTTPError(sb *strings.Builder) {
	sb.WriteString("// HTTPError represents an HTTP error with a status code\n")
	sb.WriteString("type HTTPError struct {\n")
	sb.WriteString("\tCode    int\n")
	sb.WriteString("\tMessage string\n")
	sb.WriteString("\tErr     error\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (e *HTTPError) Error() string {\n")
	sb.WriteString("\tif e.Err != nil {\n")
	sb.WriteString("\t\treturn fmt.Sprintf(\"%s: %v\", e.Message, e.Err)\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn e.Message\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (e *HTTPError) Unwrap() error {\n")
	sb.WriteString("\treturn e.Err\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// NewHTTPError creates a new HTTPError\n")
	sb.WriteString("func NewHTTPError(code int, message string) *HTTPError {\n")
	sb.WriteString("\treturn &HTTPError{Code: code, Message: message}\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// NewHTTPErrorf creates a new HTTPError with formatted message\n")
	sb.WriteString("func NewHTTPErrorf(code int, format string, args ...any) *HTTPError {\n")
	sb.WriteString("\treturn &HTTPError{Code: code, Message: fmt.Sprintf(format, args...)}\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// WrapHTTPError wraps an existing error with an HTTP status code\n")
	sb.WriteString("func WrapHTTPError(code int, err error, message string) *HTTPError {\n")
	sb.WriteString("\treturn &HTTPError{Code: code, Message: message, Err: err}\n")
	sb.WriteString("}\n\n")
}

// generateRequestTypes generates request structs for each operation
func (g *ServerGenerator) generateRequestTypes(sb *strings.Builder) error {
	if g.spec.Paths == nil {
		return nil
	}

	// Sort paths for deterministic output
	paths := make([]string, 0, len(g.spec.Paths))
	for path := range g.spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)
			requestTypeName := handlerName + "Request"

			sb.WriteString(fmt.Sprintf("// %s represents the request for %s\n", requestTypeName, handlerName))
			sb.WriteString(fmt.Sprintf("type %s struct {\n", requestTypeName))

			// Add path parameters
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param == nil {
						continue
					}

					if param.In == "path" {
						fieldName := toPascalCase(param.Name)
						fieldType := g.getParamType(param)
						if param.Description != "" {
							sb.WriteString(fmt.Sprintf("\t// %s\n", param.Description))
						}
						sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, fieldType, param.Name))
					}
				}
			}

			// Add query parameters
			if op.Parameters != nil {
				for _, param := range op.Parameters {
					if param == nil {
						continue
					}

					if param.In == "query" {
						fieldName := toPascalCase(param.Name)
						fieldType := g.getParamType(param)

						// Query params are optional by default
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

// generateResponseTypes generates response types for each operation
func (g *ServerGenerator) generateResponseTypes(sb *strings.Builder) error {
	if g.spec.Paths == nil {
		return nil
	}

	// Sort paths for deterministic output
	paths := make([]string, 0, len(g.spec.Paths))
	for path := range g.spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)
			responseTypeName := handlerName + "Response"

			// Generate response interface
			sb.WriteString(fmt.Sprintf("// %s represents possible responses for %s\n", responseTypeName, handlerName))
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

					// Skip "default" responses - these should be handled by error mechanism
					if statusCode == "default" {
						continue
					}

					// Parse status code
					statusCodeInt := parseStatusCode(statusCode)
					if statusCodeInt == 0 {
						// Skip invalid status codes
						continue
					}
					concreteTypeName := fmt.Sprintf("%s%dResponse", handlerName, statusCodeInt)

					sb.WriteString(fmt.Sprintf("// %s represents a %d response\n", concreteTypeName, statusCodeInt))
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

// generateServerInterface generates the interface that users need to implement
func (g *ServerGenerator) generateServerInterface(sb *strings.Builder) error {
	sb.WriteString("// Server represents all server handlers\n")
	sb.WriteString("type Server interface {\n")

	if g.spec.Paths == nil {
		sb.WriteString("}\n\n")
		return nil
	}

	// Sort paths for deterministic output
	paths := make([]string, 0, len(g.spec.Paths))
	for path := range g.spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)
			requestTypeName := handlerName + "Request"
			responseTypeName := handlerName + "Response"

			// Add comment with operation summary
			if op.Summary != "" {
				sb.WriteString(fmt.Sprintf("\t// %s %s\n", handlerName, op.Summary))
			}

			sb.WriteString(fmt.Sprintf("\t%s(ctx context.Context, req %s) (%s, error)\n", handlerName, requestTypeName, responseTypeName))
		}
	}

	sb.WriteString("}\n\n")
	return nil
}

// generateHandlerWrapper generates the HTTP handler wrapper with adapter functions
func (g *ServerGenerator) generateHandlerWrapper(sb *strings.Builder) {
	sb.WriteString("// ServerWrapper wraps the Server with HTTP handler logic\n")
	sb.WriteString("type ServerWrapper struct {\n")
	sb.WriteString("\tHandler Server\n")
	sb.WriteString("}\n\n")

	if g.spec.Paths == nil {
		return
	}

	// Sort paths for deterministic output
	paths := make([]string, 0, len(g.spec.Paths))
	for path := range g.spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	// Generate adapter methods for each operation
	for _, path := range paths {
		pathItem := g.spec.Paths[path]
		operations := getOperationsInOrder(pathItem)

		for _, methodOp := range operations {
			method := methodOp.Method
			op := methodOp.Operation

			handlerName := generateHandlerName(method, path, op.OperationID)
			g.generateAdapterMethod(sb, handlerName, path, op)
		}
	}

	// Generate error handler
	sb.WriteString("// handleError handles errors and writes appropriate HTTP responses\n")
	sb.WriteString("func (w *ServerWrapper) handleError(rw http.ResponseWriter, err error) {\n")
	sb.WriteString("\tvar httpErr *HTTPError\n")
	sb.WriteString("\tif errors.As(err, &httpErr) {\n")
	sb.WriteString("\t\tWriteError(rw, httpErr.Code, httpErr)\n")
	sb.WriteString("\t\treturn\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\t// Default to 500 Internal Server Error\n")
	sb.WriteString("\tWriteError(rw, http.StatusInternalServerError, err)\n")
	sb.WriteString("}\n\n")
}

// generateAdapterMethod generates an adapter method that bridges HTTP to the handler
func (g *ServerGenerator) generateAdapterMethod(sb *strings.Builder, handlerName, path string, op *openapi.Operation) {
	requestTypeName := handlerName + "Request"
	adapterMethodName := "handle" + handlerName

	sb.WriteString(fmt.Sprintf("// %s adapts HTTP request to %s handler\n", adapterMethodName, handlerName))
	sb.WriteString(fmt.Sprintf("func (w *ServerWrapper) %s(rw http.ResponseWriter, r *http.Request) {\n", adapterMethodName))
	sb.WriteString("\tctx := r.Context()\n")
	sb.WriteString(fmt.Sprintf("\treq := %s{}\n\n", requestTypeName))

	// Parse path parameters
	if op.Parameters != nil {
		for _, param := range op.Parameters {
			if param == nil {
				continue
			}

			if param.In == "path" {
				fieldName := toPascalCase(param.Name)
				g.generateParamParsing(sb, param, fieldName, true)
			}
		}
	}

	// Parse query parameters
	if op.Parameters != nil {
		for _, param := range op.Parameters {
			if param == nil {
				continue
			}

			if param.In == "query" {
				fieldName := toPascalCase(param.Name)
				g.generateParamParsing(sb, param, fieldName, false)
			}
		}
	}

	// Parse request body
	if op.RequestBody != nil {
		content := op.RequestBody.Content
		if _, ok := content["application/json"]; ok {
			sb.WriteString("\t// Parse request body\n")
			sb.WriteString("\tif err := ReadJSON(r, &req.Body); err != nil {\n")
			sb.WriteString("\t\tw.handleError(rw, NewHTTPError(http.StatusBadRequest, \"invalid request body\"))\n")
			sb.WriteString("\t\treturn\n")
			sb.WriteString("\t}\n\n")
		}
	}

	// Call the handler
	sb.WriteString("\t// Call handler\n")
	sb.WriteString(fmt.Sprintf("\tresp, err := w.Handler.%s(ctx, req)\n", handlerName))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\tw.handleError(rw, err)\n")
	sb.WriteString("\t\treturn\n")
	sb.WriteString("\t}\n\n")

	// Write response
	sb.WriteString("\t// Write response\n")
	sb.WriteString("\tWriteResponse(rw, resp)\n")
	sb.WriteString("}\n\n")
}

// generateParamParsing generates code to parse a parameter
func (g *ServerGenerator) generateParamParsing(sb *strings.Builder, param *openapi.Parameter, fieldName string, isPath bool) {
	paramType := g.getParamType(param)
	paramName := param.Name

	// Get parameter value
	if isPath {
		sb.WriteString(fmt.Sprintf("\t// Parse path parameter: %s\n", paramName))
		sb.WriteString(fmt.Sprintf("\t%sStr := router.URLParam(r, \"%s\")\n", paramName, paramName))
	} else {
		sb.WriteString(fmt.Sprintf("\t// Parse query parameter: %s\n", paramName))
		sb.WriteString(fmt.Sprintf("\t%sStr := r.URL.Query().Get(\"%s\")\n", paramName, paramName))
	}

	// Parse based on type
	baseType := strings.TrimPrefix(paramType, "*")

	switch baseType {
	case "string":
		if param.Required || isPath {
			sb.WriteString(fmt.Sprintf("\treq.%s = %sStr\n", fieldName, paramName))
		} else {
			sb.WriteString(fmt.Sprintf("\tif %sStr != \"\" {\n", paramName))
			sb.WriteString(fmt.Sprintf("\t\treq.%s = &%sStr\n", fieldName, paramName))
			sb.WriteString("\t}\n")
		}
	case "int", "int32", "int64":
		bitSize := "0"
		if baseType == "int32" {
			bitSize = "32"
		} else if baseType == "int64" {
			bitSize = "64"
		}

		if param.Required || isPath {
			sb.WriteString(fmt.Sprintf("\t%sVal, err := strconv.ParseInt(%sStr, 10, %s)\n", paramName, paramName, bitSize))
			sb.WriteString("\tif err != nil {\n")
			sb.WriteString(fmt.Sprintf("\t\tw.handleError(rw, NewHTTPError(http.StatusBadRequest, \"invalid %s parameter\"))\n", paramName))
			sb.WriteString("\t\treturn\n")
			sb.WriteString("\t}\n")
			if baseType == "int" {
				sb.WriteString(fmt.Sprintf("\treq.%s = int(%sVal)\n", fieldName, paramName))
			} else {
				sb.WriteString(fmt.Sprintf("\treq.%s = %s(%sVal)\n", fieldName, baseType, paramName))
			}
		} else {
			sb.WriteString(fmt.Sprintf("\tif %sStr != \"\" {\n", paramName))
			sb.WriteString(fmt.Sprintf("\t\t%sVal, err := strconv.ParseInt(%sStr, 10, %s)\n", paramName, paramName, bitSize))
			sb.WriteString("\t\tif err == nil {\n")
			if baseType == "int" {
				sb.WriteString(fmt.Sprintf("\t\t\t%sInt := int(%sVal)\n", paramName, paramName))
				sb.WriteString(fmt.Sprintf("\t\t\treq.%s = &%sInt\n", fieldName, paramName))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t\t%sTyped := %s(%sVal)\n", paramName, baseType, paramName))
				sb.WriteString(fmt.Sprintf("\t\t\treq.%s = &%sTyped\n", fieldName, paramName))
			}
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
		}
	case "float32", "float64":
		bitSize := "32"
		if baseType == "float64" {
			bitSize = "64"
		}

		if param.Required || isPath {
			sb.WriteString(fmt.Sprintf("\t%sVal, err := strconv.ParseFloat(%sStr, %s)\n", paramName, paramName, bitSize))
			sb.WriteString("\tif err != nil {\n")
			sb.WriteString(fmt.Sprintf("\t\tw.handleError(rw, NewHTTPError(http.StatusBadRequest, \"invalid %s parameter\"))\n", paramName))
			sb.WriteString("\t\treturn\n")
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\treq.%s = %s(%sVal)\n", fieldName, baseType, paramName))
		} else {
			sb.WriteString(fmt.Sprintf("\tif %sStr != \"\" {\n", paramName))
			sb.WriteString(fmt.Sprintf("\t\t%sVal, err := strconv.ParseFloat(%sStr, %s)\n", paramName, paramName, bitSize))
			sb.WriteString("\t\tif err == nil {\n")
			sb.WriteString(fmt.Sprintf("\t\t\t%sTyped := %s(%sVal)\n", paramName, baseType, paramName))
			sb.WriteString(fmt.Sprintf("\t\t\treq.%s = &%sTyped\n", fieldName, paramName))
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
		}
	case "bool":
		if param.Required || isPath {
			sb.WriteString(fmt.Sprintf("\t%sVal, err := strconv.ParseBool(%sStr)\n", paramName, paramName))
			sb.WriteString("\tif err != nil {\n")
			sb.WriteString(fmt.Sprintf("\t\tw.handleError(rw, NewHTTPError(http.StatusBadRequest, \"invalid %s parameter\"))\n", paramName))
			sb.WriteString("\t\treturn\n")
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\treq.%s = %sVal\n", fieldName, paramName))
		} else {
			sb.WriteString(fmt.Sprintf("\tif %sStr != \"\" {\n", paramName))
			sb.WriteString(fmt.Sprintf("\t\t%sVal, err := strconv.ParseBool(%sStr)\n", paramName, paramName))
			sb.WriteString("\t\tif err == nil {\n")
			sb.WriteString(fmt.Sprintf("\t\t\treq.%s = &%sVal\n", fieldName, paramName))
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
		}
	}

	sb.WriteString("\n")
}

// generateRouter generates the router setup functions
func (g *ServerGenerator) generateRouter(sb *strings.Builder) {
	// Generate ConfigureRouter function that works with any router
	sb.WriteString("// ConfigureRouter configures the given router with all routes.\n")
	sb.WriteString("// This function allows you to use any router that implements the router.Router interface.\n")
	sb.WriteString("//\n")
	sb.WriteString("// Example with built-in router:\n")
	sb.WriteString("//\n")
	sb.WriteString("//\tr := router.NewRouter()\n")
	sb.WriteString("//\tConfigureRouter(r, myServer)\n")
	sb.WriteString("//\n")
	sb.WriteString("// Example with custom router:\n")
	sb.WriteString("//\n")
	sb.WriteString("//\tr := myCustomRouter.New() // Must implement router.Router interface\n")
	sb.WriteString("//\tConfigureRouter(r, myServer)\n")
	sb.WriteString("func ConfigureRouter(r router.Router, si Server) {\n")
	sb.WriteString("\twrapper := &ServerWrapper{Handler: si}\n")
	sb.WriteString("\n")

	if g.spec.Paths != nil {
		// Sort paths for deterministic output
		paths := make([]string, 0, len(g.spec.Paths))
		for path := range g.spec.Paths {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
			pathItem := g.spec.Paths[path]
			routerPath := convertToRouterPath(path)
			operations := getOperationsInOrder(pathItem)

			for _, methodOp := range operations {
				method := methodOp.Method
				op := methodOp.Operation

				handlerName := generateHandlerName(method, path, op.OperationID)
				adapterMethodName := "handle" + handlerName

				sb.WriteString(fmt.Sprintf("\tr.%s(\"%s\", wrapper.%s)\n",
					getRouterMethodName(method), routerPath, adapterMethodName))
			}
		}
	}

	sb.WriteString("}\n\n")

	// Generate NewRouter function for convenience (uses built-in router)
	sb.WriteString("// NewRouter creates a new router with all routes configured using the built-in router.\n")
	sb.WriteString("// For using a custom router, use ConfigureRouter instead.\n")
	sb.WriteString("func NewRouter(si Server) *router.Mux {\n")
	sb.WriteString("\tr := router.NewRouter()\n")
	sb.WriteString("\n")
	sb.WriteString("\t// Default middleware\n")
	sb.WriteString("\tr.Use(router.Logger)\n")
	sb.WriteString("\tr.Use(router.Recoverer)\n")
	sb.WriteString("\tr.Use(router.RequestID)\n")
	sb.WriteString("\tr.Use(router.RealIP)\n")
	sb.WriteString("\n")
	sb.WriteString("\tConfigureRouter(r, si)\n")
	sb.WriteString("\treturn r\n")
	sb.WriteString("}\n\n")
}

// generateHelpers generates helper functions for request/response handling
func (g *ServerGenerator) generateHelpers(sb *strings.Builder) {
	sb.WriteString("// Helper functions for request/response handling\n\n")

	// JSON response helper
	sb.WriteString("// WriteJSON writes a JSON response\n")
	sb.WriteString("func WriteJSON(w http.ResponseWriter, status int, v any) error {\n")
	sb.WriteString("\tw.Header().Set(\"Content-Type\", \"application/json\")\n")
	sb.WriteString("\tw.WriteHeader(status)\n")
	sb.WriteString("\treturn json.NewEncoder(w).Encode(v)\n")
	sb.WriteString("}\n\n")

	// Generic response writer
	sb.WriteString("// WriteResponse writes a response based on its type\n")
	sb.WriteString("func WriteResponse(w http.ResponseWriter, resp any) error {\n")
	sb.WriteString("\t// Extract status code and body using type assertion\n")
	sb.WriteString("\ttype responseWriter interface {\n")
	sb.WriteString("\t\tStatusCode() int\n")
	sb.WriteString("\t\tResponseBody() any\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tif rw, ok := resp.(responseWriter); ok {\n")
	sb.WriteString("\t\tstatusCode := rw.StatusCode()\n")
	sb.WriteString("\t\tbody := rw.ResponseBody()\n")
	sb.WriteString("\t\t// For 204 No Content or nil body, don't write a body\n")
	sb.WriteString("\t\tif statusCode == http.StatusNoContent || body == nil {\n")
	sb.WriteString("\t\t\tw.WriteHeader(statusCode)\n")
	sb.WriteString("\t\t\treturn nil\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\treturn WriteJSON(w, statusCode, body)\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\t// Fallback to 200 OK\n")
	sb.WriteString("\treturn WriteJSON(w, http.StatusOK, resp)\n")
	sb.WriteString("}\n\n")

	// Error response helper
	sb.WriteString("// ErrorResponse represents an error response\n")
	sb.WriteString("type ErrorResponse struct {\n")
	sb.WriteString("\tError   string `json:\"error\"`\n")
	sb.WriteString("\tMessage string `json:\"message,omitempty\"`\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// WriteError writes an error response\n")
	sb.WriteString("func WriteError(w http.ResponseWriter, status int, err error) {\n")
	sb.WriteString("\tWriteJSON(w, status, ErrorResponse{\n")
	sb.WriteString("\t\tError:   http.StatusText(status),\n")
	sb.WriteString("\t\tMessage: err.Error(),\n")
	sb.WriteString("\t})\n")
	sb.WriteString("}\n\n")

	// Read JSON helper
	sb.WriteString("// ReadJSON reads and decodes JSON from request body\n")
	sb.WriteString("func ReadJSON(r *http.Request, v any) error {\n")
	sb.WriteString("\tdefer r.Body.Close()\n")
	sb.WriteString("\tbody, err := io.ReadAll(r.Body)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn err\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn json.Unmarshal(body, v)\n")
	sb.WriteString("}\n\n")
}

// Helper functions

// getParamType returns the Go type for a parameter
func (g *ServerGenerator) getParamType(param *openapi.Parameter) string {
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

// resolveSchemaType resolves a schema reference to a Go type
func (g *ServerGenerator) resolveSchemaType(schemaRef *openapi.SchemaRef) string {
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

// resolveSchemaTypeFromValue resolves the Go type from a schema value
func (g *ServerGenerator) resolveSchemaTypeFromValue(schema *openapi.Schema) string {
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

// parseStatusCode parses a status code string to int
// Returns 0 for "default" or invalid codes, which should be filtered out by the caller
func parseStatusCode(code string) int {
	statusCode, err := strconv.Atoi(code)
	if err != nil {
		return 0 // Invalid code (including "default")
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
