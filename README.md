# SpecWeaver

> A powerful OpenAPI 3.x Go code generator that weaves your API specifications into production-ready server code.

SpecWeaver automatically generates type-safe Go code from OpenAPI specifications, including data types, HTTP handlers, and routing logic. Uses a custom robust OpenAPI parser supporting versions 3.0.x, 3.1.x, and 3.2.x with no external OpenAPI dependencies.

## Features

- ✨ **Full OpenAPI 3.x Support** - Compatible with OpenAPI 3.0.x, 3.1.x, and 3.2.x
- 🔧 **Custom Robust Parser** - No external OpenAPI library dependencies
- 🎯 **Type-Safe Code** - Generates idiomatic Go structs with proper types
- 🚀 **Production Ready** - Includes error handling, middleware, and best practices
- 📝 **Documentation Preserved** - OpenAPI descriptions become Go comments
- 🔄 **Schema References** - Properly resolves `$ref` to generate correct types
- 🎨 **Idiomatic Go** - Follows Go conventions and best practices (uses `any` instead of `interface{}`)
- ⚡ **Zero Dependencies** - Custom lightweight router, no external dependencies
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

The generator creates a `Server` interface that you need to implement:

```go
package main

import (
    "net/http"
    "github.com/yourorg/yourapp/generated/api"
)

type MyServer struct {
    // Your application state
}

// Implement the interface methods
func (s *MyServer) ListPets(w http.ResponseWriter, r *http.Request) {
    pets := []api.Pet{
        {Id: 1, Name: "Fluffy", Status: api.PetStatusAvailable},
    }
    api.WriteJSON(w, http.StatusOK, pets)
}

func (s *MyServer) CreatePet(w http.ResponseWriter, r *http.Request) {
    var newPet api.NewPet
    if err := api.ReadJSON(r, &newPet); err != nil {
        api.WriteError(w, http.StatusBadRequest, err)
        return
    }
    // Process the new pet...
}

// Implement other methods...
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

## Generated Code

SpecWeaver generates two main files:

### `types.go` - Type Definitions

Contains all your data models:

```go
type Pet struct {
    Id        int64      `json:"id"`
    Name      string     `json:"name"`
    Status    PetStatus  `json:"status"`
    BirthDate *time.Time `json:"birthDate,omitempty"`
    Owner     *Owner     `json:"owner,omitempty"`
}

type PetStatus string

const (
    PetStatusAvailable PetStatus = "available"
    PetStatusPending   PetStatus = "pending"
    PetStatusSold      PetStatus = "sold"
)
```

### `server.go` - Server Code

Contains the interface and routing:

```go
type Server interface {
    ListPets(w http.ResponseWriter, r *http.Request)
    CreatePet(w http.ResponseWriter, r *http.Request)
    GetPetById(w http.ResponseWriter, r *http.Request)
    UpdatePet(w http.ResponseWriter, r *http.Request)
    DeletePet(w http.ResponseWriter, r *http.Request)
}

func NewRouter(si Server) *router.Mux {
    // Creates a configured router with all routes
}
```

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
- Complete Server interface implementation
- Request/response handling
- Query parameter parsing
- Path parameter extraction
- JSON serialization
- Error handling

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

## Type Mapping

SpecWeaver intelligently maps OpenAPI types to Go:

| OpenAPI Type | Format | Go Type |
|--------------|--------|---------|
| `string` | - | `string` |
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
- ✅ Format specifications (date-time, int64, etc.)

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
- **Zero external dependencies** - All generated code uses only Go standard library and the custom router from this project

## Development

See [CLAUDE.md](CLAUDE.md) for detailed development documentation, architecture decisions, and contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Uses [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) for YAML parsing
- Custom OpenAPI 3.x parser supporting versions 3.0 through 3.2
- Custom lightweight HTTP router with middleware support
- Inspired by the OpenAPI Generator project and the Go community's best practices
