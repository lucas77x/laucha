package ui

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestStartDetachedReapsChild starts a short-lived child through
// startDetached and confirms it never lingers as a zombie: /proc
// either drops the pid once the parent reaps it, or still reports it
// but not in zombie state.
func TestStartDetachedReapsChild(t *testing.T) {
	cmd := exec.Command("true")
	if err := startDetached(cmd); err != nil {
		t.Fatalf("startDetached: %v", err)
	}
	pid := cmd.Process.Pid

	deadline := time.Now().Add(2 * time.Second)
	for {
		if !isZombie(pid) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("pid %d still a zombie after timeout", pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// isZombie reports whether pid is a zombie process, reading its state
// from /proc/<pid>/stat. A missing /proc entry means the process was
// already fully reaped, which also counts as not a zombie.
func isZombie(pid int) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	// Format: "pid (comm) state ...". comm can contain spaces or
	// parens, so find the state right after the last ')'.
	closeParen := -1
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == ')' {
			closeParen = i
			break
		}
	}
	if closeParen < 0 || closeParen+2 >= len(data) {
		return false
	}
	return data[closeParen+2] == 'Z'
}
