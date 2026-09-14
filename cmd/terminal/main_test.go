package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestTerminalHasNoExecutionDeps(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "aurumflow/cmd/terminal").CombinedOutput()
	if err != nil {
		t.Fatalf("%s\n%s", err, out)
	}
	s := string(out)
	forbidden := []string{
		"aurumflow/internal/execution",
		"aurumflow/internal/demomut",
		"aurumflow/internal/market",
		"aurumflow/internal/mirror",
	}
	for _, pkg := range forbidden {
		if strings.Contains(s, pkg+"\n") || strings.HasSuffix(strings.TrimSpace(s), pkg) {
			t.Fatalf("forbidden dependency %s", pkg)
		}
	}
}
