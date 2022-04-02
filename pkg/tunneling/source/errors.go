package source

import "fmt"

type FatalConnectError struct {
	Err error
}

func NewFatalConnectError(err error) *FatalConnectError {
	return &FatalConnectError{Err: err}
}

func (fatalConnErr FatalConnectError) Error() string {
	return fmt.Sprintf("fatal connect error: %s", fatalConnErr.Err)
}
