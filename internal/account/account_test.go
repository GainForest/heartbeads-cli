package account

import "testing"

func TestCmdAccountNotNil(t *testing.T) {
	if CmdAccount == nil {
		t.Fatal("CmdAccount should not be nil")
	}
}

func TestCmdAccountSubcommands(t *testing.T) {
	// Verify login, logout, status subcommands exist
	names := make(map[string]bool)
	for _, cmd := range CmdAccount.Commands {
		names[cmd.Name] = true
	}
	for _, want := range []string{"login", "logout", "status"} {
		if !names[want] {
			t.Errorf("missing subcommand: %s", want)
		}
	}
}
