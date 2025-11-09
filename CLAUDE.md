# SpecWeaver - Development Documentation

This file documents the development history and architecture of SpecWeaver, an OpenAPI Go code generator.

## Project Overview

**SpecWeaver** is a code generator that converts OpenAPI 3.x specifications (supporting up to version 3.2) into production-ready Go server code. It generates:
- Type-safe Go structs from OpenAPI schemas
- HTTP server handlers and routing code
- Helper functions for JSON serialization and error handling

## Architecture

### Directory Structure

```
specweaver/
├── specweaver.go            # Public API for library usage
├── cmd/
│   └── specweaver/          # CLI entry point
│       └── main.go          # Command-line interface
├── pkg/
│   ├── openapi/             # Custom OpenAPI parser
│   │   ├── spec.go          # OpenAPI data structures
│   │   ├── parser.go        # YAML/JSON parsing and validation
│   │   └── unmarshal.go     # Custom unmarshaling for type compatibility
│   ├── parser/              # Parser coordinator
│   │   └── parser.go        # High-level parser interface
│   ├── router/              # Custom HTTP router
│   │   ├── router.go        # Router implementation
│   │   └── middleware.go    # Middleware (Logger, Recoverer, etc.)
│   ├── generator/           # Code generators
│   │   ├── generator.go     # Main generator coordinator
│   │   ├── types.go         # Type/struct generation
│   │   └── server.go        # Server code generation
├── examples/
│   ├── petstore.yaml        # Example OpenAPI spec
│   ├── server/              # Example server implementation
│   │   ├── main.go          # Reference implementation
│   │   └── api/             # Generated code (copied for example)
│   ├── library/             # Example of using SpecWeaver as a library
│   │   ├── main.go          # Library usage examples
│   │   ├── go.mod           # Module file for the example
│   │   └── README.md        # Library usage documentation
│   └── custom-router/       # Example using custom router (chi)
│       ├── main.go          # Server using chi router
│       ├── chi_adapter.go   # Chi router adapter
│       ├── go.mod           # Module file
│       └── README.md        # Custom router documentation
├── generated/               # Default output directory
├── go.mod
└── README.md
```

### Key Components

#### 1. Parser (`pkg/parser/parser.go` and `pkg/openapi/`)
- **Purpose**: Load and validate OpenAPI specifications
- **Implementation**: Custom robust OpenAPI parser (no external dependencies)
- **Features**:
  - Supports OpenAPI 3.0.x, 3.1.x, and 3.2.x
  - Validates specs before code generation
  - Handles internal references ($ref resolution)
  - Supports both YAML and JSON formats
  - Type normalization for compatibility across versions

#### 2. Type Generator (`pkg/generator/types.go`)
- **Purpose**: Convert OpenAPI schemas to Go types
- **Features**:
  - Generates structs from object schemas
  - Handles enums as string types with constants
  - Resolves schema references (`$ref`)
  - Maps OpenAPI types to idiomatic Go types:
    - `string` with `format: date-time` → `time.Time`
    - `integer` with `format: int64` → `int64`
    - `number` → `float64` (or `float32` with `format: float`)
  - Adds JSON tags with `omitempty` for optional fields
  - Uses pointers for optional non-primitive types
  - Preserves descriptions as Go comments

**Naming Conventions**:
- Converts snake_case, kebab-case, and camelCase to PascalCase
- Handles acronyms and compound words properly
- Example: `pet-status` → `PetStatus`, `birthDate` → `BirthDate`

#### 3. Router (`pkg/router/`)
- **Purpose**: Provide a default HTTP router and interface for custom routers
- **Built-in Router Features**:
  - HTTP method routing (GET, POST, PUT, DELETE, PATCH, etc.)
  - Path parameter support (`/pets/{id}`)
  - Middleware support
  - Zero external dependencies
  - Lightweight and fast
- **Built-in Middleware**:
  - `Logger`: Request logging
  - `Recoverer`: Panic recovery
  - `RequestID`: Request ID generation
  - `RealIP`: Real IP extraction from headers
- **Custom Router Support**:
  - Defines `router.Router` interface for pluggable routers
  - Any router implementing the interface can be used
  - Compatible with popular routers (chi, gorilla/mux, httprouter, etc.)
  - URL parameters must be stored in context using `router.URLParamKey`
  - See `examples/custom-router/` for implementation examples

