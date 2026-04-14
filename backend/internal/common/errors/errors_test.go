package errors

import (
	"net/http"
	"testing"
)

func TestAppErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		wantCode int
		wantMsg  string
	}{
		{"BadRequest", BadRequest("bad input"), http.StatusBadRequest, "bad input"},
		{"NotFound", NotFound("not found"), http.StatusNotFound, "not found"},
		{"Internal", Internal("internal error"), http.StatusInternalServerError, "internal error"},
		{"Conflict", Conflict("conflict"), http.StatusConflict, "conflict"},
		{"TooManyRequests", TooManyRequests("rate limited"), http.StatusTooManyRequests, "rate limited"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", tt.err.Code, tt.wantCode)
			}
			if tt.err.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", tt.err.Message, tt.wantMsg)
			}
			// Verify Error() interface
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestAppErrorImplementsError(t *testing.T) {
	var _ error = BadRequest("test")
	var _ error = NotFound("test")
	var _ error = Internal("test")
	var _ error = Conflict("test")
	var _ error = TooManyRequests("test")
}

func TestAppErrorTraceID(t *testing.T) {
	err := &AppError{
		Code:    http.StatusBadRequest,
		Message: "test",
		TraceID: "trace-123",
	}

	if err.TraceID != "trace-123" {
		t.Errorf("trace ID = %q, want 'trace-123'", err.TraceID)
	}
}
