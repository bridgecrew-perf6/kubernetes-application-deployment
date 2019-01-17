package utils

import (
	"bytes"
	"os/exec"
)

func RunCommand(command string, cmdArgs []string, dir string) (bytes.Buffer, bytes.Buffer, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command(command, cmdArgs...)

	cmd.Dir = dir
	//cmd.Stdin = stdin
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Run()

	if err != nil {
		return stdout, stderr, err
	}

	return stdout, stderr, err
}
