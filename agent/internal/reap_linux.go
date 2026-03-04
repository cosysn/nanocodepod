//go:build linux

package agent

import (
	"syscall"
	"time"
)

// reapZombiesLinux reaps zombie processes on Linux
// This is essential when running as PID 0 (init process)
func reapZombiesLinux() {
	for {
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if pid <= 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}