#### 4. Server Generator (`pkg/generator/server.go`)
- **Purpose**: Generate HTTP server code with clean, testable patterns
- **Router**: Uses custom `pkg/router` (zero dependencies)
- **Handler Pattern**: `func(ctx context.Context, req) (res, error)`
  - Request structs contain path params, query params, headers, and body
  - Response types map to OpenAPI response definitions
  - Smart error handling with HTTPError for custom status codes
- **Generated Components**:
  - **Request Types**: One per operation (e.g., `ListPetsRequest`)
    - Path parameters as required fields
    - Query parameters as optional pointer fields
    - Request body when applicable
  - **Response Types**: Interface with concrete types per status code
    - Interface (e.g., `ListPetsResponse`)
    - Concrete types (e.g., `ListPets200Response`, `ListPets500Response`)
    - Each implements `StatusCode() int` method
  - **HTTPError**: Custom error type with status code
    - `NewHTTPError(code, message)`: Create error with specific status
    - `NewHTTPErrorf(code, format, args...)`: Create with formatted message
    - `WrapHTTPError(code, err, message)`: Wrap existing error
    - Errors default to 500 unless HTTPError is used
  - **Server Interface**: Clean business logic methods
  - **ServerWrapper**: HTTP adapter that bridges to handler methods
  - **ConfigureRouter(r, si)**: Configures any router with generated routes
  - **NewRouter(si)**: Convenience function using built-in router
  - Helper functions:
    - `WriteJSON()`: Write JSON responses
    - `WriteResponse()`: Write typed response (handles status codes)
    - `WriteError()`: Write error responses
    - `ReadJSON()`: Parse JSON request bodies
- **Middleware**: Includes logging, recovery, request ID, and real IP

#### 5. Main Generator (`pkg/generator/generator.go`)
- **Purpose**: Coordinate the generation process
- **Responsibilities**:
  - Create output directory
  - Orchestrate type and server generation
  - Write generated code to files

#### 6. CLI (`cmd/specweaver/main.go`)
- **Flags**:
  - `-spec`: Path to OpenAPI spec file (required)
  - `-output`: Output directory (default: `./generated`)
  - `-package`: Package name (default: `api`)
  - `-version`: Show version information

#### 7. Public API (`specweaver.go`)
- **Purpose**: Provide a clean, high-level API for library usage
- **Features**:
  - Simple one-function generation with `Generate()`
  - Advanced usage with separate `Parser` and `Generator` types
  - Wraps internal packages for a cleaner public interface
  - Comprehensive godoc documentation
- **Use Cases**:
  - Build tool integration (mage, make, custom scripts)
  - CI/CD pipeline automation
  - Dynamic API generation tools
  - Multi-spec batch processing

## Implementation Details

### Type Resolution

The type resolution system handles several cases:

1. **Primitive Types**: Direct mapping (string, int, bool, etc.)
2. **Schema References**: Extracts type name from `$ref` paths
   - Example: `#/components/schemas/Pet` → `Pet`
3. **Arrays**: Resolves item types recursively
4. **Objects**: Generates structs or uses `map[string]any`
5. **Enums**: Creates string types with const declarations

### OpenAPI 3.x Compatibility

The generator uses a custom OpenAPI parser that supports all versions:
- **OpenAPI 3.0.x**: Handles `type` as a single string value
- **OpenAPI 3.1.x**: Handles `type` as an array of strings (JSON Schema 2020-12)
- **OpenAPI 3.2.x**: Full support for the latest specification
- Custom unmarshaling logic normalizes type fields across versions
- Uses `getSchemaType()` helper to safely extract the primary type
- Handles schemas without explicit types (inferred from properties)

### Code Generation Best Practices

1. **Idiomatic Go**:
   - Proper error handling
   - Exported types and functions
   - JSON tags for serialization
   - Interface-based design

2. **Type Safety**:
   - Strong typing for request/response bodies
   - Enum constants instead of strings
   - Pointer types for optional fields

3. **Documentation**:
   - Preserves OpenAPI descriptions as comments
   - Comments for all exported types and functions

## Usage

SpecWeaver can be used in two ways:

