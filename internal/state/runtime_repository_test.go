package state

import "testing"

func TestQueryRuntimeStates(t *testing.T) {
	states := []RuntimeState{
		{SessionID: "s1", ProjectPath: "/a", RuntimeSurface: "cli", CommandSurface: "repl", HydratedAtUnix: 10, StartupCount: 1},
		{SessionID: "s2", ProjectPath: "/b", RuntimeSurface: "ide", CommandSurface: "panel", HydratedAtUnix: 20, StartupCount: 4, LastError: "boom"},
	}
	out := QueryRuntimeStates(states, RuntimeQuery{ProjectPathContains: "/b", MinimumStartupCount: 2, LastErrorContains: "boo", Limit: 10})
	if len(out) != 1 || out[0].SessionID != "s2" {
		t.Fatalf("unexpected query result: %+v", out)
	}
}
