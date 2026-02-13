package proxy

import (
	"testing"
)

func TestBuildProxyCommands(t *testing.T) {
	commands := BuildProxyCommands()

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
