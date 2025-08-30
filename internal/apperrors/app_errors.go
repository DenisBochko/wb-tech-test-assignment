package apperrors

import (
	"errors"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrShutdown      = errors.New("shutdown error")
)
