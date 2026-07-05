package model

type ErrorCode string

const (
	CodeBadRequest      ErrorCode = "bad_request"
	CodeUnauthorized    ErrorCode = "unauthorized"
	CodeNotFound        ErrorCode = "not_found"
	CodeConflict        ErrorCode = "conflict"
	CodeValidationError ErrorCode = "validation_error"
	CodeInternalError   ErrorCode = "internal_error"
)

type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details []string  `json:"details,omitempty"`
}
