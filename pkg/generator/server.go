package generator

import (
	"fmt"
	"net/http"
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
	sb.WriteString("\t\"encoding/json\"\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString("\t\"io\"\n")
	sb.WriteString("\n")
	sb.WriteString("\t\"github.com/christopherklint97/specweaver/pkg/router\"\n")
	sb.WriteString(")\n\n")

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

// generateServerInterface generates the interface that users need to implement
func (g *ServerGenerator) generateServerInterface(sb *strings.Builder) error {
	sb.WriteString("// Server represents all server handlers\n")
	sb.WriteString("type Server interface {\n")

	if g.spec.Paths == nil {
		sb.WriteString("}\n\n")
		return nil
	}

	for path, pathItem := range g.spec.Paths {
		operations := map[string]*openapi.Operation{
			http.MethodGet:    pathItem.Get,
			http.MethodPost:   pathItem.Post,
			http.MethodPut:    pathItem.Put,
			http.MethodPatch:  pathItem.Patch,
			http.MethodDelete: pathItem.Delete,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}

			handlerName := generateHandlerName(method, path, op.OperationID)

			// Add comment with operation summary
			if op.Summary != "" {
				sb.WriteString(fmt.Sprintf("\t// %s %s\n", handlerName, op.Summary))
			}

			sb.WriteString(fmt.Sprintf("\t%s(w http.ResponseWriter, r *http.Request)\n", handlerName))
		}
	}

	sb.WriteString("}\n\n")
	return nil
}

// generateHandlerWrapper generates the HTTP handler wrapper
func (g *ServerGenerator) generateHandlerWrapper(sb *strings.Builder) {
	sb.WriteString("// ServerWrapper wraps the Server with HTTP handler logic\n")
	sb.WriteString("type ServerWrapper struct {\n")
	sb.WriteString("\tHandler Server\n")
	sb.WriteString("}\n\n")
}

// generateRouter generates the router setup function
func (g *ServerGenerator) generateRouter(sb *strings.Builder) {
	sb.WriteString("// NewRouter creates a new router with all routes configured\n")
	sb.WriteString("func NewRouter(si Server) *router.Mux {\n")
	sb.WriteString("\tr := router.NewRouter()\n")
	sb.WriteString("\n")
	sb.WriteString("\t// Middleware\n")
	sb.WriteString("\tr.Use(router.Logger)\n")
	sb.WriteString("\tr.Use(router.Recoverer)\n")
	sb.WriteString("\tr.Use(router.RequestID)\n")
	sb.WriteString("\tr.Use(router.RealIP)\n")
	sb.WriteString("\n")

	sb.WriteString("\twrapper := &ServerWrapper{Handler: si}\n")
	sb.WriteString("\n")

	if g.spec.Paths != nil {
		for path, pathItem := range g.spec.Paths {
			chiPath := convertToChiPath(path)

			operations := map[string]*openapi.Operation{
				http.MethodGet:    pathItem.Get,
				http.MethodPost:   pathItem.Post,
				http.MethodPut:    pathItem.Put,
				http.MethodPatch:  pathItem.Patch,
				http.MethodDelete: pathItem.Delete,
			}

			for method, op := range operations {
				if op == nil {
					continue
				}

				handlerName := generateHandlerName(method, path, op.OperationID)

				sb.WriteString(fmt.Sprintf("\tr.%s(\"%s\", wrapper.Handler.%s)\n",
					getRouterMethodName(method), chiPath, handlerName))
			}
		}
	}

	sb.WriteString("\n\treturn r\n")
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

// convertToChiPath converts OpenAPI path to router path format
func convertToChiPath(path string) string {
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
