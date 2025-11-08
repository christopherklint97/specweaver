package router

import (
	"context"
	"net/http"
	"strings"
)

// Mux is a simple HTTP request multiplexer
type Mux struct {
	routes     []*route
	middleware []func(http.Handler) http.Handler
	notFound   http.Handler
}

// route represents a single route
type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
	parts   []pathPart
}

// pathPart represents a part of a URL path
type pathPart struct {
	isParam bool
	value   string
}

// contextKey is a custom type for context keys
type contextKey string

const (
	// URLParamKey is the context key for URL parameters
	URLParamKey contextKey = "urlParams"
)

// NewRouter creates a new Mux router
func NewRouter() *Mux {
	return &Mux{
		routes:     make([]*route, 0),
		middleware: make([]func(http.Handler) http.Handler, 0),
		notFound:   http.NotFoundHandler(),
	}
}

// Use adds middleware to the router
func (m *Mux) Use(middleware ...func(http.Handler) http.Handler) {
	m.middleware = append(m.middleware, middleware...)
}

// Get registers a GET route
func (m *Mux) Get(pattern string, handler http.HandlerFunc) {
	m.handle(http.MethodGet, pattern, handler)
}

// Post registers a POST route
func (m *Mux) Post(pattern string, handler http.HandlerFunc) {
	m.handle(http.MethodPost, pattern, handler)
}

// Put registers a PUT route
func (m *Mux) Put(pattern string, handler http.HandlerFunc) {
	m.handle(http.MethodPut, pattern, handler)
}

// Delete registers a DELETE route
func (m *Mux) Delete(pattern string, handler http.HandlerFunc) {
	m.handle(http.MethodDelete, pattern, handler)
}

// Patch registers a PATCH route
func (m *Mux) Patch(pattern string, handler http.HandlerFunc) {
	m.handle(http.MethodPatch, pattern, handler)
}

// Options registers an OPTIONS route
func (m *Mux) Options(pattern string, handler http.HandlerFunc) {
	m.handle(http.MethodOptions, pattern, handler)
}

// Head registers a HEAD route
func (m *Mux) Head(pattern string, handler http.HandlerFunc) {
	m.handle(http.MethodHead, pattern, handler)
}

// handle registers a route with the given method and pattern
func (m *Mux) handle(method, pattern string, handler http.HandlerFunc) {
	parts := parsePattern(pattern)
	m.routes = append(m.routes, &route{
		method:  method,
		pattern: pattern,
		handler: handler,
		parts:   parts,
	})
}

// ServeHTTP implements the http.Handler interface
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Build the handler chain with middleware
	var handler http.Handler = http.HandlerFunc(m.serve)

	// Apply middleware in reverse order
	for i := len(m.middleware) - 1; i >= 0; i-- {
		handler = m.middleware[i](handler)
	}

	handler.ServeHTTP(w, r)
}

// serve handles the actual routing
func (m *Mux) serve(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Find matching route
	for _, route := range m.routes {
		if route.method != r.Method {
			continue
		}

		if params, ok := matchPattern(route.parts, path); ok {
			// Add URL parameters to context
			ctx := r.Context()
			if len(params) > 0 {
				ctx = context.WithValue(ctx, URLParamKey, params)
			}
			route.handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}

	// No route found
	m.notFound.ServeHTTP(w, r)
}

// parsePattern parses a URL pattern into parts
func parsePattern(pattern string) []pathPart {
	pattern = strings.TrimPrefix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")

	if pattern == "" {
		return []pathPart{}
	}

	segments := strings.Split(pattern, "/")
	parts := make([]pathPart, len(segments))

	for i, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			// This is a parameter
			parts[i] = pathPart{
				isParam: true,
				value:   segment[1 : len(segment)-1],
			}
		} else {
			// This is a literal segment
			parts[i] = pathPart{
				isParam: false,
				value:   segment,
			}
		}
	}

	return parts
}

// matchPattern checks if a path matches a pattern and returns parameters
func matchPattern(parts []pathPart, path string) (map[string]string, bool) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	pathSegments := []string{}
	if path != "" {
		pathSegments = strings.Split(path, "/")
	}

	// Check if the number of segments matches
	if len(parts) != len(pathSegments) {
		return nil, false
	}

	params := make(map[string]string)

	for i, part := range parts {
		if part.isParam {
			// This is a parameter, capture it
			params[part.value] = pathSegments[i]
		} else {
			// This is a literal, it must match exactly
			if part.value != pathSegments[i] {
				return nil, false
			}
		}
	}

	return params, true
}

// URLParam returns a URL parameter from the request context
func URLParam(r *http.Request, key string) string {
	ctx := r.Context()
	params, ok := ctx.Value(URLParamKey).(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}
