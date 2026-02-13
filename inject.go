package main

// assigneeCommands is the set of bd subcommands where --assignee should be
// auto-injected from the logged-in ATProto handle.
var assigneeCommands = map[string]bool{
	"create": true,
	"update": true,
	"close":  true,
	"q":      true,
}

// injectAssignee appends --assignee <handle> to the args if the subcommand
// is in assigneeCommands and --assignee/-a was not already provided.
func injectAssignee(args []string, handle string) []string {
	if len(args) == 0 || handle == "" {
		return args
	}

	// Check if the subcommand should get auto-assignee
	if !assigneeCommands[args[0]] {
		return args
	}

	// Check if --assignee or -a is already present
	for i, arg := range args {
		if arg == "--assignee" || arg == "-a" {
			return args // User explicitly set assignee
		}
		// Handle --assignee=value form
		if len(arg) > 10 && arg[:11] == "--assignee=" {
			return args
		}
		// Handle -a=value form (unlikely but defensive)
		if len(arg) > 2 && arg[:3] == "-a=" {
			return args
		}
		_ = i
	}

	// Append --assignee <handle>
	return append(args, "--assignee", handle)
}
