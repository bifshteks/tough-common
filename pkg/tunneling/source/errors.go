package source

import (
	"fmt"
	"strings"
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

func IsClosedConnError(err error) bool {
	str := err.Error()
	return strings.Contains(str, "use of closed network connection")
}
