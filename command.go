package main

import (
	"bytes"
	"os/exec"
	"syscall"
)

// A Command contains the binary executable to be run when executing commands.
type Command struct {
	bin string
}

// Exec runs the specified command and returns its output and exit code.
func (c *Command) Exec(cmds ...string) (string, int) {
	cmd := exec.Command(c.bin, cmds...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		code := 1

		// Make sure we catch errors and return the correct exit code, if possible
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				code = status.ExitStatus()
			}
		}

		return stderr.String(), code
	}

	return stdout.String(), 0
}