1. **As a CLI tool** - For manual code generation and development
2. **As a Go module library** - For programmatic integration in tools, build systems, and CI/CD

### CLI Usage

#### Generate Code

```bash
./specweaver -spec examples/petstore.yaml -output ./generated
```

### Library Usage

SpecWeaver can be imported as a Go module and used programmatically in your applications.

#### Installation

```bash
go get github.com/christopherklint97/specweaver@latest
```

#### Simple API

The simplest way to use SpecWeaver as a library:

```go
import "github.com/christopherklint97/specweaver"

// Generate code with a single function call
err := specweaver.Generate("openapi.yaml", specweaver.Options{
    OutputDir:   "./generated",
    PackageName: "api",
})
```

#### Advanced API

For more control, use the parser and generator separately:

```go
import "github.com/christopherklint97/specweaver"

// Parse the OpenAPI spec
parser := specweaver.NewParser()
err := parser.ParseFile("openapi.yaml")
if err != nil {
    return err
}

// Access the parsed specification
spec := parser.GetSpec()
fmt.Printf("Generating for: %s v%s\n", spec.Info.Title, spec.Info.Version)

// Generate code
generator := specweaver.NewGenerator(spec, specweaver.Options{
    OutputDir:   "./api",
    PackageName: "myapi",
})
err = generator.Generate()
```

#### Public API Reference

The root package (`specweaver.go`) provides a clean public API:

**Functions:**
- `Generate(specPath string, opts Options) error` - One-step generation
- `NewParser() *Parser` - Create OpenAPI parser
- `NewGenerator(spec *openapi.Document, opts Options) *Generator` - Create code generator

**Types:**
- `Options` - Configuration for code generation
  - `OutputDir` - Output directory (default: "./generated")
  - `PackageName` - Package name (default: "api")
- `Parser` - OpenAPI specification parser
- `Generator` - Code generator

**Integration Examples:**

See `examples/library/` for complete examples including:
- Simple one-function generation
- Advanced usage with parser and generator
- Build tool integration patterns
- CI/CD pipeline integration
- Multi-spec batch processing

### 2. Implement the Server Interface

The generated code uses a clean, testable pattern with `context.Context`, request structs, and response types:

```go
type MyServer struct {
    // Your state here
}

// Implement handlers with clean signature
func (s *MyServer) ListPets(ctx context.Context, req api.ListPetsRequest) (api.ListPetsResponse, error) {
    // Access query parameters
    limit := 20
    if req.Limit != nil {
        limit = int(*req.Limit)
    }

    // Business logic here
    pets := []api.Pet{...}

    // Return typed response
    return api.ListPets200Response{Body: pets}, nil
}

func (s *MyServer) CreatePet(ctx context.Context, req api.CreatePetRequest) (api.CreatePetResponse, error) {
    // Validation
    if req.Body.Name == "" {
        // Return custom error with status code
        return nil, api.NewHTTPError(http.StatusBadRequest, "name is required")
    }

    // Business logic
    pet := createPet(req.Body)

    // Return 201 Created
    return api.CreatePet201Response{Body: pet}, nil
}

func (s *MyServer) GetPetById(ctx context.Context, req api.GetPetByIdRequest) (api.GetPetByIdResponse, error) {
    pet, exists := s.findPet(req.PetId)
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
```

### 3. Start the Server

#### Using the Built-in Router

```go
server := &MyServer{}
router := api.NewRouter(server)
http.ListenAndServe(":8080", router)
```

#### Using a Custom Router

SpecWeaver supports any router that implements the `router.Router` interface:

```go
// Create your custom router (e.g., chi)
customRouter := NewChiAdapter()

// Add your middleware
customRouter.Use(middleware.Logger)
customRouter.Use(ChiURLParamMiddleware) // Required for URL params

// Configure with SpecWeaver routes
api.ConfigureRouter(customRouter, server)

// Start server
http.ListenAndServe(":8080", customRouter)
```

