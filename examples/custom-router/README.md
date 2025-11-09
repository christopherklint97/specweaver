# Custom Router Example

This example demonstrates how to use a custom HTTP router with SpecWeaver-generated code.

## Overview

SpecWeaver generates code that works with any router implementing the `router.Router` interface. This example shows how to use the popular [chi](https://github.com/go-chi/chi) router as a drop-in replacement for the built-in router.

## Router Interface

Custom routers must implement the `router.Router` interface:

```go
type Router interface {
    http.Handler

    Use(middleware ...func(http.Handler) http.Handler)
    Get(pattern string, handler http.HandlerFunc)
    Post(pattern string, handler http.HandlerFunc)
    Put(pattern string, handler http.HandlerFunc)
    Delete(pattern string, handler http.HandlerFunc)
    Patch(pattern string, handler http.HandlerFunc)
    Options(pattern string, handler http.HandlerFunc)
    Head(pattern string, handler http.HandlerFunc)
}
```

## URL Parameter Compatibility

For path parameters to work correctly, custom routers must store URL parameters in the request context using `router.URLParamKey`:

```go
params := map[string]string{"id": "123"}
ctx := context.WithValue(r.Context(), router.URLParamKey, params)
handler.ServeHTTP(w, r.WithContext(ctx))
```

## Using a Custom Router

### Step 1: Create an Adapter

Create an adapter that implements the `router.Router` interface for your chosen router:

```go
type ChiAdapter struct {
    *chi.Mux
}

func NewChiAdapter() *ChiAdapter {
    return &ChiAdapter{Mux: chi.NewMux()}
}

// Implement all Router interface methods...
```

### Step 2: Add URL Parameter Middleware

Create middleware to extract URL parameters and store them in the context:

```go
func ChiURLParamMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rctx := chi.RouteContext(r.Context())
        if rctx != nil && len(rctx.URLParams.Keys) > 0 {
            params := make(map[string]string)
            for i, key := range rctx.URLParams.Keys {
                params[key] = rctx.URLParams.Values[i]
            }
            ctx := context.WithValue(r.Context(), router.URLParamKey, params)
            r = r.WithContext(ctx)
        }
        next.ServeHTTP(w, r)
    })
}
```

### Step 3: Configure the Router

Use `ConfigureRouter` instead of `NewRouter`:

```go
// Create your custom router
chiRouter := NewChiAdapter()

// Add middleware (including the URL parameter middleware)
chiRouter.Use(middleware.Logger)
chiRouter.Use(ChiURLParamMiddleware)

// Configure with SpecWeaver routes
api.ConfigureRouter(chiRouter, server)

// Start server
http.ListenAndServe(":8080", chiRouter)
```

## Running the Example

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run the server:**
   ```bash
   go run .
   ```

3. **Test the API:**
   ```bash
   # List pets
   curl http://localhost:8080/pets

   # Get a specific pet
   curl http://localhost:8080/pets/1

   # Create a pet
   curl -X POST http://localhost:8080/pets \
     -H "Content-Type: application/json" \
     -d '{"name": "Buddy", "tag": "dog"}'
   ```

## Benefits of Custom Routers

- **Use familiar tools**: Integrate with routers you already know
- **Advanced features**: Leverage router-specific features (sub-routers, route groups, etc.)
- **Ecosystem compatibility**: Use existing middleware and tools from your router's ecosystem
- **Zero lock-in**: Switch routers without changing your business logic

## Compatible Routers

Any router can be adapted to work with SpecWeaver as long as it can:

1. Implement the `router.Router` interface
2. Support path parameters
3. Store URL parameters in context using `router.URLParamKey`

Popular routers that can be adapted:
- [chi](https://github.com/go-chi/chi) (shown in this example)
- [gorilla/mux](https://github.com/gorilla/mux)
- [httprouter](https://github.com/julienschmidt/httprouter)
- [echo](https://github.com/labstack/echo)
- [gin](https://github.com/gin-gonic/gin)
- Any other router with similar capabilities

## Files

- `chi_adapter.go` - Chi router adapter implementation
- `main.go` - Example server using chi router
- `README.md` - This file
