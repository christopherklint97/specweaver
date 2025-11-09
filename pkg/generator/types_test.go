package generator

import (
	"strings"
	"testing"

	"github.com/christopherklint97/specweaver/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTypeGenerator(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
	}

	gen := NewTypeGenerator(spec)
	require.NotNil(t, gen, "Expected type generator to be created")

	assert.Equal(t, spec, gen.spec, "Expected spec to be set")
	assert.NotNil(t, gen.generated, "Expected generated map to be initialized")
}

func TestGenerateSimpleTypes(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"Pet": {
					Value: &openapi.Schema{
						Type:        []string{"object"},
						Description: "A pet object",
						Properties: map[string]*openapi.SchemaRef{
							"name": {
								Value: &openapi.Schema{
									Type:        []string{"string"},
									Description: "Pet name",
								},
							},
							"age": {
								Value: &openapi.Schema{
									Type: []string{"integer"},
								},
							},
						},
						Required: []string{"name"},
					},
				},
			},
		},
	}

	gen := NewTypeGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Check package declaration
	assert.Contains(t, code, "package api", "Expected package declaration")

	// Check imports
	assert.Contains(t, code, "import", "Expected import section")

	// Check type declaration
	assert.Contains(t, code, "type Pet struct", "Expected Pet struct declaration")

	// Check fields
	assert.Contains(t, code, "Name string", "Expected Name field")
	assert.Contains(t, code, "Age int", "Expected Age field")

	// Check JSON tags
	assert.Contains(t, code, `json:"name"`, "Expected JSON tag for name field")
	assert.Contains(t, code, `json:"age,omitempty"`, "Expected JSON tag with omitempty for optional field")

	// Check description comment
	assert.Contains(t, code, "// Pet A pet object", "Expected type description comment")
}

func TestGenerateEnum(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"PetStatus": {
					Value: &openapi.Schema{
						Type:        []string{"string"},
						Description: "Pet status",
						Enum:        []any{"available", "pending", "sold"},
					},
				},
			},
		},
	}

	gen := NewTypeGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Check type declaration
	assert.Contains(t, code, "type PetStatus string", "Expected PetStatus type declaration")

	// Check const declaration
	assert.Contains(t, code, "const (", "Expected const declaration")

	// Check enum constants
	assert.Contains(t, code, "PetStatusAvailable", "Expected PetStatusAvailable constant")
	assert.Contains(t, code, "PetStatusPending", "Expected PetStatusPending constant")
	assert.Contains(t, code, "PetStatusSold", "Expected PetStatusSold constant")

	// Check enum values
	assert.Contains(t, code, `= "available"`, "Expected available enum value")
}

func TestGenerateArrayType(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"PetList": {
					Value: &openapi.Schema{
						Type: []string{"array"},
						Items: &openapi.SchemaRef{
							Value: &openapi.Schema{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		},
	}

	gen := NewTypeGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	assert.Contains(t, code, "type PetList []string", "Expected PetList array type")
}

func TestGenerateNestedObject(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"Owner": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"name": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
						},
						Required: []string{"name"},
					},
				},
				"Pet": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"name": {
								Value: &openapi.Schema{
									Type: []string{"string"},
								},
							},
							"owner": {
								Ref: "#/components/schemas/Owner",
							},
						},
						Required: []string{"name"},
					},
				},
			},
		},
	}

	gen := NewTypeGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Check that both types are generated
	assert.Contains(t, code, "type Owner struct", "Expected Owner struct")
	assert.Contains(t, code, "type Pet struct", "Expected Pet struct")

	// Check that Pet has Owner field with pointer (optional)
	assert.Contains(t, code, "Owner *Owner", "Expected Owner field in Pet with pointer type")
}

func TestGenerateTimeType(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"Event": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"timestamp": {
								Value: &openapi.Schema{
									Type:   []string{"string"},
									Format: "date-time",
								},
							},
						},
						Required: []string{"timestamp"},
					},
				},
			},
		},
	}

	gen := NewTypeGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Check time import
	assert.Contains(t, code, `"time"`, "Expected time import")

	// Check time.Time field (with proper spacing and JSON tag)
	assert.Contains(t, code, "time.Time", "Expected time.Time type in generated code")
	assert.Contains(t, code, "Timestamp", "Expected Timestamp field name")
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"pet", "Pet"},
		{"pet_store", "PetStore"},
		{"pet-store", "PetStore"},
		{"petStore", "PetStore"},
		{"PetStore", "PetStore"},
		{"pet_id", "PetId"},
		{"birth_date", "BirthDate"},
		{"birthDate", "BirthDate"},
		{"user-name", "UserName"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			assert.Equal(t, tt.expected, result, "toPascalCase(%s)", tt.input)
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"pet", []string{"pet"}},
		{"pet_store", []string{"pet", "store"}},
		{"pet-store", []string{"pet", "store"}},
		{"petStore", []string{"pet", "Store"}},
		{"PetStore", []string{"Pet", "Store"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitWords(tt.input)
			assert.Len(t, result, len(tt.expected), "Expected %d words", len(tt.expected))

			for i, word := range result {
				assert.Equal(t, tt.expected[i], word, "Word %d", i)
			}
		})
	}
}

