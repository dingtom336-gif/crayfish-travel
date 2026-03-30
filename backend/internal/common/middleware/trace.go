package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TraceIDHeader is the HTTP header key for trace ID propagation.
	TraceIDHeader = "X-Trace-ID"
	// TraceIDKey is the context key for trace ID retrieval.
	TraceIDKey = "trace_id"
)

// TraceID injects X-Trace-ID into the request context and response header.
// If the incoming request already carries X-Trace-ID, it is reused;
// otherwise a new UUID is generated.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)
		c.Next()
	}
}

// GetTraceID extracts the trace ID from a gin context.
func GetTraceID(c *gin.Context) string {
	if id, exists := c.Get(TraceIDKey); exists {
		return id.(string)
	}
	return ""
}

// WithTraceID returns a new stdlib context carrying the trace ID.
func WithTraceID(parent context.Context, traceID string) context.Context {
	return context.WithValue(parent, TraceIDKey, traceID)
}

// TraceIDFromContext extracts trace ID from a stdlib context.
func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(TraceIDKey).(string); ok {
		return id
	}
	return ""
}
