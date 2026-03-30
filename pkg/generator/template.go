package generator

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Template data structures for types.go.tmpl

// TypesTemplateData holds data for generating types.go
type TypesTemplateData struct {
	UsesTime bool
	UsesDate bool
	Types    []TypeDefinition
}

// TypeDefinition represents a single type to generate
type TypeDefinition struct {
	Name        string
	Description string
	Kind        string // "struct", "enum", "alias", "array"
	Fields      []FieldDefinition
	EnumValues  []EnumValue
	AliasType   string
	ItemType    string
}

// FieldDefinition represents a struct field
type FieldDefinition struct {
	Name        string
	Type        string
	JSONTag     string
	Description string
}

// EnumValue represents an enum constant value
type EnumValue struct {
	ConstName string
	Value     string
}

// Template data structures for server.go.tmpl

// ServerTemplateData holds data for generating server.go
type ServerTemplateData struct {
	NeedsStrconv       bool
	HasSecuritySchemes bool
	RequestTypes       []RequestTypeData
	ResponseTypes      []ResponseTypeData
	Operations         []OperationData
	SecuritySchemes    []SecuritySchemeData
	Routes             []RouteData
}

// RequestTypeData holds data for a request type
type RequestTypeData struct {
	Name        string
	HandlerName string
	Fields      []FieldDefinition
}

// ResponseTypeData holds data for a response type
type ResponseTypeData struct {
	InterfaceName string
	HandlerName   string
	ConcreteTypes []ConcreteResponseType
}

// ConcreteResponseType holds data for a concrete response type
type ConcreteResponseType struct {
	Name       string
	StatusCode int
	HasBody    bool
	BodyType   string
}

// OperationData holds data for an operation
type OperationData struct {
	HandlerName      string
	Summary          string
	RequestTypeName  string
	ResponseTypeName string
	PathParams       []ParamData
	QueryParams      []ParamData
	HasRequestBody   bool
}

// ParamData holds data for a parameter
type ParamData struct {
	ParamName string
	FieldName string
	BaseType  string
	BitSize   string
	Required  bool
}

// SecuritySchemeData holds data for a security scheme
type SecuritySchemeData struct {
	Name      string
	Type      string
	Scheme    string
	In        string
	ParamName string
}

// RouteData holds data for a route
type RouteData struct {
	Method      string
	Path        string
	HandlerName string
	HasSecurity bool
	SecurityReqs string
}

// Template data structures for auth.go.tmpl

// AuthTemplateData holds data for generating auth.go
type AuthTemplateData struct {
	Schemes []AuthSchemeData
}

// AuthSchemeData holds data for an auth scheme
type AuthSchemeData struct {
	Name       string
	MethodName string
	Type       string
	Scheme     string
}

// Template data structures for webhooks.go.tmpl

// WebhooksTemplateData holds data for generating webhooks.go
type WebhooksTemplateData struct {
	RequestTypes  []WebhookRequestTypeData
	ResponseTypes []WebhookResponseTypeData
	Operations    []WebhookOperationData
}

// WebhookRequestTypeData holds data for a webhook request type
type WebhookRequestTypeData struct {
	Name        string
	WebhookName string
	Fields      []FieldDefinition
}

// WebhookResponseTypeData holds data for a webhook response type
type WebhookResponseTypeData struct {
	InterfaceName string
	WebhookName   string
	ConcreteTypes []ConcreteResponseType
}

// WebhookOperationData holds data for a webhook operation
type WebhookOperationData struct {
	HandlerName      string
	WebhookName      string
	Summary          string
	RequestTypeName  string
	ResponseTypeName string
	Method           string
	HasRequestBody   bool
	HasHeaders       bool
	Headers          []HeaderData
	Responses        []WebhookResponseData
}

// HeaderData holds data for a header parameter
type HeaderData struct {
	HeaderName string
	FieldName  string
	Required   bool
}

// WebhookResponseData holds data for a webhook response
type WebhookResponseData struct {
	StatusCode int
	TypeName   string
	HasBody    bool
}

// executeTemplate executes a template with the given data
func executeTemplate(name string, data any) (string, error) {
	tmplContent, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(name).Parse(string(tmplContent))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
