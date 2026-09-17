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
		// #nosec G204
		cmd = exec.Command(path)
	} else {
		// #nosec G204
		cmd = exec.Command(path, params...)

	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd
}
