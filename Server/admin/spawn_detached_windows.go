//go:build windows

package admin

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// spawnDetached starts a new process that is not attached to the current one.
func spawnDetached(exePath string, args []string) error {
	cmd := exec.Command(exePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x00000008, // DETACHED_PROCESS
		}
	}

	return cmd.Start()
}
