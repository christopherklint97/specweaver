# SpecWeaver

> A powerful OpenAPI 3.x Go code generator that weaves your API specifications into production-ready server code.

SpecWeaver automatically generates type-safe Go code from OpenAPI specifications, including data types, HTTP handlers, and routing logic. Uses a custom robust OpenAPI parser supporting versions 3.0.x, 3.1.x, and 3.2.x with no external OpenAPI dependencies.

## Features

- ✨ **Full OpenAPI 3.x Support** - Compatible with OpenAPI 3.0.x, 3.1.x, and 3.2.x
- 🔧 **Custom Robust Parser** - No external OpenAPI library dependencies
- 🎯 **Type-Safe Code** - Generates idiomatic Go structs with proper types
- 🔐 **Authentication Support** - Automatic generation of auth middleware for all OpenAPI security schemes (Basic, Bearer, API Key, OAuth2, OIDC)
- 🚀 **Production Ready** - Includes error handling, middleware, and best practices
- 📝 **Documentation Preserved** - OpenAPI descriptions become Go comments
- 🔄 **Schema References** - Properly resolves `$ref` to generate correct types
- 🎨 **Idiomatic Go** - Follows Go conventions and best practices (uses `any` instead of `interface{}`)
- ⚡ **Zero Dependencies** - Custom lightweight router, no external dependencies
- 🔌 **Custom Router Support** - Use any router (chi, gorilla/mux, etc.) via pluggable interface
- 📄 **Format Support** - Works with both YAML and JSON specifications

## Installation

### As a CLI Tool

```bash
# Clone the repository
git clone https://github.com/christopherklint97/specweaver.git
cd specweaver

# Build the tool
go build -o specweaver ./cmd/specweaver

# Optionally, install globally
go install ./cmd/specweaver
```

### As a Go Module Library

```bash
# Add to your project
go get github.com/christopherklint97/specweaver@latest
```

Then import in your Go code:

```go
import "github.com/christopherklint97/specweaver"
```

## Quick Start

### CLI Usage

### 1. Generate Code from OpenAPI Spec

```bash
./specweaver -spec examples/petstore.yaml -output ./generated
```

**Options:**
- `-spec` - Path to your OpenAPI specification file (YAML or JSON) - **required**
- `-output` - Output directory for generated code (default: `./generated`)
- `-package` - Package name for generated code (default: `api`)
- `-version` - Show version information

### 2. Implement the Generated Interface

The generator creates a `Server` interface with clean, testable methods using `context.Context`:

```go
package main

import (
    "context"
    "net/http"
    "github.com/yourorg/yourapp/generated/api"
)

type MyServer struct {
    // Your application state
}

// Implement the interface methods with context-based handlers
func (s *MyServer) ListPets(ctx context.Context, req api.ListPetsRequest) (api.ListPetsResponse, error) {
    // Access query parameters
    limit := 20
    if req.Limit != nil {
        limit = int(*req.Limit)
    }

    pets := []api.Pet{
        {Id: 1, Name: "Fluffy", Status: api.PetStatusAvailable},
    }

    // Return typed response
    return api.ListPets200Response{Body: pets}, nil
}

func (s *MyServer) CreatePet(ctx context.Context, req api.CreatePetRequest) (api.CreatePetResponse, error) {
    // Validation with custom status code
    if req.Body.Name == "" {
        return nil, api.NewHTTPError(http.StatusBadRequest, "name is required")
    }

    // Process the new pet...
    pet := api.Pet{
        Id:   1,
        Name: req.Body.Name,
    }

    // Return 201 Created response
    return api.CreatePet201Response{Body: pet}, nil
}

func (s *MyServer) GetPetById(ctx context.Context, req api.GetPetByIdRequest) (api.GetPetByIdResponse, error) {
    // Access path parameters
    petId := req.PetId

    pet, exists := s.findPet(petId)
    if !exists {
        // Return 404 response (not an error!)
        return api.GetPetById404Response{
            Body: api.Error{
                Error:   "Not Found",
                Message: "pet not found",
            },
        }, nil
    }

    return api.GetPetById200Response{Body: pet}, nil
}

// Helper method
func (s *MyServer) findPet(id int64) (api.Pet, bool) {
    // Your pet lookup logic
    return api.Pet{}, false
}
```

### 3. Start Your Server

```go
func main() {
    server := &MyServer{}
    router := api.NewRouter(server)

    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", router)
}
```

### Library Usage

You can also use SpecWeaver as a Go module library to programmatically generate code in your applications, build tools, or CI/CD pipelines.

#### Simple API

```go
package main

import (
    "log"
    "github.com/christopherklint97/specweaver"
)

func main() {
    // Generate code with a single function call
    err := specweaver.Generate("openapi.yaml", specweaver.Options{
        OutputDir:   "./generated",
        PackageName: "api",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

#### Advanced API

For more control, use the parser and generator separately:

```go
package main

import (
    "fmt"
    "log"
    "github.com/christopherklint97/specweaver"
)

