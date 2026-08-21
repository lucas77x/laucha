package ui

import (
	"os"
	"os/exec"
)

// knownTerminals maps emulators to the flag that runs a command;
// checked in order, the Debian alternatives wrapper first.
var knownTerminals = []struct {
	name string
	flag string
}{
	{"x-terminal-emulator", "-e"},
	{"mate-terminal", "-x"},
	{"gnome-terminal", "--"},
	{"konsole", "-e"},
	{"xfce4-terminal", "-x"},
	{"alacritty", "-e"},
	{"kitty", "--"},
	{"xterm", "-e"},
}

// terminalCommand wraps argv so it runs inside the user's terminal
// emulator ($TERMINAL wins, then well-known ones). When none is
// found the command runs directly as a last resort.
func terminalCommand(argv []string) []string {
	if term := os.Getenv("TERMINAL"); term != "" {
		return append([]string{term, "-e"}, argv...)
	}
	for _, t := range knownTerminals {
		if _, err := exec.LookPath(t.name); err == nil {
			return append([]string{t.name, t.flag}, argv...)
		}
	}
	return argv
}
