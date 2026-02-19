package proxy

import (
	"testing"
)

func TestBuildProxyCommands(t *testing.T) {
	commands := BuildProxyCommands()

	// Should contain essential commands (existing + new dolt-era commands)
	want := map[string]bool{
		// Core workflow
		"init":   false,
		"list":   false,
		"ready":  false,
		"create": false,
		"close":  false,
		"show":   false,
		"update": false,
		"sync":   false,
		// Dolt-era commands (beads v0.50+)
		"vc":        false,
		"sql":       false,
		"dolt":      false,
		"mol":       false,
		"gate":      false,
		"where":     false,
		"validate":  false,
		"duplicate": false,
		"supersede": false,
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

func TestSyncCommandUsesSpecialAction(t *testing.T) {
	commands := BuildProxyCommands()

	for _, cmd := range commands {
		if cmd.Name == "sync" {
			// Verify sync has a non-nil action (SyncAction, not ProxyAction)
			if cmd.Action == nil {
				t.Error("sync command should have a non-nil action")
			}
			return
		}
	}
	t.Error("sync command not found in proxy commands")
}

func TestCommentNotProxied(t *testing.T) {
	commands := BuildProxyCommands()

	for _, cmd := range commands {
		if cmd.Name == "comment" {
			t.Error("comment should NOT be a proxy command — it conflicts with hb native comment")
		}
	}
}
