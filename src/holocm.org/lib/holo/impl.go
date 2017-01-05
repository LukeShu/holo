package holo

import (
	"fmt"
	"io"
	"os"
)

type applyError struct {
	error
}

func (e applyError) isApplyResult() {}

func (e applyError) ExitCode() int {
	return 1
}

func NewApplyError(err error) ApplyError {
	return applyError{err}
}

func (a ApplyMessage) isApplyResult() {}

func (e ApplyMessage) ExitCode() int {
	return 0
}

func (e ApplyMessage) Send() {
	file := os.NewFile(3, "/dev/fd/3")
	_, err := io.WriteString(file, e.msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
}

type applyApplied struct{}

func (e applyApplied) isApplyResult() {}
func (e applyApplied) ExitCode() int  { return 0 }
