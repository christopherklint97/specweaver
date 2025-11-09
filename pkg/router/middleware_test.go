package router

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	middleware := Logger(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check that log was written
	logOutput := buf.String()
	assert.Contains(t, logOutput, "GET", "Expected log to contain method GET")
	assert.Contains(t, logOutput, "/test", "Expected log to contain path /test")
	assert.Contains(t, logOutput, "200", "Expected log to contain status code 200")
}

func TestLoggerWithDifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			middleware := Logger(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			// Just verify that the log was written
			logOutput := buf.String()
			assert.NotEmpty(t, logOutput, "Expected log output to be written")
		})
	}
}

func TestLoggingResponseWriter(t *testing.T) {
	t.Run("WriteHeader captures status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		lrw.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, lrw.statusCode)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Default status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Write without calling WriteHeader
		_, _ = lrw.Write([]byte("test"))

		assert.Equal(t, http.StatusOK, lrw.statusCode)
	})
}

func TestRecoverer(t *testing.T) {
	t.Run("Recover from panic", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		middleware := Recoverer(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "panic recovered", "Expected log to contain panic message")
		assert.Contains(t, logOutput, "test panic", "Expected log to contain panic value")
	})

	t.Run("No panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("no panic"))
		})

		middleware := Recoverer(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "no panic", w.Body.String())
	})
}

func TestRequestID(t *testing.T) {
	t.Run("Generate request ID", func(t *testing.T) {
		var capturedID string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedID = GetRequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.NotEmpty(t, capturedID, "Expected request ID to be generated")

		headerID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, headerID, "Expected X-Request-ID header to be set")
		assert.Equal(t, capturedID, headerID, "Expected header ID to match context ID")
	})

	t.Run("Use existing request ID", func(t *testing.T) {
		existingID := "existing-id-123"
		var capturedID string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedID = GetRequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, existingID, capturedID)
		assert.Equal(t, existingID, w.Header().Get("X-Request-ID"))
	})
}

func TestGetRequestID(t *testing.T) {
	t.Run("Get request ID from context", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := GetRequestID(r.Context())
			assert.NotEmpty(t, id, "Expected request ID to be in context")
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)
	})

	t.Run("No request ID in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		id := GetRequestID(req.Context())

		assert.Empty(t, id)
	})
}

func TestRealIP(t *testing.T) {
	t.Run("X-Forwarded-For header", func(t *testing.T) {
		var capturedIP string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedIP = r.RemoteAddr
			w.WriteHeader(http.StatusOK)
		})

		middleware := RealIP(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, "1.2.3.4", capturedIP)
	})

	t.Run("X-Real-IP header", func(t *testing.T) {
		var capturedIP string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedIP = r.RemoteAddr
			w.WriteHeader(http.StatusOK)
		})

		middleware := RealIP(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Real-IP", "5.6.7.8")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, "5.6.7.8", capturedIP)
	})

	t.Run("X-Forwarded-For takes precedence", func(t *testing.T) {
		var capturedIP string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedIP = r.RemoteAddr
			w.WriteHeader(http.StatusOK)
		})

		middleware := RealIP(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		req.Header.Set("X-Real-IP", "5.6.7.8")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, "1.2.3.4", capturedIP, "Expected IP from X-Forwarded-For")
	})

	t.Run("No IP headers", func(t *testing.T) {
		originalIP := "10.0.0.1:12345"
		var capturedIP string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedIP = r.RemoteAddr
			w.WriteHeader(http.StatusOK)
		})

		middleware := RealIP(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = originalIP
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, originalIP, capturedIP)
	})
}

func TestMiddlewareChaining(t *testing.T) {
	// Capture log output to prevent nil pointer issues
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	var order []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")

		// Check that all middleware modifications are present
		assert.NotEmpty(t, GetRequestID(r.Context()), "Expected request ID to be in context")

		w.WriteHeader(http.StatusOK)
	})

	// Chain middleware: Recoverer -> RequestID -> RealIP -> Logger
	chain := Recoverer(
		RequestID(
			RealIP(
				Logger(handler),
			),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	chain.ServeHTTP(w, req)

	require.Len(t, order, 1)
	assert.Equal(t, "handler", order[0], "Expected handler to be called")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddlewareWithRouter(t *testing.T) {
	// Capture log output to prevent nil pointer issues
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	router := NewRouter()

	// Add middleware
	router.Use(Logger)
	router.Use(Recoverer)
	router.Use(RequestID)
	router.Use(RealIP)

	var requestID string

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		requestID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, requestID, "Expected request ID to be available in handler")
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"), "Expected X-Request-ID header to be set")
}
