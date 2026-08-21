package instance

import (
	"testing"
	"time"
)

func TestNotifyWithoutRunningInstance(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	if NotifyRunning() {
		t.Error("NotifyRunning = true, want false with no instance")
	}
}

func TestSecondLaunchNotifiesFirst(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	shown := make(chan struct{}, 1)
	stop, err := Listen(func() { shown <- struct{}{} })
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer stop()

	if !NotifyRunning() {
		t.Fatal("NotifyRunning = false, want true while listening")
	}
	select {
	case <-shown:
	case <-time.After(2 * time.Second):
		t.Fatal("show callback never fired")
	}
}
