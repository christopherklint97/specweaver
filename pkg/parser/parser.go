package parser

import (
	"fmt"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// Parser handles OpenAPI specification parsing
type Parser struct {
	spec *openapi.Document
}

// New creates a new Parser instance
func New() *Parser {
	return &Parser{}
}

// ParseFile loads and parses an OpenAPI specification from a file
// Supports OpenAPI 3.0.x, 3.1.x, and 3.2.x
func (p *Parser) ParseFile(filePath string) error {
	spec, err := openapi.Load(filePath)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	p.spec = spec
	return nil
}

// GetSpec returns the parsed OpenAPI specification
func (p *Parser) GetSpec() *openapi.Document {
	return p.spec
}

// GetVersion returns the OpenAPI version from the spec
func (p *Parser) GetVersion() string {
	if p.spec != nil {
		return p.spec.OpenAPI
	}
	return ""
}
