package main

import (
	"slices"
	"testing"
)

func TestHasFlag(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		flags []string
		want  bool
	}{
		{
			name:  "exact match --reason",
			args:  []string{"--reason", "foo"},
			flags: []string{"--reason"},
			want:  true,
		},
		{
			name:  "exact match -r",
			args:  []string{"-r", "foo"},
			flags: []string{"-r"},
			want:  true,
		},
		{
			name:  "--reason=value form",
			args:  []string{"--reason=done"},
			flags: []string{"--reason"},
			want:  true,
		},
		{
			name:  "no match",
			args:  []string{"--other"},
			flags: []string{"--reason"},
			want:  false,
		},
		{
			name:  "empty args",
			args:  []string{},
			flags: []string{"--reason"},
			want:  false,
		},
		{
			name:  "multiple flags, one matches",
			args:  []string{"--assignee", "alice"},
			flags: []string{"--assignee", "-a"},
			want:  true,
		},
		{
			name:  "multiple flags, short form matches",
			args:  []string{"-a", "alice"},
			flags: []string{"--assignee", "-a"},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasFlag(tt.args, tt.flags...)
			if got != tt.want {
				t.Errorf("hasFlag(%v, %v) = %v, want %v", tt.args, tt.flags, got, tt.want)
			}
		})
	}
}

