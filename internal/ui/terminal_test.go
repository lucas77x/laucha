package ui

import (
	"testing"
)

func TestTerminalCommandHonorsTerminalEnv(t *testing.T) {
	t.Setenv("TERMINAL", "myterm")

	got := terminalCommand([]string{"htop", "--tree"})

	want := []string{"myterm", "-e", "htop", "--tree"}
	if len(got) != len(want) {
		t.Fatalf("terminalCommand = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("terminalCommand[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
