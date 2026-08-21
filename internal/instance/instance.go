// Package instance ensures only one laucha runs per session: a
// second launch tells the first one to show its bar, then exits.
package instance

import (
	"net"
	"os"
	"path/filepath"
)

const showCommand = "show\n"

// NotifyRunning reports whether another instance is already running,
// after asking it to show its bar.
func NotifyRunning() bool {
	conn, err := net.Dial("unix", socketPath())
	if err != nil {
		return false
	}
	defer conn.Close()
	_, err = conn.Write([]byte(showCommand))
	return err == nil
}

// Listen serves show requests from later launches; onShow is called
// from a background goroutine. The returned function stops listening
// and removes the socket.
func Listen(onShow func()) (func(), error) {
	path := socketPath()
	_ = os.Remove(path) // stale socket left by a crash; Dial failed on it
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, len(showCommand))
			_, _ = conn.Read(buf)
			conn.Close()
			onShow()
		}
	}()
	return func() {
		listener.Close()
		os.Remove(path)
	}, nil
}

func socketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		// Never fall back to a world-writable directory: any local
		// user could squat or signal the socket there.
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.TempDir(), "laucha.sock")
		}
		dir = filepath.Join(home, ".local", "share", "laucha")
		_ = os.MkdirAll(dir, 0o700)
	}
	return filepath.Join(dir, "laucha.sock")
}
