package source

import (
	"errors"
	"fmt"
)

// FatalConnectError is used when connection cannot be established and there is no chance
// that it will change (in other words the error is not temporary, so there is no need
// to retry connecting)
type FatalConnectError struct {
	Err error
}

func NewFatalConnectError(err error) *FatalConnectError {
	return &FatalConnectError{Err: err}
}

func (fatalConnErr FatalConnectError) Error() string {
	return fmt.Sprintf("fatal connect error: %s", fatalConnErr.Err)
}

var ErrNetClosing = errors.New("use of closed network connection")
