package capbypass

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAPIErrorMessage covers APIError.Error(): description preferred, code fallback.
func TestAPIErrorMessage(t *testing.T) {
	t.Run("uses description when present", func(t *testing.T) {
		e := &APIError{ErrorCode: "ERROR_CODE", ErrorDescription: "human readable"}
		assert.Equal(t, "human readable", e.Error())
	})

	t.Run("falls back to code when description empty", func(t *testing.T) {
		e := &APIError{ErrorCode: "ERROR_CODE"}
		assert.Equal(t, "ERROR_CODE", e.Error())
	})
}

// TestAPIErrorConstructors covers the new*Error constructors (their Error() string
// + that they carry the expected wrapped APIError fields).
func TestAPIErrorConstructors(t *testing.T) {
	t.Run("authentication", func(t *testing.T) {
		e := newAuthenticationError("ERROR_KEY_DOES_NOT_EXIST", "bad key")
		assert.Equal(t, "bad key", e.Error())
		assert.Equal(t, "ERROR_KEY_DOES_NOT_EXIST", e.ErrorCode)
		assert.Equal(t, 1, e.ErrorID)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		e := newInsufficientBalanceError("ERROR_ZERO_BALANCE", "no funds")
		assert.Equal(t, "no funds", e.Error())
		assert.Equal(t, "ERROR_ZERO_BALANCE", e.ErrorCode)
	})

	t.Run("validation", func(t *testing.T) {
		e := newValidationError("ERROR_INVALID_TASK_DATA", "bad data")
		assert.Equal(t, "bad data", e.Error())
		assert.Equal(t, "ERROR_INVALID_TASK_DATA", e.ErrorCode)
	})

	t.Run("task not found", func(t *testing.T) {
		e := newTaskNotFoundError("ERROR_TASK_NOT_FOUND", "missing")
		assert.Equal(t, "missing", e.Error())
		assert.Equal(t, 16, e.ErrorID)
	})

	t.Run("solver", func(t *testing.T) {
		e := newSolverError("could not solve")
		assert.Equal(t, "could not solve", e.Error())
		assert.Equal(t, "SOLVER_FAILED", e.ErrorCode)
	})

	t.Run("timeout", func(t *testing.T) {
		e := newTimeoutError()
		assert.Equal(t, "Task solving timed out", e.Error())
		assert.Equal(t, "TIMEOUT", e.ErrorCode)
	})

	t.Run("internal", func(t *testing.T) {
		e := newInternalError("ERROR_INTERNAL", "boom")
		assert.Equal(t, "boom", e.Error())
		assert.Equal(t, "ERROR_INTERNAL", e.ErrorCode)
		assert.Equal(t, 1, e.ErrorID)
	})
}

// TestErrNetwork covers ErrNetwork.Error() and Unwrap().
func TestErrNetwork(t *testing.T) {
	inner := errors.New("dial tcp: connection refused")
	e := &ErrNetwork{Message: "Connection failed", Err: inner}

	assert.Equal(t, "Network error: Connection failed", e.Error())
	assert.Equal(t, inner, e.Unwrap())
	assert.ErrorIs(t, e, inner)
}

// TestErrGateway covers ErrGateway.Error().
func TestErrGateway(t *testing.T) {
	e := &ErrGateway{StatusCode: 503, Message: "service unavailable"}
	assert.Equal(t, "Gateway error: HTTP 503 - service unavailable", e.Error())
}

// TestErrServer covers ErrServer.Error().
func TestErrServer(t *testing.T) {
	e := &ErrServer{StatusCode: 500, Message: "internal"}
	assert.Equal(t, "Server error: HTTP 500 - internal", e.Error())
}

// TestErrRateLimit covers ErrRateLimit.Error().
func TestErrRateLimit(t *testing.T) {
	e := &ErrRateLimit{Message: "too many requests"}
	assert.Equal(t, "Rate limit exceeded: too many requests", e.Error())
}

// TestErrParse covers ErrParse.Error() and Unwrap().
func TestErrParse(t *testing.T) {
	inner := errors.New("unexpected end of JSON input")
	e := &ErrParse{Message: "Failed to parse", Err: inner}

	assert.Equal(t, "Parse error: Failed to parse", e.Error())
	assert.Equal(t, inner, e.Unwrap())
	assert.ErrorIs(t, e, inner)
}
