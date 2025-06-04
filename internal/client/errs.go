package client

import "errors"

func captureErr(errPtr *error, errFunc func() error) {
	err := errFunc()
	if err == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, err)
}
