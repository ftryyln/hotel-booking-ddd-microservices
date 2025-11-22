package dto

// ErrorResponse mirrors API error contract.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// SuccessResponse is a generic success payload.
type SuccessResponse struct {
	ID      string `json:"id,omitempty"`
	Message string `json:"message"`
}
