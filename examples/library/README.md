# SpecWeaver Library Usage Example

This example demonstrates how to use SpecWeaver as a Go module library in your own projects.

## Overview

SpecWeaver can be used as a library to programmatically generate Go code from OpenAPI specifications. This is useful when you want to integrate code generation into your build process, tools, or applications.

## Installation

To use SpecWeaver as a library in your Go project:

```bash
go get github.com/christopherklint97/specweaver@latest
```

## Simple Usage

The simplest way to use SpecWeaver is with the `Generate` function:

```go
package main

import (
    "log"
    "github.com/christopherklint97/specweaver"
)

func main() {
    err := specweaver.Generate("openapi.yaml", specweaver.Options{
        OutputDir:   "./generated",
        PackageName: "api",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Advanced Usage

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
    fmt.Printf("Generating code for: %s\n", spec.Info.Title)

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

## Running This Example

```bash
# Run the example
go run main.go

# Or build and run
go build -o library-example
./library-example
```

## Integration Scenarios

### Build Tool Integration

```go
// In your build tool (e.g., mage, make, custom script)
func GenerateAPI() error {
    return specweaver.Generate("api/openapi.yaml", specweaver.Options{
        OutputDir:   "internal/api/generated",
        PackageName: "api",
    })
}
```

### CI/CD Pipeline

```go
// In your CI/CD script
func main() {
    specs := []string{
        "specs/users-api.yaml",
        "specs/products-api.yaml",
    }

    for _, spec := range specs {
        err := specweaver.Generate(spec, specweaver.Options{
            OutputDir:   "./generated/" + filepath.Base(spec),
            PackageName: "api",
        })
        if err != nil {
            log.Fatalf("Failed to generate %s: %v", spec, err)
        }
    }
}
```

### Custom Code Generator

```go
// In your custom generator tool
func generateFromMultipleSpecs(specs []string, output string) error {
    for i, specPath := range specs {
        parser := specweaver.NewParser()
        if err := parser.ParseFile(specPath); err != nil {
            return err
        }

        // You can inspect and modify the spec before generation
        spec := parser.GetSpec()
        fmt.Printf("Processing %s (v%s)\n", spec.Info.Title, spec.Info.Version)

        // Generate with custom naming
        pkgName := fmt.Sprintf("api%d", i)
        gen := specweaver.NewGenerator(spec, specweaver.Options{
            OutputDir:   filepath.Join(output, pkgName),
            PackageName: pkgName,
        })

        if err := gen.Generate(); err != nil {
            return err
        }
    }
    return nil
}
```

## API Reference

### Functions

- `Generate(specPath string, opts Options) error` - One-step generation from spec file
- `NewParser() *Parser` - Create a new OpenAPI parser
- `NewGenerator(spec *openapi.Document, opts Options) *Generator` - Create a code generator

### Types

```go
type Options struct {
    OutputDir   string  // Output directory (default: "./generated")
    PackageName string  // Package name (default: "api")
}
```

### Parser Methods

- `ParseFile(filePath string) error` - Parse OpenAPI spec from file
- `GetSpec() *openapi.Document` - Get the parsed specification
- `GetVersion() string` - Get OpenAPI version

### Generator Methods

- `Generate() error` - Generate all code (types and server)

## Supported OpenAPI Versions

- OpenAPI 3.0.x
- OpenAPI 3.1.x
- OpenAPI 3.2.x

Both YAML and JSON formats are supported.

## Learn More

- [Main README](../../README.md) - Project overview and CLI usage
- [CLAUDE.md](../../CLAUDE.md) - Detailed architecture and development documentation
- [Server Example](../server/) - Example of using generated code
