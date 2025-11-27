package traefik_plugin_blockip

import (
	"fmt"
)

// BlockIPError represents custom errors for the BlockIP plugin
type BlockIPError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *BlockIPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewBlockIPError creates a new BlockIPError
func NewBlockIPError(code, message string, cause error) *BlockIPError {
	return &BlockIPError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Common error codes
const (
	ErrCodeInvalidConfig   = "INVALID_CONFIG"
	ErrCodeNilHandler      = "NIL_HANDLER"
	ErrCodeInvalidStatusCode = "INVALID_STATUS_CODE"
	ErrCodeInvalidCIDR     = "INVALID_CIDR"
	ErrCodeInvalidIP       = "INVALID_IP"
	ErrCodeParseError      = "PARSE_ERROR"
	ErrCodeInternalError   = "INTERNAL_ERROR"
)

// Predefined errors
var (
	ErrConfigNil      = NewBlockIPError(ErrCodeInvalidConfig, "configuration is nil", nil)
	ErrNextHandlerNil = NewBlockIPError(ErrCodeNilHandler, "next handler is nil", nil)
)