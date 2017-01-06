package holo

import (
	"fmt"
	"io"
	"os"
)

var (
	// Indicates that the entity has been successfully modified to
	// be in the desired state.
	ApplyApplied = applyApplied{}

	// Indicates that the entity is already in the desired state,
	// so no changes have been made. Holo will format its output
	// accordingly (at the time of this writing, by omitting the
	// entity from the output).
	ApplyAlreadyApplied = ApplyMessage{"not changed\n"}

	// Indicates that the entity was provisioned by this plugin,
	// but has been changed by a user or external application
	// since then.  Holo will output an error message indicating
	// that "--force" is needed to overwrite these manual changes.
	ApplyExternallyChanged = ApplyMessage{"requires --force to overwrite\n"}

	// Indicateq that the entity was provisioned by this plugin,
	// but has been deleted by the user or external application
	// since then.  Holo will output an error message indicating
	// that "--force" is needed to overwrite these manual changes.
	ApplyExternallyDeleted = ApplyMessage{"requires --force to restore\n"}
)

type ApplyResult interface {
	isApplyResult()
	ExitCode() int
}

type ApplyMessage struct {
	msg string
}

type ApplyError interface {
	ApplyResult
	error
}

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