func main() {
    // Parse the OpenAPI specification
    parser := specweaver.NewParser()
    err := parser.ParseFile("openapi.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Access the parsed spec
    spec := parser.GetSpec()
    fmt.Printf("Generating code for: %s v%s\n", spec.Info.Title, spec.Info.Version)

    // Generate code with custom options
    generator := specweaver.NewGenerator(spec, specweaver.Options{
        OutputDir:   "./api",
        PackageName: "myapi",
    })

    err = generator.Generate()
    if err != nil {
        log.Fatal(err)
    }
}
```

**Use Cases:**
- Build tool integration (mage, make, custom scripts)
- CI/CD pipeline automation
- Dynamic API generation
- Multi-spec batch processing

See [examples/library/](examples/library/) for complete examples and integration patterns.

### Custom Router Support

SpecWeaver supports using any HTTP router that implements the `router.Router` interface. This allows you to use popular routers like chi, gorilla/mux, or httprouter with SpecWeaver-generated code.

#### Using the Built-in Router

```go
server := &MyServer{}
router := api.NewRouter(server) // Uses built-in router with default middleware
http.ListenAndServe(":8080", router)
```

#### Using a Custom Router

```go
// Create your custom router
customRouter := NewChiAdapter() // Adapter for chi router

// Add your middleware
customRouter.Use(middleware.Logger)
customRouter.Use(ChiURLParamMiddleware) // Required for URL parameter compatibility

// Configure with SpecWeaver routes
api.ConfigureRouter(customRouter, server)

// Start server
http.ListenAndServe(":8080", customRouter)
```

**Requirements for custom routers:**
1. Implement the `router.Router` interface
2. Store URL parameters in context using `router.URLParamKey`
3. Support standard HTTP methods (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
4. Support middleware via `Use` method

See [examples/custom-router/](examples/custom-router/) for a complete chi router implementation and adapter example.

## Generated Code

SpecWeaver generates two main files:

### `types.go` - Type Definitions

Contains all your data models:

```go
type Pet struct {
    Id         int64      `json:"id"`
    Name       string     `json:"name"`
    Status     PetStatus  `json:"status"`
    BirthDate  *date.Date `json:"birthDate,omitempty"`
    CreatedAt  *time.Time `json:"createdAt,omitempty"`
    Owner      *Owner     `json:"owner,omitempty"`
}

type PetStatus string

const (
    PetStatusAvailable PetStatus = "available"
    PetStatusPending   PetStatus = "pending"
    PetStatusSold      PetStatus = "sold"
)
```

### `server.go` - Server Code

Contains clean, context-based server interface, request/response types, and routing:

```go
// Request types for each operation
type ListPetsRequest struct {
    Limit *int32  // Query parameter
    Tag   *string // Query parameter
}

type CreatePetRequest struct {
    Body NewPet // Request body
}

type GetPetByIdRequest struct {
    PetId int64 // Path parameter
}

// Response interfaces and concrete types
type ListPetsResponse interface {
    StatusCode() int
}

type ListPets200Response struct {
    Body []Pet
}

func (r ListPets200Response) StatusCode() int { return 200 }

type ListPets500Response struct {
    Body Error
}

func (r ListPets500Response) StatusCode() int { return 500 }

// HTTPError for custom error status codes
type HTTPError struct {
    Code    int
    Message string
    Err     error
}

func NewHTTPError(code int, message string) *HTTPError
func NewHTTPErrorf(code int, format string, args ...any) *HTTPError
func WrapHTTPError(code int, err error, message string) *HTTPError

// Server interface with context-based handlers
type Server interface {
    ListPets(ctx context.Context, req ListPetsRequest) (ListPetsResponse, error)
    CreatePet(ctx context.Context, req CreatePetRequest) (CreatePetResponse, error)
    GetPetById(ctx context.Context, req GetPetByIdRequest) (GetPetByIdResponse, error)
    UpdatePet(ctx context.Context, req UpdatePetRequest) (UpdatePetResponse, error)
    DeletePet(ctx context.Context, req DeletePetRequest) (DeletePetResponse, error)
}

// Router setup functions
func NewRouter(si Server) *router.Mux
func ConfigureRouter(r router.Router, si Server)

// Helper functions
func WriteJSON(w http.ResponseWriter, code int, data any) error
func WriteResponse(w http.ResponseWriter, resp interface{ StatusCode() int }) error
func WriteError(w http.ResponseWriter, code int, err error) error
func ReadJSON(r *http.Request, v any) error
```

### Handler Pattern Benefits

1. **Testability**: No HTTP dependencies in business logic
2. **Type Safety**: Compile-time checks for all parameters and responses
3. **Smart Errors**: HTTPError provides custom status codes, defaults to 500
4. **Clean Separation**: HTTP adapter layer handles parsing/serialization
5. **Context Support**: Pass deadlines, cancellation, and request-scoped values

## Examples

### Server Implementation Example

A complete working example is available in `examples/server/`:

```bash
# Run the example server
cd examples/server
go run main.go

# Test the API
curl http://localhost:8080/pets
```

This example demonstrates:
- Complete Server interface implementation with context-based handlers
- Type-safe request structs with path and query parameters
- Response types for different status codes
- HTTPError for custom error status codes
- Clean separation of business logic from HTTP concerns
- Context usage for request-scoped values

### Library Usage Example

See `examples/library/` for examples of using SpecWeaver as a Go module library:

```bash
# Run the library usage example
cd examples/library
go run main.go
```

This example demonstrates:
- Simple one-function generation
- Advanced usage with parser and generator
- Accessing the parsed OpenAPI spec
- Integration patterns for build tools and CI/CD

### Custom Router Example

See `examples/custom-router/` for a complete example of using a custom router (chi) with SpecWeaver:

```bash
# Run the custom router example
cd examples/custom-router
go run main.go

# Test the API
curl http://localhost:8080/pets
```

This example demonstrates:
- Creating an adapter for the chi router
- Implementing the `router.Router` interface
- URL parameter compatibility middleware
- Using chi-specific middleware with SpecWeaver
- Configuring routes with `ConfigureRouter`

## Type Mapping

SpecWeaver intelligently maps OpenAPI types to Go:

| OpenAPI Type | Format | Go Type |
|--------------|--------|---------|
| `string` | - | `string` |
| `string` | `date` | `date.Date` |
| `string` | `date-time` | `time.Time` |
| `integer` | `int32` | `int` |
| `integer` | `int64` | `int64` |
| `number` | `float` | `float32` |
| `number` | `double` | `float64` |
| `boolean` | - | `bool` |
| `array` | - | `[]T` |
| `object` | - | `struct` |
| enum | - | `type` + constants |

## Best Practices

1. **Use Operation IDs**: Define `operationId` in your OpenAPI spec for cleaner handler names
2. **Schema References**: Use `$ref` to reuse schemas and avoid duplication
3. **Descriptions**: Add descriptions to schemas and properties - they become Go comments
4. **Required Fields**: Mark fields as required in the spec for proper validation
5. **Enums**: Use enums for fixed sets of values to get type-safe constants

## OpenAPI 3.x Features Supported

- ✅ Component schemas (objects, arrays, primitives)
- ✅ Schema references (`$ref`)
- ✅ Enums with const generation
- ✅ Required vs optional fields
- ✅ All HTTP methods (GET, POST, PUT, PATCH, DELETE)
- ✅ Path parameters
- ✅ Query parameters
- ✅ Request/response bodies
- ✅ Nested objects
- ✅ Format specifications (date, date-time, int64, float, etc.)

## Project Structure

```
specweaver/
├── cmd/specweaver/     # CLI tool
├── pkg/
│   ├── openapi/        # Custom OpenAPI parser (3.0-3.2 support)
│   ├── parser/         # Parser coordinator
│   ├── router/         # Custom lightweight HTTP router
│   └── generator/      # Code generators
├── examples/           # Example specs and implementations
└── generated/          # Default output directory
```

## Dependencies

### Build Dependencies
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing
- **No external OpenAPI library dependencies** - Custom implementation for maximum control
- **No external routing dependencies** - Custom lightweight router

### Generated Code Dependencies
- **Minimal dependencies** - Generated code uses only:
  - Go standard library
  - `github.com/christopherklint97/specweaver/pkg/router` - Custom router (no external deps)
  - `google.golang.org/genproto/googleapis/type/date` - Only when `format: date` is used

## Development

### Makefile Commands

SpecWeaver includes a comprehensive Makefile for development automation. Run `make help` to see all available targets.

#### Build Commands

```bash
make build         # Build the specweaver binary
make install       # Install to GOPATH/bin
make clean         # Clean build artifacts and generated files
make all           # Run clean, fmt, vet, test, and build
```

#### Testing Commands

```bash
make test          # Run all tests
make test-coverage # Run tests with coverage report (generates HTML)
make test-race     # Run tests with race detector
make test-verbose  # Run tests with verbose output
make test-bench    # Run benchmark tests
```

#### Code Quality Commands

```bash
make fmt           # Format all Go code
make vet           # Run go vet
make lint          # Run golangci-lint (if installed)
make check         # Run fmt, vet, and test (pre-commit checks)
```

#### Code Generation Commands

```bash
make generate          # Generate code from petstore example
make generate-examples # Regenerate code for all examples
```

#### Example Commands

```bash
make example-server        # Run the example server
make example-library       # Run the library usage example
make example-custom-router # Run the custom router example
```

#### Dependency Commands

```bash
make deps          # Download dependencies
make update-deps   # Update dependencies to latest versions
make tidy          # Run go mod tidy
```

#### Development Workflow

```bash
make dev           # Quick development cycle (clean, fmt, vet, build)
make watch         # Watch for changes and rebuild (requires entr)
make version       # Show Go and module versions
```

#### Utility Commands

```bash
make tree          # Show project structure
make size          # Show binary size
make todo          # Show TODO and FIXME comments in code
```

### Development Documentation

See [CLAUDE.md](CLAUDE.md) for detailed development documentation, architecture decisions, and contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Uses [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) for YAML parsing
- Custom OpenAPI 3.x parser supporting versions 3.0 through 3.2
- Custom lightweight HTTP router with middleware support
- Inspired by the OpenAPI Generator project and the Go community's best practices
