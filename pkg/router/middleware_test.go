package router

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	middleware := Logger(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check that log was written
	logOutput := buf.String()
	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain method GET")
	}

	if !strings.Contains(logOutput, "/test") {
		t.Error("Expected log to contain path /test")
	}

	if !strings.Contains(logOutput, "200") {
		t.Error("Expected log to contain status code 200")
	}
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
			if len(logOutput) == 0 {
				t.Error("Expected log output to be written")
			}
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

		if lrw.statusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, lrw.statusCode)
		}

		if w.Code != http.StatusCreated {
			t.Errorf("Expected underlying writer to have code %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("Default status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Write without calling WriteHeader
		lrw.Write([]byte("test"))

		if lrw.statusCode != http.StatusOK {
			t.Errorf("Expected default status code %d, got %d", http.StatusOK, lrw.statusCode)
		}
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

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		logOutput := buf.String()
		if !strings.Contains(logOutput, "panic recovered") {
			t.Error("Expected log to contain panic message")
		}

		if !strings.Contains(logOutput, "test panic") {
			t.Error("Expected log to contain panic value")
		}
	})

	t.Run("No panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("no panic"))
		})

		middleware := Recoverer(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body != "no panic" {
			t.Errorf("Expected body 'no panic', got %s", body)
		}
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

		if capturedID == "" {
			t.Error("Expected request ID to be generated")
		}

		headerID := w.Header().Get("X-Request-ID")
		if headerID == "" {
			t.Error("Expected X-Request-ID header to be set")
		}

		if headerID != capturedID {
			t.Error("Expected header ID to match context ID")
		}
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

		if capturedID != existingID {
			t.Errorf("Expected request ID %s, got %s", existingID, capturedID)
		}

		headerID := w.Header().Get("X-Request-ID")
		if headerID != existingID {
			t.Errorf("Expected header ID %s, got %s", existingID, headerID)
		}
	})
}

func TestGetRequestID(t *testing.T) {
	t.Run("Get request ID from context", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := GetRequestID(r.Context())
			if id == "" {
				t.Error("Expected request ID to be in context")
			}
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

		if id != "" {
			t.Errorf("Expected empty string, got %s", id)
		}
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

		if capturedIP != "1.2.3.4" {
			t.Errorf("Expected IP 1.2.3.4, got %s", capturedIP)
		}
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

		if capturedIP != "5.6.7.8" {
			t.Errorf("Expected IP 5.6.7.8, got %s", capturedIP)
		}
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

		if capturedIP != "1.2.3.4" {
			t.Errorf("Expected IP 1.2.3.4 from X-Forwarded-For, got %s", capturedIP)
		}
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

		if capturedIP != originalIP {
			t.Errorf("Expected original IP %s, got %s", originalIP, capturedIP)
		}
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
		if GetRequestID(r.Context()) == "" {
			t.Error("Expected request ID to be in context")
		}

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

	if len(order) != 1 || order[0] != "handler" {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
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

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if requestID == "" {
		t.Error("Expected request ID to be available in handler")
	}

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}