func TestGetSessionID(t *testing.T) {
	tests := []struct {
		name            string
		claudeSessionID string
		opencodeSession string
		want            string
		setClaude       bool
		setOpencode     bool
	}{
		{
			name:            "CLAUDE_SESSION_ID set",
			claudeSessionID: "claude-123",
			setClaude:       true,
			want:            "claude-123",
		},
		{
			name:            "OPENCODE_SESSION set",
			opencodeSession: "oc-456",
			setOpencode:     true,
			want:            "oc-456",
		},
		{
			name:            "both set, CLAUDE_SESSION_ID wins",
			claudeSessionID: "claude-123",
			opencodeSession: "oc-456",
			setClaude:       true,
			setOpencode:     true,
			want:            "claude-123",
		},
		{
			name: "neither set",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setClaude {
				t.Setenv("CLAUDE_SESSION_ID", tt.claudeSessionID)
			}
			if tt.setOpencode {
				t.Setenv("OPENCODE_SESSION", tt.opencodeSession)
			}
			got := getSessionID()
			if got != tt.want {
				t.Errorf("getSessionID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetLatestGitCommit(t *testing.T) {
	// We are in a git repo with commits, so this should return a non-empty string
	got := getLatestGitCommit()
	if got == "" {
		t.Error("getLatestGitCommit() returned empty string, expected non-empty (we are in a git repo)")
	}
}

func TestInjectFlags(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		handle            string
		claudeSessionID   string
		wantActor         bool
		wantAssignee      bool
		wantReason        bool
		wantSession       bool
		wantActorValue    string
		wantAssigneeValue string
		wantSessionValue  string
		checkExact        bool // if true, check exact args match
		wantExact         []string
	}{
		{
			name:           "create gets actor only, NO assignee",
			args:           []string{"create", "Fix bug"},
			handle:         "alice.bsky.social",
			wantActor:      true,
			wantAssignee:   false,
			wantActorValue: "alice.bsky.social",
		},
		{
			name:              "update gets assignee AND actor",
			args:              []string{"update", "bd-123", "--status", "in_progress"},
			handle:            "alice.bsky.social",
			wantActor:         true,
			wantAssignee:      true,
			wantActorValue:    "alice.bsky.social",
			wantAssigneeValue: "alice.bsky.social",
		},
		{
			name:           "close gets actor, reason, NO assignee",
			args:           []string{"close", "bd-123"},
			handle:         "alice.bsky.social",
			wantActor:      true,
			wantAssignee:   false,
			wantReason:     true,
			wantActorValue: "alice.bsky.social",
		},
		{
			name:           "close with explicit --reason not doubled",
			args:           []string{"close", "bd-123", "--reason", "manual"},
			handle:         "alice.bsky.social",
			wantActor:      true,
			wantAssignee:   false,
			wantReason:     false, // already present, so NOT injected
			wantActorValue: "alice.bsky.social",
		},
		{
			name:           "q gets actor, NO assignee",
			args:           []string{"q", "Quick task"},
			handle:         "alice.bsky.social",
			wantActor:      true,
			wantAssignee:   false,
			wantActorValue: "alice.bsky.social",
		},
		{
			name:           "list gets actor, NO assignee",
			args:           []string{"list"},
			handle:         "alice.bsky.social",
			wantActor:      true,
			wantAssignee:   false,
			wantActorValue: "alice.bsky.social",
		},
		{
			name:              "create with explicit --assignee not overridden",
			args:              []string{"create", "Fix bug", "--assignee", "bob"},
			handle:            "alice.bsky.social",
			wantActor:         true,
			wantAssignee:      false, // already present, so NOT injected
			wantActorValue:    "alice.bsky.social",
			wantAssigneeValue: "bob",
		},
		{
			name:           "create with explicit --actor not overridden",
			args:           []string{"create", "Fix bug", "--actor", "custom"},
			handle:         "alice.bsky.social",
			wantActor:      false, // already present, so NOT injected
			wantAssignee:   false, // create no longer auto-injects assignee
			wantActorValue: "custom",
		},
		{
			name:             "close with CLAUDE_SESSION_ID gets --session",
			args:             []string{"close", "bd-123"},
			handle:           "alice.bsky.social",
			claudeSessionID:  "sess-1",
			wantActor:        true,
			wantAssignee:     false,
			wantReason:       true,
			wantSession:      true,
			wantActorValue:   "alice.bsky.social",
			wantSessionValue: "sess-1",
		},
		{
			name:             "close with explicit --session not overridden",
			args:             []string{"close", "bd-123", "--session", "mine"},
			handle:           "alice.bsky.social",
			claudeSessionID:  "sess-1",
			wantActor:        true,
			wantAssignee:     false,
			wantReason:       true,
			wantSession:      false, // already present, so NOT injected
			wantActorValue:   "alice.bsky.social",
			wantSessionValue: "mine",
		},
		{
			name:         "empty handle unchanged",
			args:         []string{"ready"},
			handle:       "",
			wantActor:    false,
			wantAssignee: false,
			checkExact:   true,
			wantExact:    []string{"ready"},
		},
		{
			name:              "update with CLAUDE_SESSION_ID gets --session",
			args:              []string{"update", "bd-123", "--status", "in_progress"},
			handle:            "alice.bsky.social",
			claudeSessionID:   "sess-2",
			wantActor:         true,
			wantAssignee:      true,
			wantSession:       true,
			wantActorValue:    "alice.bsky.social",
			wantAssigneeValue: "alice.bsky.social",
			wantSessionValue:  "sess-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.claudeSessionID != "" {
				t.Setenv("CLAUDE_SESSION_ID", tt.claudeSessionID)
			}

			got := injectFlags(tt.args, tt.handle)

			if tt.checkExact {
				if !slices.Equal(got, tt.wantExact) {
					t.Errorf("injectFlags(%v, %q) = %v, want %v", tt.args, tt.handle, got, tt.wantExact)
				}
				return
			}

			// Check --actor
			actorIdx := slices.Index(got, "--actor")
			if tt.wantActor {
				if actorIdx == -1 {
					t.Errorf("expected --actor in result, got: %v", got)
					return
				}
				if actorIdx+1 >= len(got) || got[actorIdx+1] != tt.wantActorValue {
					t.Errorf("expected --actor %s, got: %v", tt.wantActorValue, got)
				}
			} else if tt.wantActorValue != "" {
				// User override case — check the value is preserved
				if actorIdx == -1 {
					t.Errorf("expected --actor in result (user override), got: %v", got)
					return
				}
				if actorIdx+1 >= len(got) || got[actorIdx+1] != tt.wantActorValue {
					t.Errorf("expected --actor %s (user override), got: %v", tt.wantActorValue, got)
				}
			}

			// Check --assignee
			assigneeIdx := slices.Index(got, "--assignee")
			if tt.wantAssignee {
				if assigneeIdx == -1 {
					t.Errorf("expected --assignee in result, got: %v", got)
					return
				}
				if assigneeIdx+1 >= len(got) || got[assigneeIdx+1] != tt.wantAssigneeValue {
					t.Errorf("expected --assignee %s, got: %v", tt.wantAssigneeValue, got)
				}
			} else if tt.wantAssigneeValue != "" {
				// User override case — check the value is preserved
				if assigneeIdx == -1 {
					t.Errorf("expected --assignee in result (user override), got: %v", got)
					return
				}
				if assigneeIdx+1 >= len(got) || got[assigneeIdx+1] != tt.wantAssigneeValue {
					t.Errorf("expected --assignee %s (user override), got: %v", tt.wantAssigneeValue, got)
				}
			} else {
				// Should NOT have --assignee
				if assigneeIdx != -1 {
					t.Errorf("expected NO --assignee in result, got: %v", got)
				}
			}

			// Check --reason
			reasonIdx := slices.Index(got, "--reason")
			if tt.wantReason {
				if reasonIdx == -1 {
					t.Errorf("expected --reason in result, got: %v", got)
					return
				}
				// Just check it's present and has a value
				if reasonIdx+1 >= len(got) {
					t.Errorf("expected --reason with value, got: %v", got)
				}
			} else if !tt.wantReason && slices.Contains(tt.args, "--reason") {
				// User override case — should still be present
				if reasonIdx == -1 {
					t.Errorf("expected --reason in result (user override), got: %v", got)
				}
			}

			// Check --session
			sessionIdx := slices.Index(got, "--session")
			if tt.wantSession {
				if sessionIdx == -1 {
					t.Errorf("expected --session in result, got: %v", got)
					return
				}
				if sessionIdx+1 >= len(got) || got[sessionIdx+1] != tt.wantSessionValue {
					t.Errorf("expected --session %s, got: %v", tt.wantSessionValue, got)
				}
			} else if tt.wantSessionValue != "" {
				// User override case — check the value is preserved
				if sessionIdx == -1 {
					t.Errorf("expected --session in result (user override), got: %v", got)
					return
				}
				if sessionIdx+1 >= len(got) || got[sessionIdx+1] != tt.wantSessionValue {
					t.Errorf("expected --session %s (user override), got: %v", tt.wantSessionValue, got)
				}
			}
		})
	}
}
