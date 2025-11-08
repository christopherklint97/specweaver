package generator

import (
	"strings"
	"testing"

	"github.com/christopherklint97/specweaver/pkg/openapi"
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
	if gen == nil {
		t.Fatal("Expected type generator to be created")
	}

	if gen.spec != spec {
		t.Error("Expected spec to be set")
	}

	if gen.generated == nil {
		t.Error("Expected generated map to be initialized")
	}
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
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check package declaration
	if !strings.Contains(code, "package api") {
		t.Error("Expected package declaration")
	}

	// Check imports
	if !strings.Contains(code, "import") {
		t.Error("Expected import section")
	}

	// Check type declaration
	if !strings.Contains(code, "type Pet struct") {
		t.Error("Expected Pet struct declaration")
	}

	// Check fields
	if !strings.Contains(code, "Name string") {
		t.Error("Expected Name field")
	}

	if !strings.Contains(code, "Age int") {
		t.Error("Expected Age field")
	}

	// Check JSON tags
	if !strings.Contains(code, `json:"name"`) {
		t.Error("Expected JSON tag for name field")
	}

	if !strings.Contains(code, `json:"age,omitempty"`) {
		t.Error("Expected JSON tag with omitempty for optional field")
	}

	// Check description comment
	if !strings.Contains(code, "// Pet A pet object") {
		t.Error("Expected type description comment")
	}
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
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check type declaration
	if !strings.Contains(code, "type PetStatus string") {
		t.Error("Expected PetStatus type declaration")
	}

	// Check const declaration
	if !strings.Contains(code, "const (") {
		t.Error("Expected const declaration")
	}

	// Check enum constants
	if !strings.Contains(code, "PetStatusAvailable") {
		t.Error("Expected PetStatusAvailable constant")
	}

	if !strings.Contains(code, "PetStatusPending") {
		t.Error("Expected PetStatusPending constant")
	}

	if !strings.Contains(code, "PetStatusSold") {
		t.Error("Expected PetStatusSold constant")
	}

	// Check enum values
	if !strings.Contains(code, `= "available"`) {
		t.Error("Expected available enum value")
	}
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
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(code, "type PetList []string") {
		t.Error("Expected PetList array type")
	}
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
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that both types are generated
	if !strings.Contains(code, "type Owner struct") {
		t.Error("Expected Owner struct")
	}

	if !strings.Contains(code, "type Pet struct") {
		t.Error("Expected Pet struct")
	}

	// Check that Pet has Owner field with pointer (optional)
	if !strings.Contains(code, "Owner *Owner") {
		t.Error("Expected Owner field in Pet with pointer type")
	}
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
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check time import
	if !strings.Contains(code, `"time"`) {
		t.Error("Expected time import")
	}

	// Check time.Time field (with proper spacing and JSON tag)
	if !strings.Contains(code, "time.Time") {
		t.Error("Expected time.Time type in generated code")
	}

	if !strings.Contains(code, "Timestamp") {
		t.Error("Expected Timestamp field name")
	}
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
			if result != tt.expected {
				t.Errorf("toPascalCase(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
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
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d words, got %d", len(tt.expected), len(result))
			}

			for i, word := range result {
				if word != tt.expected[i] {
					t.Errorf("Word %d: expected %s, got %s", i, tt.expected[i], word)
				}
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
			if result != tt.expected {
				t.Errorf("isPrimitiveType(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
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
			if result != tt.expected {
				t.Errorf("resolveType() = %s, expected %s", result, tt.expected)
			}
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
			if result != tt.expected {
				t.Errorf("resolveTypeWithRef() = %s, expected %s", result, tt.expected)
			}
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
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should still generate valid Go code with package and imports
	if !strings.Contains(code, "package api") {
		t.Error("Expected package declaration")
	}

	// Should not have any type declarations (besides imports)
	if strings.Contains(code, "type ") {
		t.Error("Expected no type declarations for empty spec")
	}
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
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check all types are generated
	if !strings.Contains(code, "type Pet struct") {
		t.Error("Expected Pet struct")
	}

	if !strings.Contains(code, "type Owner struct") {
		t.Error("Expected Owner struct")
	}

	if !strings.Contains(code, "type Store struct") {
		t.Error("Expected Store struct")
	}
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
			if result != tt.expected {
				t.Errorf("contains() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
