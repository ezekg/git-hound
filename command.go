package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
)

// A Command contains the binary executable to be run when executing commands.
type Command struct {
	Bin string
}

// Exec runs the specified command and returns its output and exit code.
func (c *Command) Exec(cmds ...string) (string, int) {
	cmd := exec.Command(c.Bin, cmds...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	reader, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			stdout.WriteString(fmt.Sprintf("%s\n", scanner.Bytes()))
		}
	}()

	if err := cmd.Start(); err != nil {
		return "Failed to spawn command\n", 1
	}

	if err := cmd.Wait(); err != nil {
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