**Custom Router Requirements:**
1. Implement `router.Router` interface (http.Handler + routing methods)
2. Store URL parameters in context using `router.URLParamKey`
3. Support standard HTTP methods (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
4. Support middleware via `Use` method

See `examples/custom-router/` for a complete chi router implementation.

### Benefits of the New Pattern

1. **Testability**: No HTTP dependencies in business logic
2. **Type Safety**: Compile-time checks for all parameters and responses
3. **Smart Errors**: HTTPError provides custom status codes, defaults to 500
4. **Clean Separation**: HTTP adapter layer (ServerWrapper) handles parsing/serialization
5. **Context Support**: Pass deadlines, cancellation, and request-scoped values

See `examples/server/main.go` for a complete implementation.

## Dependencies

### Runtime Dependencies
- `gopkg.in/yaml.v3` - YAML parsing
- No external OpenAPI library dependencies (custom implementation)
- No external routing dependencies (custom router)

### Generated Code Dependencies
Generated code requires:
- `encoding/json` - JSON serialization
- `net/http` - HTTP server
- `io` - Request body reading
- `time` - DateTime handling
- `github.com/christopherklint97/specweaver/pkg/router` - Custom HTTP router (no external dependencies)

## Development History

### Initial Implementation (2025-11-08)

1. **Project Setup**
   - Created Go module structure
   - Set up package organization
   - Added dependencies

2. **Parser Implementation**
   - Created custom OpenAPI parser (`pkg/openapi/`)
   - Support for OpenAPI 3.0.x, 3.1.x, and 3.2.x
   - YAML and JSON format support
   - Reference resolution ($ref handling)
   - Type normalization across versions
   - Validation support

3. **Type Generator**
   - Initial schema to struct conversion
   - Added enum support
   - Implemented reference resolution
   - Fixed naming conventions for proper PascalCase

4. **Server Generator**
   - Created ServerInterface design
   - Implemented router generation
   - Added helper functions
   - Integrated chi middleware

5. **CLI Tool**
   - Command-line flag parsing
   - User-friendly output
   - Error handling

6. **Testing & Examples**
   - Created pet store example spec
   - Built reference implementation
   - Validated generated code quality

### Key Improvements Made

1. **Custom OpenAPI Parser**: Replaced external dependency with robust custom implementation
   - Supports OpenAPI 3.0.x, 3.1.x, and 3.2.x
   - No external OpenAPI library dependencies
   - Better control over parsing and validation
2. **Type System Compatibility**: Handles both single type (3.0) and array types (3.1+)
3. **Naming Improvements**: Enhanced PascalCase conversion to handle compound words correctly
4. **Reference Resolution**: Proper handling of `$ref` to generate correct type names instead of `map[string]any`
5. **Code Quality**: Added comprehensive comments and documentation
6. **Modern Go**: Using `any` instead of `interface{}` throughout
7. **Custom Router**: Replaced chi router with lightweight custom implementation
   - Zero external dependencies
   - Full middleware support
   - Path parameter routing

### Server Method Pattern Refactoring (2025-11-08)

**Major architectural improvement** to the server interface design:

#### Changes Made:

1. **New Handler Signature**:
   - **Before**: `func(w http.ResponseWriter, r *http.Request)`
   - **After**: `func(ctx context.Context, req XRequest) (XResponse, error)`

2. **Request Structs**: Generated for each operation
   - Contains path parameters (required fields)
   - Contains query parameters (optional pointer fields)
   - Contains request body when applicable
   - Example: `ListPetsRequest`, `CreatePetRequest`

3. **Response Types**: Interface-based with concrete implementations
   - Response interface per operation (e.g., `ListPetsResponse`)
   - Concrete types for each status code (e.g., `ListPets200Response`, `ListPets500Response`)
   - Each concrete type implements `StatusCode() int` method
   - Allows type-safe response handling

4. **HTTPError Type**: Smart error handling
   - `NewHTTPError(code, message)`: Create error with specific HTTP status code
   - `NewHTTPErrorf(code, format, args...)`: Create with formatted message
   - `WrapHTTPError(code, err, message)`: Wrap existing error with status code
   - Default behavior: errors return 500 unless HTTPError is used
   - Error handling in ServerWrapper checks for HTTPError type

5. **ServerWrapper**: HTTP adapter layer
   - Bridges HTTP requests to clean handler methods
   - Parses path parameters, query parameters, and request body
   - Calls handler with typed request struct
   - Handles response serialization based on type
   - Manages error handling (HTTPError vs standard errors)

#### Benefits:

- **Testability**: Business logic has no HTTP dependencies
- **Type Safety**: Compile-time checks for parameters and responses
- **Cleaner Code**: Separation of HTTP concerns from business logic
- **Better Errors**: Explicit control over HTTP status codes
- **Context Support**: Native support for deadlines, cancellation, request-scoped values
- **OpenAPI Alignment**: Response types directly map to spec definitions

## Future Enhancements (Potential)

1. **Request/Response Validation**: Add OpenAPI validation middleware
2. **Client Generation**: Generate Go client code
3. **Custom Templates**: Allow users to customize generated code
4. **OpenAPI Extensions**: Support vendor extensions (x-*)
5. **Authentication**: Generate auth middleware from security schemes
6. **Testing**: Generate test stubs
7. **Documentation**: Generate Markdown docs from spec

## Testing Guidelines

### When to Add Tests

**IMPORTANT**: Tests should be considered and added **AFTER** implementing all changes, not before or during development.

Once your implementation is complete, consider whether tests would provide actual value. Only add tests if they meaningfully verify:

1. **Happy Paths**: Normal, expected use cases that demonstrate the feature works as intended
2. **Logic Verification**: Core business logic and important behavior that must remain correct
3. **Reasonable Edge Cases**: Boundary conditions, error handling, and important corner cases

**Do NOT add tests for**:
- Trivial getters/setters
- Simple pass-through functions
- Code that's unlikely to break or has minimal logic
- Every possible permutation (focus on meaningful cases)

### Testing Best Practices

**Always use `testify/assert` and `testify/require` for all test assertions.**

**NEVER use these standard Go testing methods:**
- ❌ `t.Error()` or `t.Errorf()`
- ❌ `t.Fatal()` or `t.Fatalf()`

**Instead, use:**
- ✅ `assert.Equal(t, expected, actual)` - For non-critical checks
- ✅ `assert.NotNil(t, value)` - Verify non-nil values
- ✅ `assert.NoError(t, err)` - Check for no errors
- ✅ `assert.Contains(t, haystack, needle)` - String/slice contains
- ✅ `assert.True(t, condition)` / `assert.False(t, condition)` - Boolean checks
- ✅ `require.NoError(t, err)` - Critical checks that should stop test execution
- ✅ `require.NotNil(t, value)` - Critical nil checks

**Key Difference**:
- Use `require.*` for critical checks where test cannot continue if assertion fails (e.g., nil pointer checks, fatal errors)
- Use `assert.*` for non-critical checks where test can continue even if assertion fails

### Example Test Structure

```go
func TestFeatureName(t *testing.T) {
    // Setup
    input := setupTestData()

    // Execute
    result, err := FeatureFunction(input)

    // Assert - critical checks first (use require)
    require.NoError(t, err, "Feature should not return error")
    require.NotNil(t, result, "Result should not be nil")

    // Assert - detailed checks (use assert)
    assert.Equal(t, expectedValue, result.Field, "Field should match expected")
    assert.Len(t, result.Items, 3, "Should have 3 items")
    assert.Contains(t, result.Message, "success", "Message should contain success")
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run specific package tests
go test -v ./pkg/generator/

# Run specific test
go test -v -run TestFeatureName ./pkg/generator/
```

### CI Integration

All tests automatically run on:
- Every pull request
- Every push to main branch
- Go version: 1.25.4

The CI pipeline includes:
- Unit tests with race detection
- Coverage reporting
- Linting with golangci-lint
- Build verification

## Testing

The generator has been tested with:
- OpenAPI 3.1.0 specification (examples/petstore.yaml)
- Compatible with OpenAPI 3.0.x, 3.1.x, and 3.2.x
- Various schema types (objects, arrays, enums, primitives)
- Schema references ($ref resolution)
- Optional and required fields
- Multiple HTTP methods and paths
- Both YAML and JSON input formats

## Contributing

When making changes:
1. Follow Go best practices and idioms
2. Maintain backward compatibility
3. Update this file with significant changes
4. Test with various OpenAPI specs
5. Ensure generated code compiles and runs

## Version History

- **v0.1.0** (2025-11-08): Initial release
  - OpenAPI 3.x parser
  - Type generation
  - Server generation
  - CLI tool
  - Example implementation
