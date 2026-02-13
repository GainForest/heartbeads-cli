package main

import (
	"slices"
	"testing"
)

func TestInjectAssignee(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		handle       string
		wantAssignee bool
		wantArgs     []string // if nil, just check wantAssignee
	}{
		{
			name:         "create gets assignee",
			args:         []string{"create", "Fix bug", "--priority", "1"},
			handle:       "alice.bsky.social",
			wantAssignee: true,
		},
		{
			name:         "create with explicit --assignee unchanged",
			args:         []string{"create", "Fix bug", "--assignee", "bob.bsky.social"},
			handle:       "alice.bsky.social",
			wantAssignee: false,
			wantArgs:     []string{"create", "Fix bug", "--assignee", "bob.bsky.social"},
		},
		{
			name:         "create with explicit -a unchanged",
			args:         []string{"create", "Fix bug", "-a", "bob.bsky.social"},
			handle:       "alice.bsky.social",
			wantAssignee: false,
			wantArgs:     []string{"create", "Fix bug", "-a", "bob.bsky.social"},
		},
		{
			name:         "list not in assigneeCommands",
			args:         []string{"list"},
			handle:       "alice.bsky.social",
			wantAssignee: false,
			wantArgs:     []string{"list"},
		},
		{
			name:         "close gets assignee",
			args:         []string{"close", "bd-123"},
			handle:       "alice.bsky.social",
			wantAssignee: true,
		},
		{
			name:         "update gets assignee",
			args:         []string{"update", "bd-123", "--status", "in_progress"},
			handle:       "alice.bsky.social",
			wantAssignee: true,
		},
		{
			name:         "q gets assignee",
			args:         []string{"q", "Quick task"},
			handle:       "alice.bsky.social",
			wantAssignee: true,
		},
		{
			name:         "ready not in assigneeCommands",
			args:         []string{"ready"},
			handle:       "alice.bsky.social",
			wantAssignee: false,
			wantArgs:     []string{"ready"},
		},
		{
			name:         "empty args unchanged",
			args:         []string{},
			handle:       "alice.bsky.social",
			wantAssignee: false,
			wantArgs:     []string{},
		},
		{
			name:         "empty handle no injection",
			args:         []string{"create", "Fix bug"},
			handle:       "",
			wantAssignee: false,
			wantArgs:     []string{"create", "Fix bug"},
		},
		{
			name:         "create with --assignee=value form",
			args:         []string{"create", "Fix bug", "--assignee=bob.bsky.social"},
			handle:       "alice.bsky.social",
			wantAssignee: false,
			wantArgs:     []string{"create", "Fix bug", "--assignee=bob.bsky.social"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectAssignee(tt.args, tt.handle)

			if tt.wantAssignee {
				// Should have --assignee appended
				idx := slices.Index(got, "--assignee")
				if idx == -1 {
					t.Errorf("expected --assignee in result, got: %v", got)
					return
				}
				if idx+1 >= len(got) || got[idx+1] != tt.handle {
					t.Errorf("expected --assignee %s, got: %v", tt.handle, got)
				}
			} else if tt.wantArgs != nil {
				// Should match expected args exactly
				if !slices.Equal(got, tt.wantArgs) {
					t.Errorf("args mismatch:\n  got:  %v\n  want: %v", got, tt.wantArgs)
				}
			}
		})
	}
}
