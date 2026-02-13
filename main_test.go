package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "version flag prints hb version",
			args:       []string{"hb", "--version"},
			wantErr:    false,
			wantOutput: "hb version",
		},
		{
			name:       "no args shows help",
			args:       []string{"hb"},
			wantErr:    false,
			wantOutput: "hb",
		},
		{
			name:       "help flag shows usage",
			args:       []string{"hb", "--help"},
			wantErr:    false,
			wantOutput: "Authenticated beads CLI for AI agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runWithOutput(tt.args, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("runWithOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q does not contain %q", output, tt.wantOutput)
			}
		})
	}
}
