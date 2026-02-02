package capbypass

import "fmt"

// APIError represents a CapBypass API error with Capsolver-compatible structure.
type APIError struct {
	ErrorID          int    `json:"errorId,omitempty"`
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

func (e *APIError) Error() string {
	if e.ErrorDescription != "" {
		return e.ErrorDescription
	}
	return e.ErrorCode
}

// Capsolver API Errors

// ErrAuthentication indicates invalid API key.
type ErrAuthentication struct {
	APIError
}

func newAuthenticationError(code, desc string) *ErrAuthentication {
	return &ErrAuthentication{APIError{ErrorID: 1, ErrorCode: code, ErrorDescription: desc}}
}

// ErrInsufficientBalance indicates zero or insufficient account balance.
type ErrInsufficientBalance struct {
	APIError
}

func newInsufficientBalanceError(code, desc string) *ErrInsufficientBalance {
	return &ErrInsufficientBalance{APIError{ErrorID: 1, ErrorCode: code, ErrorDescription: desc}}
}

// ErrValidation indicates invalid task data or parameters.
type ErrValidation struct {
	APIError
}

func newValidationError(code, desc string) *ErrValidation {
	return &ErrValidation{APIError{ErrorID: 1, ErrorCode: code, ErrorDescription: desc}}
}

// ErrTaskNotFound indicates the task ID does not exist.
type ErrTaskNotFound struct {
	APIError
}

func newTaskNotFoundError(code, desc string) *ErrTaskNotFound {
	return &ErrTaskNotFound{APIError{ErrorID: 16, ErrorCode: code, ErrorDescription: desc}}
}

// ErrSolver indicates the CAPTCHA could not be solved.
type ErrSolver struct {
	APIError
}

func newSolverError(desc string) *ErrSolver {
	return &ErrSolver{APIError{ErrorID: 0, ErrorCode: "SOLVER_FAILED", ErrorDescription: desc}}
}

// ErrTimeout indicates the solve operation exceeded the timeout.
type ErrTimeout struct {
	APIError
}

func newTimeoutError() *ErrTimeout {
	return &ErrTimeout{APIError{ErrorID: 0, ErrorCode: "TIMEOUT", ErrorDescription: "Task solving timed out"}}
}

// ErrInternal indicates an internal server error.
type ErrInternal struct {
	APIError
}

func newInternalError(code, desc string) *ErrInternal {
	return &ErrInternal{APIError{ErrorID: 1, ErrorCode: code, ErrorDescription: desc}}
}

// HTTP-Layer Errors

// ErrNetwork indicates a network connection error.
type ErrNetwork struct {
	Message string
	Err     error
}

func (e *ErrNetwork) Error() string {
	return fmt.Sprintf("Network error: %s", e.Message)
}

func (e *ErrNetwork) Unwrap() error {
	return e.Err
}

// ErrGateway indicates a gateway error (HTTP 502/503/504).
type ErrGateway struct {
	StatusCode int
	Message    string
}

func (e *ErrGateway) Error() string {
	return fmt.Sprintf("Gateway error: HTTP %d - %s", e.StatusCode, e.Message)
}

// ErrServer indicates a server error (HTTP 500).
type ErrServer struct {
	StatusCode int
	Message    string
}

func (e *ErrServer) Error() string {
	return fmt.Sprintf("Server error: HTTP %d - %s", e.StatusCode, e.Message)
}

// ErrRateLimit indicates rate limiting (HTTP 429).
type ErrRateLimit struct {
	Message string
}

func (e *ErrRateLimit) Error() string {
	return fmt.Sprintf("Rate limit exceeded: %s", e.Message)
}

// ErrParse indicates a JSON parsing error.
type ErrParse struct {
	Message string
	Err     error
}

func (e *ErrParse) Error() string {
	return fmt.Sprintf("Parse error: %s", e.Message)
}

func (e *ErrParse) Unwrap() error {
	return e.Err
}
