package client

import (
	"errors"
	"fmt"
)

// Exported errors

type CosmosNotFoundError struct {
	id string
}

func NewCosmosNotFoundError(id string) *CosmosNotFoundError {
	return &CosmosNotFoundError{id}
}

func (e *CosmosNotFoundError) Error() string {
	return fmt.Sprintf("CosmosDB %s not found", e.id)
}

// Helper function to capture errors from deferred functions

func captureErr(errPtr *error, errFunc func() error) {
	err := errFunc()
	if err == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, err)
}
