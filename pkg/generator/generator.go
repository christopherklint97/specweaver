package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// Generator coordinates the generation of Go code from OpenAPI specs
type Generator struct {
	spec       *openapi.Document
	outputDir  string
	packageName string
}

// Config holds generator configuration
type Config struct {
	OutputDir   string
	PackageName string
}

// NewGenerator creates a new Generator instance
func NewGenerator(spec *openapi.Document, config Config) *Generator {
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

// Generate generates all code (types, server, and auth)
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

	// Generate auth (if security schemes are defined)
	if err := g.generateAuth(); err != nil {
		return fmt.Errorf("failed to generate auth: %w", err)
	}

	// Generate webhooks (if webhooks are defined)
	if err := g.generateWebhooks(); err != nil {
		return fmt.Errorf("failed to generate webhooks: %w", err)
	}

	fmt.Printf("✓ Code generated successfully in %s/\n", g.outputDir)
	fmt.Printf("  - types.go: Type definitions\n")
	fmt.Printf("  - server.go: Server handlers and router\n")
	if g.hasSecuritySchemes() {
		fmt.Printf("  - auth.go: Authentication middleware and types\n")
	}
	if g.hasWebhooks() {
		fmt.Printf("  - webhooks.go: Webhook client and sender functions\n")
	}

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

// generateAuth generates authentication code
func (g *Generator) generateAuth() error {
	// Only generate auth.go if there are security schemes
	if !g.hasSecuritySchemes() {
		return nil
	}

	authGen := NewAuthGenerator(g.spec)
	code, err := authGen.Generate()
	if err != nil {
		return err
	}

	outputPath := filepath.Join(g.outputDir, "auth.go")
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// generateWebhooks generates webhook client code
func (g *Generator) generateWebhooks() error {
	// Only generate webhooks.go if there are webhooks
	if !g.hasWebhooks() {
		return nil
	}

	webhookGen := NewWebhookGenerator(g.spec)
	code, err := webhookGen.Generate()
	if err != nil {
		return err
	}

	// Only write the file if there's content
	if code == "" {
		return nil
	}

	outputPath := filepath.Join(g.outputDir, "webhooks.go")
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write webhooks file: %w", err)
	}

	return nil
}

// hasSecuritySchemes checks if the spec defines any security schemes
func (g *Generator) hasSecuritySchemes() bool {
	return g.spec.Components != nil &&
		g.spec.Components.SecuritySchemes != nil &&
		len(g.spec.Components.SecuritySchemes) > 0
}

// hasWebhooks checks if the spec defines any webhooks
func (g *Generator) hasWebhooks() bool {
	return g.spec.Webhooks != nil && len(g.spec.Webhooks) > 0
}
