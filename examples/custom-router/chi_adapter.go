package main

import (
	"context"
	"net/http"

	"github.com/christopherklint97/specweaver/pkg/router"
	"github.com/go-chi/chi/v5"
)

// ChiAdapter wraps chi.Mux to implement the router.Router interface
type ChiAdapter struct {
	*chi.Mux
}

// NewChiAdapter creates a new chi router adapter
func NewChiAdapter() *ChiAdapter {
	return &ChiAdapter{
		Mux: chi.NewMux(),
	}
}

// Get registers a GET route
func (c *ChiAdapter) Get(pattern string, handler http.HandlerFunc) {
	c.Mux.Get(pattern, handler)
}

// Post registers a POST route
func (c *ChiAdapter) Post(pattern string, handler http.HandlerFunc) {
	c.Mux.Post(pattern, handler)
}

// Put registers a PUT route
func (c *ChiAdapter) Put(pattern string, handler http.HandlerFunc) {
	c.Mux.Put(pattern, handler)
}

// Delete registers a DELETE route
func (c *ChiAdapter) Delete(pattern string, handler http.HandlerFunc) {
	c.Mux.Delete(pattern, handler)
}

// Patch registers a PATCH route
func (c *ChiAdapter) Patch(pattern string, handler http.HandlerFunc) {
	c.Mux.Patch(pattern, handler)
}

// Options registers an OPTIONS route
func (c *ChiAdapter) Options(pattern string, handler http.HandlerFunc) {
	c.Mux.Options(pattern, handler)
}

// Head registers a HEAD route
func (c *ChiAdapter) Head(pattern string, handler http.HandlerFunc) {
	c.Mux.Head(pattern, handler)
}

// Use adds middleware to the router
// Chi middleware needs to be adapted to match the expected signature
func (c *ChiAdapter) Use(middleware ...func(http.Handler) http.Handler) {
	for _, m := range middleware {
		c.Mux.Use(m)
	}
}

// ChiURLParamMiddleware is middleware that extracts chi URL parameters and stores them
// in the context using the router.URLParamKey so they're compatible with SpecWeaver's expectations
func ChiURLParamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract all chi URL parameters
		rctx := chi.RouteContext(r.Context())
		if rctx != nil && len(rctx.URLParams.Keys) > 0 {
			params := make(map[string]string)
			for i, key := range rctx.URLParams.Keys {
				if i < len(rctx.URLParams.Values) {
					params[key] = rctx.URLParams.Values[i]
				}
			}
			// Store in context using SpecWeaver's URLParamKey
			ctx := context.WithValue(r.Context(), router.URLParamKey, params)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
