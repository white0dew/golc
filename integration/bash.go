package integration

import (
	"context"
	"os/exec"
	"strings"
)

type BashProcess struct{}

func NewBashProcess() *BashProcess {
	return &BashProcess{}
}

func (bp *BashProcess) Run(ctx context.Context, commands []string) (string, error) {
	command := strings.Join(commands, ";")

	cmd := exec.Command("bash", "-c", command) //nolint gosec

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdoutStderr)), nil
}
