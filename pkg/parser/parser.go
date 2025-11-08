package parser

import (
	"context"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

// Parser handles OpenAPI specification parsing
type Parser struct {
	spec *openapi3.T
}

// New creates a new Parser instance
func New() *Parser {
	return &Parser{}
}

// ParseFile loads and parses an OpenAPI specification from a file
func (p *Parser) ParseFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromData(data)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the specification
	if err := spec.Validate(context.Background()); err != nil {
		return fmt.Errorf("OpenAPI spec validation failed: %w", err)
	}

	p.spec = spec
	return nil
}

// GetSpec returns the parsed OpenAPI specification
func (p *Parser) GetSpec() *openapi3.T {
	return p.spec
}

// GetVersion returns the OpenAPI version from the spec
func (p *Parser) GetVersion() string {
	if p.spec != nil {
		return p.spec.OpenAPI
	}
	return ""
}
