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
├── cmd/
│   └── specweaver/          # CLI entry point
│       └── main.go          # Command-line interface
├── pkg/
│   ├── parser/              # OpenAPI parser
│   │   └── parser.go        # Loads and validates OpenAPI specs
│   ├── generator/           # Code generators
│   │   ├── generator.go     # Main generator coordinator
│   │   ├── types.go         # Type/struct generation
│   │   └── server.go        # Server code generation
├── examples/
│   ├── petstore.yaml        # Example OpenAPI spec
│   └── server/              # Example server implementation
│       ├── main.go          # Reference implementation
│       └── api/             # Generated code (copied for example)
├── generated/               # Default output directory
├── go.mod
└── README.md
```

### Key Components

#### 1. Parser (`pkg/parser/parser.go`)
- **Purpose**: Load and validate OpenAPI specifications
- **Library**: Uses `github.com/getkin/kin-openapi/openapi3` for spec parsing
- **Features**:
  - Supports OpenAPI 3.x (tested with 3.1.0, compatible with 3.2)
  - Validates specs before code generation
  - Handles external references

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

#### 3. Server Generator (`pkg/generator/server.go`)
- **Purpose**: Generate HTTP server code
- **Router**: Uses `github.com/go-chi/chi/v5` (lightweight, idiomatic)
- **Generated Components**:
  - `ServerInterface`: Interface with all handler methods
  - `NewRouter()`: Function to create configured router
  - Helper functions:
    - `WriteJSON()`: Write JSON responses
    - `WriteError()`: Write error responses
    - `ReadJSON()`: Parse JSON request bodies
- **Middleware**: Includes logging, recovery, request ID, and real IP

#### 4. Main Generator (`pkg/generator/generator.go`)
- **Purpose**: Coordinate the generation process
- **Responsibilities**:
  - Create output directory
  - Orchestrate type and server generation
  - Write generated code to files

#### 5. CLI (`cmd/specweaver/main.go`)
- **Flags**:
  - `-spec`: Path to OpenAPI spec file (required)
  - `-output`: Output directory (default: `./generated`)
  - `-package`: Package name (default: `api`)
  - `-version`: Show version information

## Implementation Details

### Type Resolution

The type resolution system handles several cases:

1. **Primitive Types**: Direct mapping (string, int, bool, etc.)
2. **Schema References**: Extracts type name from `$ref` paths
   - Example: `#/components/schemas/Pet` → `Pet`
3. **Arrays**: Resolves item types recursively
4. **Objects**: Generates structs or uses `map[string]interface{}`
5. **Enums**: Creates string types with const declarations

### OpenAPI 3.x Compatibility

The generator handles the OpenAPI 3.x type system where `schema.Type` is `*openapi3.Types` (a slice of strings):
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

## Usage Example

### 1. Generate Code

```bash
./specweaver -spec examples/petstore.yaml -output ./generated
```

### 2. Implement the Interface

```go
type MyServer struct {
    // Your state here
}

func (s *MyServer) ListPets(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

// Implement other methods...
```

### 3. Start the Server

```go
server := &MyServer{}
router := api.NewRouter(server)
http.ListenAndServe(":8080", router)
```

See `examples/server/main.go` for a complete implementation.

## Dependencies

### Runtime Dependencies
- `github.com/getkin/kin-openapi/openapi3` - OpenAPI parsing
- `github.com/go-chi/chi/v5` - HTTP routing (in generated code)

### Generated Code Dependencies
Generated code requires:
- `encoding/json` - JSON serialization
- `net/http` - HTTP server
- `io` - Request body reading
- `time` - DateTime handling
- `github.com/go-chi/chi/v5` - Router

## Development History

### Initial Implementation (2025-11-08)

1. **Project Setup**
   - Created Go module structure
   - Set up package organization
   - Added dependencies

2. **Parser Implementation**
   - Integrated kin-openapi library
   - Added validation support
   - Implemented file loading

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

1. **Type System Compatibility**: Updated to work with OpenAPI 3.x type system (slice-based types)
2. **Naming Improvements**: Enhanced PascalCase conversion to handle compound words correctly
3. **Reference Resolution**: Proper handling of `$ref` to generate correct type names instead of `map[string]interface{}`
4. **Code Quality**: Added comprehensive comments and documentation

## Future Enhancements (Potential)

1. **Request/Response Validation**: Add OpenAPI validation middleware
2. **Client Generation**: Generate Go client code
3. **Custom Templates**: Allow users to customize generated code
4. **OpenAPI Extensions**: Support vendor extensions (x-*)
5. **Authentication**: Generate auth middleware from security schemes
6. **Testing**: Generate test stubs
7. **Documentation**: Generate Markdown docs from spec

## Testing

The generator has been tested with:
- OpenAPI 3.1.0 specification (examples/petstore.yaml)
- Various schema types (objects, arrays, enums, primitives)
- Schema references
- Optional and required fields
- Multiple HTTP methods and paths

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