func TestIsPrimitiveType(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"string", true},
		{"int", true},
		{"int32", true},
		{"int64", true},
		{"float32", true},
		{"float64", true},
		{"bool", true},
		{"byte", true},
		{"Pet", false},
		{"[]string", false},
		{"map[string]any", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isPrimitiveType(tt.input)
			assert.Equal(t, tt.expected, result, "isPrimitiveType(%s)", tt.input)
		})
	}
}

func TestResolveType(t *testing.T) {
	gen := NewTypeGenerator(&openapi.Document{})

	tests := []struct {
		name     string
		schema   *openapi.Schema
		expected string
	}{
		{
			name:     "String type",
			schema:   &openapi.Schema{Type: []string{"string"}},
			expected: "string",
		},
		{
			name:     "Integer type",
			schema:   &openapi.Schema{Type: []string{"integer"}},
			expected: "int",
		},
		{
			name:     "Int64 type",
			schema:   &openapi.Schema{Type: []string{"integer"}, Format: "int64"},
			expected: "int64",
		},
		{
			name:     "Number type",
			schema:   &openapi.Schema{Type: []string{"number"}},
			expected: "float64",
		},
		{
			name:     "Float type",
			schema:   &openapi.Schema{Type: []string{"number"}, Format: "float"},
			expected: "float32",
		},
		{
			name:     "Boolean type",
			schema:   &openapi.Schema{Type: []string{"boolean"}},
			expected: "bool",
		},
		{
			name: "Array of strings",
			schema: &openapi.Schema{
				Type: []string{"array"},
				Items: &openapi.SchemaRef{
					Value: &openapi.Schema{
						Type: []string{"string"},
					},
				},
			},
			expected: "[]string",
		},
		{
			name:     "DateTime type",
			schema:   &openapi.Schema{Type: []string{"string"}, Format: "date-time"},
			expected: "time.Time",
		},
		{
			name:     "Nil schema",
			schema:   nil,
			expected: "any",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.resolveType(tt.schema)
			assert.Equal(t, tt.expected, result, "resolveType()")
		})
	}
}

func TestResolveTypeWithRef(t *testing.T) {
	gen := NewTypeGenerator(&openapi.Document{})

	tests := []struct {
		name     string
		ref      *openapi.SchemaRef
		expected string
	}{
		{
			name: "Direct value",
			ref: &openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: []string{"string"},
				},
			},
			expected: "string",
		},
		{
			name: "Component reference",
			ref: &openapi.SchemaRef{
				Ref: "#/components/schemas/Pet",
			},
			expected: "Pet",
		},
		{
			name: "Nested reference",
			ref: &openapi.SchemaRef{
				Ref: "#/components/schemas/PetStore/Owner",
			},
			expected: "Owner",
		},
		{
			name:     "Nil reference",
			ref:      nil,
			expected: "any",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.resolveTypeWithRef(tt.ref)
			assert.Equal(t, tt.expected, result, "resolveTypeWithRef()")
		})
	}
}

func TestGenerateWithNoComponents(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{},
	}

	gen := NewTypeGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Should still generate valid Go code with package and imports
	assert.Contains(t, code, "package api", "Expected package declaration")

	// Should not have any type declarations (besides imports)
	assert.False(t, strings.Contains(code, "type "), "Expected no type declarations for empty spec")
}

func TestGenerateMultipleTypes(t *testing.T) {
	spec := &openapi.Document{
		OpenAPI: "3.1.0",
		Info: &openapi.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.SchemaRef{
				"Pet": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"name": {Value: &openapi.Schema{Type: []string{"string"}}},
						},
					},
				},
				"Owner": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"name": {Value: &openapi.Schema{Type: []string{"string"}}},
						},
					},
				},
				"Store": {
					Value: &openapi.Schema{
						Type: []string{"object"},
						Properties: map[string]*openapi.SchemaRef{
							"address": {Value: &openapi.Schema{Type: []string{"string"}}},
						},
					},
				},
			},
		},
	}

	gen := NewTypeGenerator(spec)
	code, err := gen.Generate()
	require.NoError(t, err, "Generate should not fail")

	// Check all types are generated
	assert.Contains(t, code, "type Pet struct", "Expected Pet struct")
	assert.Contains(t, code, "type Owner struct", "Expected Owner struct")
	assert.Contains(t, code, "type Store struct", "Expected Store struct")
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"Item exists", []string{"a", "b", "c"}, "b", true},
		{"Item does not exist", []string{"a", "b", "c"}, "d", false},
		{"Empty slice", []string{}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result, "contains()")
		})
	}
}
