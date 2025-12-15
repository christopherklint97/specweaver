package generator

import (
	"sort"

	"github.com/christopherklint97/specweaver/pkg/openapi"
)

// AuthGenerator generates authentication code from OpenAPI security schemes
type AuthGenerator struct {
	spec *openapi.Document
}

// NewAuthGenerator creates a new AuthGenerator instance
func NewAuthGenerator(spec *openapi.Document) *AuthGenerator {
	return &AuthGenerator{
		spec: spec,
	}
}

// Generate generates authentication code
func (g *AuthGenerator) Generate() (string, error) {
	// Build template data
	data := AuthTemplateData{
		Schemes: g.buildSchemes(),
	}

	// Execute template
	return executeTemplate("auth.go.tmpl", data)
}

// buildSchemes builds auth scheme data for the template
func (g *AuthGenerator) buildSchemes() []AuthSchemeData {
	if g.spec.Components == nil || g.spec.Components.SecuritySchemes == nil {
		return nil
	}

	var result []AuthSchemeData

	// Get security scheme names in sorted order
	schemes := make([]string, 0, len(g.spec.Components.SecuritySchemes))
	for name := range g.spec.Components.SecuritySchemes {
		schemes = append(schemes, name)
	}
	sort.Strings(schemes)

	for _, name := range schemes {
		scheme := g.spec.Components.SecuritySchemes[name]
		if scheme == nil {
			continue
		}

		result = append(result, AuthSchemeData{
			Name:       name,
			MethodName: toPascalCase(name),
			Type:       scheme.Type,
			Scheme:     scheme.Scheme,
		})
	}

	return result
}
