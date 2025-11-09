// Package specweaver provides an OpenAPI 3.x code generator for Go.
//
// SpecWeaver converts OpenAPI 3.x specifications (supporting versions 3.0, 3.1, and 3.2)
// into production-ready Go server code with type-safe handlers and routing.
//
// # Basic Usage
//
//	import "github.com/christopherklint97/specweaver"
//
//	// Generate code from an OpenAPI spec file
//	err := specweaver.Generate("openapi.yaml", specweaver.Options{
//		OutputDir:   "./generated",
//		PackageName: "api",
//	})
//
// # Advanced Usage
//
//	// Use the parser and generator separately for more control
//	parser := specweaver.NewParser()
//	err := parser.ParseFile("openapi.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	generator := specweaver.NewGenerator(parser.GetSpec(), specweaver.Options{
//		OutputDir:   "./api",
//		PackageName: "myapi",
//	})
//	err = generator.Generate()
package specweaver

import (
	"fmt"

	"github.com/christopherklint97/specweaver/pkg/generator"
	"github.com/christopherklint97/specweaver/pkg/openapi"
	"github.com/christopherklint97/specweaver/pkg/parser"
)

// Version is the current version of SpecWeaver
const Version = "0.1.0"

// Options configures the code generation process
type Options struct {
	// OutputDir is the directory where generated code will be written
	// Default: "./generated"
	OutputDir string

	// PackageName is the name of the generated Go package
	// Default: "api"
	PackageName string
}

// Generate is a convenience function that parses an OpenAPI spec file
// and generates Go code in a single call.
//
// Example:
//
//	err := specweaver.Generate("petstore.yaml", specweaver.Options{
//		OutputDir:   "./api",
//		PackageName: "petstore",
//	})
func Generate(specPath string, opts Options) error {
	// Parse the spec
	p := parser.New()
	if err := p.ParseFile(specPath); err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	// Generate code
	config := generator.Config{
		OutputDir:   opts.OutputDir,
		PackageName: opts.PackageName,
	}

	gen := generator.NewGenerator(p.GetSpec(), config)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	return nil
}

// Parser provides access to the OpenAPI parser
type Parser struct {
	p *parser.Parser
}

// NewParser creates a new OpenAPI parser instance
func NewParser() *Parser {
	return &Parser{
		p: parser.New(),
	}
}

// ParseFile loads and parses an OpenAPI specification from a file.
// Supports OpenAPI 3.0.x, 3.1.x, and 3.2.x in both YAML and JSON formats.
func (p *Parser) ParseFile(filePath string) error {
	return p.p.ParseFile(filePath)
}

// GetSpec returns the parsed OpenAPI specification document
func (p *Parser) GetSpec() *openapi.Document {
	return p.p.GetSpec()
}

// GetVersion returns the OpenAPI version from the parsed specification
func (p *Parser) GetVersion() string {
	return p.p.GetVersion()
}

// Generator coordinates the generation of Go code from OpenAPI specs
type Generator struct {
	g *generator.Generator
}

// NewGenerator creates a new code generator instance for the given OpenAPI specification
func NewGenerator(spec *openapi.Document, opts Options) *Generator {
	config := generator.Config{
		OutputDir:   opts.OutputDir,
		PackageName: opts.PackageName,
	}

	return &Generator{
		g: generator.NewGenerator(spec, config),
	}
}

// Generate generates all Go code (types and server) to the configured output directory
func (g *Generator) Generate() error {
	return g.g.Generate()
}
