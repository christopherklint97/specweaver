package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

// Generator coordinates the generation of Go code from OpenAPI specs
type Generator struct {
	spec       *openapi3.T
	outputDir  string
	packageName string
}

// Config holds generator configuration
type Config struct {
	OutputDir   string
	PackageName string
}

// NewGenerator creates a new Generator instance
func NewGenerator(spec *openapi3.T, config Config) *Generator {
	if config.PackageName == "" {
		config.PackageName = "api"
	}
	if config.OutputDir == "" {
		config.OutputDir = "./generated"
	}

	return &Generator{
		spec:        spec,
		outputDir:   config.OutputDir,
		packageName: config.PackageName,
	}
}

// Generate generates all code (types and server)
func (g *Generator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate types
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	// Generate server
	if err := g.generateServer(); err != nil {
		return fmt.Errorf("failed to generate server: %w", err)
	}

	fmt.Printf("✓ Code generated successfully in %s/\n", g.outputDir)
	fmt.Printf("  - types.go: Type definitions\n")
	fmt.Printf("  - server.go: Server handlers and router\n")

	return nil
}

// generateTypes generates type definitions
func (g *Generator) generateTypes() error {
	typeGen := NewTypeGenerator(g.spec)
	code, err := typeGen.Generate()
	if err != nil {
		return err
	}

	outputPath := filepath.Join(g.outputDir, "types.go")
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write types file: %w", err)
	}

	return nil
}

// generateServer generates server code
func (g *Generator) generateServer() error {
	serverGen := NewServerGenerator(g.spec)
	code, err := serverGen.Generate()
	if err != nil {
		return err
	}

	outputPath := filepath.Join(g.outputDir, "server.go")
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write server file: %w", err)
	}

	return nil
}
