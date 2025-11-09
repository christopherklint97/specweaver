package router

import "net/http"

// Router is the interface that custom routers must implement to work with SpecWeaver.
// Any router that implements this interface can be used as a drop-in replacement for
// the built-in router.
//
// The router must:
//  1. Support all standard HTTP methods (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
//  2. Support middleware via the Use method
//  3. Support path parameters in the format {paramName}
//  4. Store path parameters in the request context using URLParamKey
//  5. Implement http.Handler interface
//
// Example of storing path parameters in context:
//
//	params := map[string]string{"id": "123"}
//	ctx := context.WithValue(r.Context(), router.URLParamKey, params)
//	handler.ServeHTTP(w, r.WithContext(ctx))
type Router interface {
	http.Handler

	// Use adds middleware to the router
	Use(middleware ...func(http.Handler) http.Handler)

	// Get registers a GET route
	Get(pattern string, handler http.HandlerFunc)

	// Post registers a POST route
	Post(pattern string, handler http.HandlerFunc)

	// Put registers a PUT route
	Put(pattern string, handler http.HandlerFunc)

	// Delete registers a DELETE route
	Delete(pattern string, handler http.HandlerFunc)

	// Patch registers a PATCH route
	Patch(pattern string, handler http.HandlerFunc)

	// Options registers an OPTIONS route
	Options(pattern string, handler http.HandlerFunc)

	// Head registers a HEAD route
	Head(pattern string, handler http.HandlerFunc)
}
