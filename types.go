package capbypass

// Task type constants for CapBypass API.
const (
	// AWS WAF task types
	TaskTypeAntiAwsWaf          = "AntiAwsWafTask"
	TaskTypeAntiAwsWafProxyLess = "AntiAwsWafTaskProxyLess"

	// reCAPTCHA v2 task types
	TaskTypeReCaptchaV2          = "ReCaptchaV2Task"
	TaskTypeReCaptchaV2ProxyLess = "ReCaptchaV2TaskProxyLess"

	// reCAPTCHA v3 task types
	TaskTypeReCaptchaV3          = "ReCaptchaV3Task"
	TaskTypeReCaptchaV3ProxyLess = "ReCaptchaV3TaskProxyLess"

	// reCAPTCHA v3 Enterprise task types
	TaskTypeReCaptchaV3Enterprise          = "ReCaptchaV3EnterpriseTask"
	TaskTypeReCaptchaV3EnterpriseProxyLess = "ReCaptchaV3EnterpriseTaskProxyLess"
)

// Task represents a CAPTCHA solving task.
type Task map[string]interface{}

// TaskResult represents the response from getTaskResult.
type TaskResult struct {
	ErrorID          int                    `json:"errorId"`
	ErrorCode        string                 `json:"errorCode,omitempty"`
	ErrorDescription string                 `json:"errorDescription,omitempty"`
	Status           string                 `json:"status,omitempty"`
	Solution         map[string]interface{} `json:"solution,omitempty"`
}

// CreateTaskRequest represents the request for createTask.
type CreateTaskRequest struct {
	ClientKey string `json:"clientKey"`
	Task      Task   `json:"task"`
}

// CreateTaskResponse represents the response from createTask.
type CreateTaskResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
	TaskID           string `json:"taskId,omitempty"`
}

// GetTaskResultRequest represents the request for getTaskResult.
type GetTaskResultRequest struct {
	ClientKey string `json:"clientKey"`
	TaskID    string `json:"taskId"`
}

// GetBalanceRequest represents the request for getBalance.
type GetBalanceRequest struct {
	ClientKey string `json:"clientKey"`
}

// GetBalanceResponse represents the response from getBalance.
type GetBalanceResponse struct {
	ErrorID          int     `json:"errorId"`
	ErrorCode        string  `json:"errorCode,omitempty"`
	ErrorDescription string  `json:"errorDescription,omitempty"`
	Balance          float64 `json:"balance,omitempty"`
}
