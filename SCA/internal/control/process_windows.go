//go:build windows

package control

import (
	"os/exec"
	"strconv"
)

func configureProcess(command *exec.Cmd) {
}

func terminateProcessTree(command *exec.Cmd) {
	if command.Process == nil {
		return
	}
	_ = exec.Command("taskkill", "/PID", strconv.Itoa(command.Process.Pid), "/T", "/F").Run()
}
