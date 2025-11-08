package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	require.NotNil(t, router, "Expected router to be created")
	assert.NotNil(t, router.routes, "Expected routes to be initialized")
	assert.NotNil(t, router.middleware, "Expected middleware to be initialized")
	assert.NotNil(t, router.notFound, "Expected notFound handler to be initialized")
}

func TestRouterGet(t *testing.T) {
	router := NewRouter()
	called := false

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("GET test"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, called, "Handler was not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "GET test", w.Body.String())
}

func TestRouterPost(t *testing.T) {
	router := NewRouter()
	called := false

	router.Post("/items", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("POST items"))
	})

	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, called, "Handler was not called")
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRouterPut(t *testing.T) {
	router := NewRouter()

	router.Put("/items/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouterDelete(t *testing.T) {
	router := NewRouter()

	router.Delete("/items/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRouterPatch(t *testing.T) {
	router := NewRouter()

	router.Patch("/items/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPatch, "/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouterOptions(t *testing.T) {
	router := NewRouter()

	router.Options("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/items", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouterHead(t *testing.T) {
	router := NewRouter()

	router.Head("/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodHead, "/items", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouterURLParams(t *testing.T) {
	router := NewRouter()

	router.Get("/users/{userId}/posts/{postId}", func(w http.ResponseWriter, r *http.Request) {
		userId := URLParam(r, "userId")
		postId := URLParam(r, "postId")

		assert.Equal(t, "123", userId)
		assert.Equal(t, "456", postId)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouterURLParamNotFound(t *testing.T) {
	router := NewRouter()

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		param := URLParam(r, "nonexistent")
		assert.Empty(t, param, "Expected empty string for non-existent param")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
}

func TestRouterNotFound(t *testing.T) {
	router := NewRouter()

	router.Get("/existing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouterMethodNotAllowed(t *testing.T) {
	router := NewRouter()

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Try POST on a GET-only route
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouterMiddleware(t *testing.T) {
	router := NewRouter()

	// Middleware that adds a header
	headerMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	router.Use(headerMiddleware)

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "middleware", w.Header().Get("X-Test"), "Expected middleware to set header")
}

func TestRouterMultipleMiddleware(t *testing.T) {
	router := NewRouter()

	order := []string{}

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1-before")
			next.ServeHTTP(w, r)
			order = append(order, "m1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2-before")
			next.ServeHTTP(w, r)
			order = append(order, "m2-after")
		})
	}

	router.Use(middleware1, middleware2)

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	require.Len(t, order, len(expected))

	for i, v := range expected {
		assert.Equal(t, v, order[i])
	}
}

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []pathPart
	}{
		{
			name:     "Root path",
			pattern:  "/",
			expected: []pathPart{},
		},
		{
			name:    "Simple path",
			pattern: "/users",
			expected: []pathPart{
				{isParam: false, value: "users"},
			},
		},
		{
			name:    "Path with parameter",
			pattern: "/users/{id}",
			expected: []pathPart{
				{isParam: false, value: "users"},
				{isParam: true, value: "id"},
			},
		},
		{
			name:    "Multiple parameters",
			pattern: "/users/{userId}/posts/{postId}",
			expected: []pathPart{
				{isParam: false, value: "users"},
				{isParam: true, value: "userId"},
				{isParam: false, value: "posts"},
				{isParam: true, value: "postId"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := parsePattern(tt.pattern)

			require.Len(t, parts, len(tt.expected))

			for i, expected := range tt.expected {
				assert.Equal(t, expected.isParam, parts[i].isParam, "Part %d: isParam mismatch", i)
				assert.Equal(t, expected.value, parts[i].value, "Part %d: value mismatch", i)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name           string
		pattern        string
		path           string
		shouldMatch    bool
		expectedParams map[string]string
	}{
		{
			name:           "Exact match",
			pattern:        "/users",
			path:           "/users",
			shouldMatch:    true,
			expectedParams: map[string]string{},
		},
		{
			name:        "No match",
			pattern:     "/users",
			path:        "/posts",
			shouldMatch: false,
		},
		{
			name:        "Single parameter",
			pattern:     "/users/{id}",
			path:        "/users/123",
			shouldMatch: true,
			expectedParams: map[string]string{
				"id": "123",
			},
		},
		{
			name:        "Multiple parameters",
			pattern:     "/users/{userId}/posts/{postId}",
			path:        "/users/123/posts/456",
			shouldMatch: true,
			expectedParams: map[string]string{
				"userId": "123",
				"postId": "456",
			},
		},
		{
			name:        "Length mismatch",
			pattern:     "/users/{id}",
			path:        "/users",
			shouldMatch: false,
		},
		{
			name:           "Root path",
			pattern:        "/",
			path:           "/",
			shouldMatch:    true,
			expectedParams: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := parsePattern(tt.pattern)
			params, matched := matchPattern(parts, tt.path)

			assert.Equal(t, tt.shouldMatch, matched)

			if tt.shouldMatch {
				require.Len(t, params, len(tt.expectedParams))

				for key, expectedValue := range tt.expectedParams {
					assert.Equal(t, expectedValue, params[key], "Param %s mismatch", key)
				}
			}
		})
	}
}

func TestRouterTrailingSlash(t *testing.T) {
	router := NewRouter()

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name string
		path string
		code int
	}{
		{"Without trailing slash", "/test", http.StatusOK},
		{"With trailing slash", "/test/", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestRouterComplexRouting(t *testing.T) {
	router := NewRouter()

	// Register multiple routes
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("root"))
	})

	router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("users"))
	})

	router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := URLParam(r, "id")
		_, _ = w.Write([]byte("user-" + id))
	})

	router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("user-created"))
	})

	router.Get("/users/{id}/posts/{postId}", func(w http.ResponseWriter, r *http.Request) {
		userId := URLParam(r, "id")
		postId := URLParam(r, "postId")
		_, _ = w.Write([]byte("user-" + userId + "-post-" + postId))
	})

	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		{"Root", http.MethodGet, "/", http.StatusOK, "root"},
		{"List users", http.MethodGet, "/users", http.StatusOK, "users"},
		{"Get user", http.MethodGet, "/users/123", http.StatusOK, "user-123"},
		{"Create user", http.MethodPost, "/users", http.StatusCreated, "user-created"},
		{"Get user post", http.MethodGet, "/users/123/posts/456", http.StatusOK, "user-123-post-456"},
		{"Not found", http.MethodGet, "/nonexistent", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedBody != "" {
				body, err := io.ReadAll(w.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}
