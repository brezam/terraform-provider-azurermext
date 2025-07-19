package client

import (
	"errors"
	"fmt"
)

// Exported errors

type NotFoundError struct {
	id string
}

func NewNotFoundError(id string) *NotFoundError {
	return &NotFoundError{id}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Resource %s not found", e.id)
}

// Helper function to capture errors from deferred functions

func captureErr(errPtr *error, errFunc func() error) {
	err := errFunc()
	if err == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, err)
}
