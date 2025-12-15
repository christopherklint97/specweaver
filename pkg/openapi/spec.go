package openapi

// Document represents the root OpenAPI specification document
// Supports OpenAPI 3.0.x, 3.1.x, and 3.2.x
type Document struct {
	OpenAPI    string                `yaml:"openapi" json:"openapi"`
	Info       *Info                 `yaml:"info" json:"info"`
	Servers    []*Server             `yaml:"servers,omitempty" json:"servers,omitempty"`
	Paths      Paths                 `yaml:"paths,omitempty" json:"paths,omitempty"`
	Webhooks   Webhooks              `yaml:"webhooks,omitempty" json:"webhooks,omitempty"`
	Components *Components           `yaml:"components,omitempty" json:"components,omitempty"`
	Security   []SecurityRequirement `yaml:"security,omitempty" json:"security,omitempty"`
	Tags       []*Tag                `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Internal fields for reference resolution
	refCache map[string]any
}

// Info provides metadata about the API
type Info struct {
	Title       string  `yaml:"title" json:"title"`
	Version     string  `yaml:"version" json:"version"`
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Contact     *Contact `yaml:"contact,omitempty" json:"contact,omitempty"`
	License     *License `yaml:"license,omitempty" json:"license,omitempty"`
}

// Contact contains contact information
type Contact struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	URL   string `yaml:"url,omitempty" json:"url,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

// License contains license information
type License struct {
	Name string `yaml:"name" json:"name"`
	URL  string `yaml:"url,omitempty" json:"url,omitempty"`
}

// Server represents a server definition
type Server struct {
	URL         string                     `yaml:"url" json:"url"`
	Description string                     `yaml:"description,omitempty" json:"description,omitempty"`
	Variables   map[string]*ServerVariable `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// ServerVariable represents a server variable
type ServerVariable struct {
	Enum        []string `yaml:"enum,omitempty" json:"enum,omitempty"`
	Default     string   `yaml:"default" json:"default"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
}

// Paths holds the relative paths to the individual endpoints
type Paths map[string]*PathItem

// Webhooks holds the webhooks that may be received as part of this API
// Available in OpenAPI 3.1+
type Webhooks map[string]*PathItem

// PathItem describes the operations available on a single path
type PathItem struct {
	Ref         string      `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	Summary     string      `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Get         *Operation  `yaml:"get,omitempty" json:"get,omitempty"`
	Put         *Operation  `yaml:"put,omitempty" json:"put,omitempty"`
	Post        *Operation  `yaml:"post,omitempty" json:"post,omitempty"`
	Delete      *Operation  `yaml:"delete,omitempty" json:"delete,omitempty"`
	Options     *Operation  `yaml:"options,omitempty" json:"options,omitempty"`
	Head        *Operation  `yaml:"head,omitempty" json:"head,omitempty"`
	Patch       *Operation  `yaml:"patch,omitempty" json:"patch,omitempty"`
	Trace       *Operation  `yaml:"trace,omitempty" json:"trace,omitempty"`
	Servers     []*Server   `yaml:"servers,omitempty" json:"servers,omitempty"`
	Parameters  []*Parameter `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// Operation describes a single API operation on a path
type Operation struct {
	Tags        []string              `yaml:"tags,omitempty" json:"tags,omitempty"`
	Summary     string                `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string                `yaml:"description,omitempty" json:"description,omitempty"`
	OperationID string                `yaml:"operationId,omitempty" json:"operationId,omitempty"`
	Parameters  []*Parameter          `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	RequestBody *RequestBody          `yaml:"requestBody,omitempty" json:"requestBody,omitempty"`
	Responses   Responses             `yaml:"responses,omitempty" json:"responses,omitempty"`
	Deprecated  bool                  `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
	Security    []SecurityRequirement `yaml:"security,omitempty" json:"security,omitempty"`
	Servers     []*Server             `yaml:"servers,omitempty" json:"servers,omitempty"`
}

// Parameter describes a single operation parameter
type Parameter struct {
	Name            string      `yaml:"name" json:"name"`
	In              string      `yaml:"in" json:"in"` // query, header, path, cookie
	Description     string      `yaml:"description,omitempty" json:"description,omitempty"`
	Required        bool        `yaml:"required,omitempty" json:"required,omitempty"`
	Deprecated      bool        `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
	AllowEmptyValue bool        `yaml:"allowEmptyValue,omitempty" json:"allowEmptyValue,omitempty"`
	Schema          *SchemaRef  `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example         any         `yaml:"example,omitempty" json:"example,omitempty"`
	Ref             string      `yaml:"$ref,omitempty" json:"$ref,omitempty"`
}

// RequestBody describes a request body
type RequestBody struct {
	Description string               `yaml:"description,omitempty" json:"description,omitempty"`
	Content     map[string]*MediaType `yaml:"content" json:"content"`
	Required    bool                 `yaml:"required,omitempty" json:"required,omitempty"`
	Ref         string               `yaml:"$ref,omitempty" json:"$ref,omitempty"`
}

// MediaType describes a media type
type MediaType struct {
	Schema   *SchemaRef         `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example  any                `yaml:"example,omitempty" json:"example,omitempty"`
	Examples map[string]*Example `yaml:"examples,omitempty" json:"examples,omitempty"`
}

// Example describes an example value
type Example struct {
	Summary     string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Value       any    `yaml:"value,omitempty" json:"value,omitempty"`
	Ref         string `yaml:"$ref,omitempty" json:"$ref,omitempty"`
}

// Responses is a container for the expected responses of an operation
type Responses map[string]*Response

// Response describes a single response from an API operation
type Response struct {
	Description string                `yaml:"description" json:"description"`
	Content     map[string]*MediaType `yaml:"content,omitempty" json:"content,omitempty"`
	Headers     map[string]*Header    `yaml:"headers,omitempty" json:"headers,omitempty"`
	Ref         string                `yaml:"$ref,omitempty" json:"$ref,omitempty"`
}

// Header describes a header parameter
type Header struct {
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool       `yaml:"required,omitempty" json:"required,omitempty"`
	Deprecated  bool       `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
	Schema      *SchemaRef `yaml:"schema,omitempty" json:"schema,omitempty"`
	Ref         string     `yaml:"$ref,omitempty" json:"$ref,omitempty"`
}

// Components holds a set of reusable objects for different aspects of the OAS
type Components struct {
	Schemas         map[string]*SchemaRef         `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	Responses       map[string]*Response          `yaml:"responses,omitempty" json:"responses,omitempty"`
	Parameters      map[string]*Parameter         `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Examples        map[string]*Example           `yaml:"examples,omitempty" json:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody       `yaml:"requestBodies,omitempty" json:"requestBodies,omitempty"`
	Headers         map[string]*Header            `yaml:"headers,omitempty" json:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme    `yaml:"securitySchemes,omitempty" json:"securitySchemes,omitempty"`
}

// SchemaRef is a wrapper that can contain either a Schema or a reference
type SchemaRef struct {
	Ref   string  `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	Value *Schema `yaml:",inline" json:",inline"`
}

// Schema describes the schema of input/output data
// Based on JSON Schema Draft 2020-12 (for OpenAPI 3.1+)
type Schema struct {
	// Core properties
	Type        []string           `yaml:"type,omitempty" json:"type,omitempty"` // Can be array in OpenAPI 3.1+
	Format      string             `yaml:"format,omitempty" json:"format,omitempty"`
	Title       string             `yaml:"title,omitempty" json:"title,omitempty"`
	Description string             `yaml:"description,omitempty" json:"description,omitempty"`
	Default     any                `yaml:"default,omitempty" json:"default,omitempty"`
	Example     any                `yaml:"example,omitempty" json:"example,omitempty"`

	// Validation properties
	MultipleOf       *float64 `yaml:"multipleOf,omitempty" json:"multipleOf,omitempty"`
	Maximum          *float64 `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	ExclusiveMaximum *float64 `yaml:"exclusiveMaximum,omitempty" json:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	ExclusiveMinimum *float64 `yaml:"exclusiveMinimum,omitempty" json:"exclusiveMinimum,omitempty"`
	MaxLength        *int     `yaml:"maxLength,omitempty" json:"maxLength,omitempty"`
	MinLength        *int     `yaml:"minLength,omitempty" json:"minLength,omitempty"`
	Pattern          string   `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	MaxItems         *int     `yaml:"maxItems,omitempty" json:"maxItems,omitempty"`
	MinItems         *int     `yaml:"minItems,omitempty" json:"minItems,omitempty"`
	UniqueItems      bool     `yaml:"uniqueItems,omitempty" json:"uniqueItems,omitempty"`
	MaxProperties    *int     `yaml:"maxProperties,omitempty" json:"maxProperties,omitempty"`
	MinProperties    *int     `yaml:"minProperties,omitempty" json:"minProperties,omitempty"`

	// Object properties
	Properties           map[string]*SchemaRef `yaml:"properties,omitempty" json:"properties,omitempty"`
	Required             []string              `yaml:"required,omitempty" json:"required,omitempty"`
	AdditionalProperties *SchemaRef            `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`

	// Array properties
	Items *SchemaRef `yaml:"items,omitempty" json:"items,omitempty"`

	// Enum
	Enum []any `yaml:"enum,omitempty" json:"enum,omitempty"`

	// Composition
	AllOf []*SchemaRef `yaml:"allOf,omitempty" json:"allOf,omitempty"`
	OneOf []*SchemaRef `yaml:"oneOf,omitempty" json:"oneOf,omitempty"`
	AnyOf []*SchemaRef `yaml:"anyOf,omitempty" json:"anyOf,omitempty"`
	Not   *SchemaRef   `yaml:"not,omitempty" json:"not,omitempty"`

	// Other
	Nullable   bool `yaml:"nullable,omitempty" json:"nullable,omitempty"` // OpenAPI 3.0 specific
	ReadOnly   bool `yaml:"readOnly,omitempty" json:"readOnly,omitempty"`
	WriteOnly  bool `yaml:"writeOnly,omitempty" json:"writeOnly,omitempty"`
	Deprecated bool `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
}

// SecurityScheme defines a security scheme
type SecurityScheme struct {
	Type             string            `yaml:"type" json:"type"` // apiKey, http, oauth2, openIdConnect
	Description      string            `yaml:"description,omitempty" json:"description,omitempty"`
	Name             string            `yaml:"name,omitempty" json:"name,omitempty"`
	In               string            `yaml:"in,omitempty" json:"in,omitempty"`
	Scheme           string            `yaml:"scheme,omitempty" json:"scheme,omitempty"`
	BearerFormat     string            `yaml:"bearerFormat,omitempty" json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows       `yaml:"flows,omitempty" json:"flows,omitempty"`
	OpenIDConnectURL string            `yaml:"openIdConnectUrl,omitempty" json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows allows configuration of the supported OAuth Flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `yaml:"implicit,omitempty" json:"implicit,omitempty"`
	Password          *OAuthFlow `yaml:"password,omitempty" json:"password,omitempty"`
	ClientCredentials *OAuthFlow `yaml:"clientCredentials,omitempty" json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `yaml:"authorizationCode,omitempty" json:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth 2.0 flow
type OAuthFlow struct {
	AuthorizationURL string            `yaml:"authorizationUrl,omitempty" json:"authorizationUrl,omitempty"`
	TokenURL         string            `yaml:"tokenUrl,omitempty" json:"tokenUrl,omitempty"`
	RefreshURL       string            `yaml:"refreshUrl,omitempty" json:"refreshUrl,omitempty"`
	Scopes           map[string]string `yaml:"scopes" json:"scopes"`
}

// SecurityRequirement lists the required security schemes
type SecurityRequirement map[string][]string

// Tag adds metadata to a single tag
type Tag struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// GetSchemaType returns the primary type of the schema
// Handles both single type (string) and array of types (OpenAPI 3.1+)
func (s *Schema) GetSchemaType() string {
	if s == nil || len(s.Type) == 0 {
		return ""
	}
	return s.Type[0]
}

// IsRefOnly returns true if this SchemaRef only contains a reference
func (sr *SchemaRef) IsRefOnly() bool {
	return sr != nil && sr.Ref != ""
}
