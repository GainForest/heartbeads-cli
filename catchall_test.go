package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCatchallNoAuth(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "some-unknown-command"}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown command without auth")
	}

	if !strings.Contains(err.Error(), "Not logged in") {
		t.Errorf("expected 'Not logged in' error, got: %v", err)
	}
}

func TestCatchallHelp(t *testing.T) {
	var buf bytes.Buffer
	err := runWithOutput([]string{"hb"}, &buf)
	if err != nil {
		t.Fatalf("expected help to succeed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "account") {
		t.Errorf("help should mention 'account', got: %s", output)
	}
}
