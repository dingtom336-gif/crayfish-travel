package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestTraceIDGenerated verifies that TraceID middleware generates a UUID when none is provided.
func TestTraceIDGenerated(t *testing.T) {
	r := gin.New()
	r.Use(TraceID())
	r.GET("/test", func(c *gin.Context) {
		traceID := GetTraceID(c)
		c.JSON(http.StatusOK, gin.H{"trace_id": traceID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Response header should have a trace ID
	traceID := w.Header().Get(TraceIDHeader)
	if traceID == "" {
		t.Error("expected non-empty X-Trace-ID header")
	}

	// UUID format: 8-4-4-4-12
	if len(traceID) != 36 {
		t.Errorf("expected UUID-length trace ID (36 chars), got %d chars: %q", len(traceID), traceID)
	}
}

// TestTraceIDPreserved verifies that a client-provided trace ID is reused.
func TestTraceIDPreserved(t *testing.T) {
	r := gin.New()
	r.Use(TraceID())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"trace_id": GetTraceID(c)})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(TraceIDHeader, "custom-trace-id")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(TraceIDHeader); got != "custom-trace-id" {
		t.Errorf("expected trace ID 'custom-trace-id', got %q", got)
	}
}

// TestWithTraceIDContext verifies stdlib context propagation.
func TestWithTraceIDContext(t *testing.T) {
	ctx := WithTraceID(context.Background(), "ctx-trace-123")
	got := TraceIDFromContext(ctx)
	if got != "ctx-trace-123" {
		t.Errorf("expected 'ctx-trace-123', got %q", got)
	}
}

// TestTraceIDFromEmptyContext returns empty string for missing trace ID.
func TestTraceIDFromEmptyContext(t *testing.T) {
	got := TraceIDFromContext(context.Background())
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// TestCORSAllowedOrigin verifies CORS allows whitelisted origins.
func TestCORSAllowedOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:3000", "http://example.com"}))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Errorf("expected CORS origin 'http://example.com', got %q", got)
	}
}

// TestCORSWildcardWhenEmpty verifies CORS uses wildcard when no origins configured.
func TestCORSWildcardWhenEmpty(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{}))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://anything.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected wildcard CORS, got %q", got)
	}
}

// TestCORSBlocksUnknownOrigin verifies unknown origins are not allowed.
func TestCORSBlocksUnknownOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:3000"}))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS origin for unknown origin, got %q", got)
	}
}

// TestCORSPreflight verifies OPTIONS returns 204 with correct headers.
func TestCORSPreflight(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:3000"}))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

// TestAdminAuthValidToken verifies correct token passes.
func TestAdminAuthValidToken(t *testing.T) {
	r := gin.New()
	r.Use(AdminAuth("test-secret"))
	r.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "admin")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("X-Admin-Token", "test-secret")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", w.Code)
	}
}

// TestAdminAuthInvalidToken verifies wrong token is rejected.
func TestAdminAuthInvalidToken(t *testing.T) {
	r := gin.New()
	r.Use(AdminAuth("test-secret"))
	r.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "admin")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("X-Admin-Token", "wrong")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", w.Code)
	}
}

// TestAdminAuthEmptyToken verifies admin is disabled when token is empty.
func TestAdminAuthEmptyToken(t *testing.T) {
	r := gin.New()
	r.Use(AdminAuth(""))
	r.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "admin")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 when admin disabled, got %d", w.Code)
	}
}
