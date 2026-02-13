package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildProxyCommands(t *testing.T) {
	commands := buildProxyCommands()

	// Should contain essential commands
	want := map[string]bool{
		"init":   false,
		"list":   false,
		"ready":  false,
		"create": false,
		"close":  false,
		"show":   false,
		"update": false,
		"sync":   false,
	}

	for _, cmd := range commands {
		if _, ok := want[cmd.Name]; ok {
			want[cmd.Name] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Errorf("expected proxy command %q not found", name)
		}
	}
}

func TestProxyAction_NotLoggedIn(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "list"}, &buf)
	if err == nil {
		t.Fatal("expected error when not logged in")
	}

	if !strings.Contains(err.Error(), "Not logged in") {
		t.Errorf("expected 'Not logged in' error, got: %v", err)
	}
}

func TestProxyHelp_NoAuthRequired(t *testing.T) {
	// Help should always work without auth
	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "--help"}, &buf)
	if err != nil {
		t.Fatalf("help should work without auth: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "account") {
		t.Errorf("help should mention 'account' command, got: %s", output)
	}
	if !strings.Contains(output, "list") {
		t.Errorf("help should mention 'list' command, got: %s", output)
	}
}
