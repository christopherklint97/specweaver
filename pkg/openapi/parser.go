package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load parses an OpenAPI specification from a file
// Supports both JSON and YAML formats
// Supports OpenAPI 3.0.x, 3.1.x, and 3.2.x
func Load(filePath string) (*Document, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadFromData(data, filePath)
}

// LoadFromData parses an OpenAPI specification from bytes
func LoadFromData(data []byte, sourcePath string) (*Document, error) {
	doc := &Document{
		refCache: make(map[string]any),
	}

	// Try to detect format and unmarshal
	ext := strings.ToLower(filepath.Ext(sourcePath))
	if ext == ".json" {
		if err := json.Unmarshal(data, doc); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		// Default to YAML (supports .yaml, .yml, and files without extension)
		if err := yaml.Unmarshal(data, doc); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	}

	// Normalize the schema type fields (handle both string and array)
	if err := normalizeDocument(doc); err != nil {
		return nil, fmt.Errorf("failed to normalize document: %w", err)
	}

	// Validate minimum requirements
	if err := validateDocument(doc); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return doc, nil
}

// normalizeDocument normalizes type fields to always be arrays
// This handles the difference between OpenAPI 3.0 (type: string) and 3.1+ (type: [string])
func normalizeDocument(doc *Document) error {
	// Normalize schemas in components
	if doc.Components != nil && doc.Components.Schemas != nil {
		for _, schemaRef := range doc.Components.Schemas {
			if err := normalizeSchemaRef(schemaRef); err != nil {
				return err
			}
		}
	}

	// Normalize schemas in paths
	if doc.Paths != nil {
		for _, pathItem := range doc.Paths {
			if err := normalizePathItem(pathItem); err != nil {
				return err
			}
		}
	}

	return nil
}

// normalizePathItem normalizes schemas in a path item
func normalizePathItem(item *PathItem) error {
	if item == nil {
		return nil
	}

	operations := []*Operation{
		item.Get, item.Put, item.Post, item.Delete,
		item.Options, item.Head, item.Patch, item.Trace,
	}

	for _, op := range operations {
		if err := normalizeOperation(op); err != nil {
			return err
		}
	}

	// Normalize parameters
	for _, param := range item.Parameters {
		if param != nil && param.Schema != nil {
			if err := normalizeSchemaRef(param.Schema); err != nil {
				return err
			}
		}
	}

	return nil
}

