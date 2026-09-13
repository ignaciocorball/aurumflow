package killswitch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHaltFromConfig(t *testing.T) {
	s := New(true, filepath.Join(t.TempDir(), "missing"))
	if !s.HaltNewOrders() {
		t.Fatal("config enabled")
	}
}

func TestHaltFromEnv(t *testing.T) {
	t.Setenv(EnvName, "1")
	s := New(false, filepath.Join(t.TempDir(), "missing"))
	if !s.HaltNewOrders() {
		t.Fatal("env")
	}
}

func TestHaltFromFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "kill")
	if err := os.WriteFile(p, []byte("halt"), 0644); err != nil {
		t.Fatal(err)
	}
	s := New(false, p)
	if !s.HaltNewOrders() {
		t.Fatal("file")
	}
}

func TestResumeAndHalt(t *testing.T) {
	s := New(false, filepath.Join(t.TempDir(), "missing"))
	if s.HaltNewOrders() {
		t.Fatal("should be clear")
	}
	s.Halt()
	if !s.HaltNewOrders() {
		t.Fatal("halt")
	}
	s.Resume()
	if s.HaltNewOrders() {
		t.Fatal("resume")
	}
}
