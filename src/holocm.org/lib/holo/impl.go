package holo

import (
	"fmt"
	"io"
	"os"
)

func applyApplied() {
	os.Exit(0)
}

func applyAlreadyApplied() {
	file := os.NewFile(3, "/dev/fd/3")
	_, err := io.WriteString(file, "not changed\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
	os.Exit(0)
}

func applyExternallyChanged() {
	file := os.NewFile(3, "/dev/fd/3")
	_, err := io.WriteString(file, "requires --force to overwrite\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
	os.Exit(0)
}

func applyExternallyDeleted() {
	file := os.NewFile(3, "/dev/fd/3")
	_, err := io.WriteString(file, "requires --force to restore\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
	os.Exit(0)
}

type applyResult func()

func (fn applyResult) isApplyResult() {}
func (fn applyResult) Exit() {
	fn()
}

func applyErr(n int) ApplyResult {
	if n < 1 {
		panic("n < 1")
	}
	return applyResult(func() { os.Exit(n) })
}