// normalizeOperation normalizes schemas in an operation
func normalizeOperation(op *Operation) error {
	if op == nil {
		return nil
	}

	// Normalize parameters
	for _, param := range op.Parameters {
		if param != nil && param.Schema != nil {
			if err := normalizeSchemaRef(param.Schema); err != nil {
				return err
			}
		}
	}

	// Normalize request body
	if op.RequestBody != nil {
		for _, mediaType := range op.RequestBody.Content {
			if mediaType != nil && mediaType.Schema != nil {
				if err := normalizeSchemaRef(mediaType.Schema); err != nil {
					return err
				}
			}
		}
	}

	// Normalize responses
	for _, response := range op.Responses {
		if response != nil {
			for _, mediaType := range response.Content {
				if mediaType != nil && mediaType.Schema != nil {
					if err := normalizeSchemaRef(mediaType.Schema); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// normalizeSchemaRef normalizes a schema reference
func normalizeSchemaRef(ref *SchemaRef) error {
	if ref == nil || ref.Value == nil {
		return nil
	}

	return normalizeSchema(ref.Value)
}

// normalizeSchema ensures the type field is always an array
func normalizeSchema(schema *Schema) error {
	if schema == nil {
		return nil
	}

	// Type is already normalized if it's already an array or empty
	// Nothing to do in that case

	// Normalize nested schemas
	if schema.Properties != nil {
		for _, prop := range schema.Properties {
			if err := normalizeSchemaRef(prop); err != nil {
				return err
			}
		}
	}

	if schema.Items != nil {
		if err := normalizeSchemaRef(schema.Items); err != nil {
			return err
		}
	}

	if schema.AdditionalProperties != nil {
		if err := normalizeSchemaRef(schema.AdditionalProperties); err != nil {
			return err
		}
	}

	// Normalize composition schemas
	for _, s := range schema.AllOf {
		if err := normalizeSchemaRef(s); err != nil {
			return err
		}
	}
	for _, s := range schema.OneOf {
		if err := normalizeSchemaRef(s); err != nil {
			return err
		}
	}
	for _, s := range schema.AnyOf {
		if err := normalizeSchemaRef(s); err != nil {
			return err
		}
	}
	if schema.Not != nil {
		if err := normalizeSchemaRef(schema.Not); err != nil {
			return err
		}
	}

	return nil
}

// validateDocument performs basic validation on the document
func validateDocument(doc *Document) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	if doc.OpenAPI == "" {
		return fmt.Errorf("openapi field is required")
	}

	// Validate version format (should be 3.x.x)
	if !strings.HasPrefix(doc.OpenAPI, "3.") {
		return fmt.Errorf("unsupported OpenAPI version: %s (only 3.x is supported)", doc.OpenAPI)
	}

	if doc.Info == nil {
		return fmt.Errorf("info field is required")
	}

	if doc.Info.Title == "" {
		return fmt.Errorf("info.title is required")
	}

	if doc.Info.Version == "" {
		return fmt.Errorf("info.version is required")
	}

	// At least one of paths, components, or webhooks should be present
	// (webhooks not yet implemented, so we check paths or components)
	if doc.Paths == nil && doc.Components == nil {
		return fmt.Errorf("document must have at least one of: paths, components")
	}

	return nil
}

// ResolveSchemaRef resolves a schema reference to its actual schema
func (doc *Document) ResolveSchemaRef(ref *SchemaRef) (*Schema, error) {
	if ref == nil {
		return nil, fmt.Errorf("schema reference is nil")
	}

	// If no $ref, return the value directly
	if ref.Ref == "" {
		return ref.Value, nil
	}

	// Parse the reference
	schema, err := doc.resolveReference(ref.Ref)
	if err != nil {
		return nil, err
	}

	// Type assert to Schema
	s, ok := schema.(*Schema)
	if !ok {
		return nil, fmt.Errorf("reference does not resolve to a schema: %s", ref.Ref)
	}

	return s, nil
}

// resolveReference resolves a $ref to the actual object
func (doc *Document) resolveReference(refPath string) (any, error) {
	// Only support local references for now (#/...)
	if !strings.HasPrefix(refPath, "#/") {
		return nil, fmt.Errorf("external references not supported: %s", refPath)
	}

	// Check cache
	if cached, ok := doc.refCache[refPath]; ok {
		return cached, nil
	}

	// Parse the reference path
	parts := strings.Split(refPath[2:], "/") // Remove "#/" prefix
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid reference path: %s", refPath)
	}

	// Navigate to the referenced object
	var current any = doc
	for i, part := range parts {
		// Unescape JSON pointer tokens
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")

		switch i {
		case 0: // Should be "components"
			if part != "components" {
				return nil, fmt.Errorf("unsupported reference root: %s (expected 'components')", part)
			}
			if doc.Components == nil {
				return nil, fmt.Errorf("components not defined in document")
			}
			current = doc.Components

		case 1: // Component type (schemas, responses, etc.)
			components := current.(*Components)
			switch part {
			case "schemas":
				if components.Schemas == nil {
					return nil, fmt.Errorf("schemas not defined in components")
				}
				current = components.Schemas
			case "responses":
				if components.Responses == nil {
					return nil, fmt.Errorf("responses not defined in components")
				}
				current = components.Responses
			case "parameters":
				if components.Parameters == nil {
					return nil, fmt.Errorf("parameters not defined in components")
				}
				current = components.Parameters
			case "requestBodies":
				if components.RequestBodies == nil {
					return nil, fmt.Errorf("requestBodies not defined in components")
				}
				current = components.RequestBodies
			default:
				return nil, fmt.Errorf("unsupported component type: %s", part)
			}

		case 2: // Component name
			switch v := current.(type) {
			case map[string]*SchemaRef:
				schemaRef, ok := v[part]
				if !ok {
					return nil, fmt.Errorf("schema not found: %s", part)
				}
				// Cache and return the schema value
				result := schemaRef.Value
				doc.refCache[refPath] = result
				return result, nil
			case map[string]*Response:
				response, ok := v[part]
				if !ok {
					return nil, fmt.Errorf("response not found: %s", part)
				}
				doc.refCache[refPath] = response
				return response, nil
			case map[string]*Parameter:
				param, ok := v[part]
				if !ok {
					return nil, fmt.Errorf("parameter not found: %s", part)
				}
				doc.refCache[refPath] = param
				return param, nil
			case map[string]*RequestBody:
				reqBody, ok := v[part]
				if !ok {
					return nil, fmt.Errorf("requestBody not found: %s", part)
				}
				doc.refCache[refPath] = reqBody
				return reqBody, nil
			default:
				return nil, fmt.Errorf("unexpected type at component name level: %T", v)
			}

		default:
			return nil, fmt.Errorf("reference path too deep: %s", refPath)
		}
	}

	return nil, fmt.Errorf("failed to resolve reference: %s", refPath)
}

// GetSchemaByRef retrieves a schema by its reference path (e.g., "#/components/schemas/Pet")
func (doc *Document) GetSchemaByRef(refPath string) (*Schema, error) {
	obj, err := doc.resolveReference(refPath)
	if err != nil {
		return nil, err
	}

	schema, ok := obj.(*Schema)
	if !ok {
		return nil, fmt.Errorf("reference does not point to a schema: %s", refPath)
	}

	return schema, nil
}

// GetSchemaByName retrieves a schema from components.schemas by name
func (doc *Document) GetSchemaByName(name string) (*Schema, error) {
	if doc.Components == nil || doc.Components.Schemas == nil {
		return nil, fmt.Errorf("no schemas defined in components")
	}

	schemaRef, ok := doc.Components.Schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema not found: %s", name)
	}

	return doc.ResolveSchemaRef(schemaRef)
}
