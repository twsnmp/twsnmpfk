//go:build windows
// +build windows

package cmd

import (
	"os/exec"
	"syscall"
)

func GetCmd(path string, params []string) *exec.Cmd {
	var cmd *exec.Cmd
	if params == nil {
		cmd = exec.Command(path)
	} else {
		cmd = exec.Command(path, params...)

	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
