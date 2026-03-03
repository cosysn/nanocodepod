//go:build !linux

package agent

// reapZombiesLinux is a no-op on non-Linux systems (Windows, etc.)
func reapZombiesLinux() {
	// No-op on Windows and other platforms
	select {}
}
