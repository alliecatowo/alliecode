package commands

import "testing"

func TestPermissionRetryCommandsDedupesAndSorts(t *testing.T) {
	commands := permissionRetryCommands([]PermissionDenial{{Command: "z cmd"}, {Command: "a cmd"}, {Command: "z cmd"}})
	if len(commands) != 2 || commands[0] != "a cmd" || commands[1] != "z cmd" {
		t.Fatalf("unexpected commands: %#v", commands)
	}
}
