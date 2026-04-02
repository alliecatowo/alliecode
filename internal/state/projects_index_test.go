package state

import "testing"

func TestBuildProjectIndex(t *testing.T) {
	registry := ProjectsRegistry{Projects: map[string]ProjectMetadata{
		"/repo/a": {LastSessionID: "s-a", OpenCount: 3, DistinctSessionCount: 2, LastOpenedAt: 10, LastRuntimeSurface: "cli", LastProvider: "openai"},
		"/repo/b": {LastSessionID: "s-b", OpenCount: 1, DistinctSessionCount: 1, LastOpenedAt: 20, LastRuntimeSurface: "ide", LastProvider: "anthropic"},
	}}
	out := BuildProjectIndex(registry, ProjectIndexQuery{PathContains: "/repo", RuntimeSurface: "cli", MinOpenCount: 2, Provider: "openai", Limit: 5})
	if len(out) != 1 || out[0].Path != "/repo/a" {
		t.Fatalf("unexpected project index output: %+v", out)
	}
}

func TestBuildProjectIndexWithOffset(t *testing.T) {
	registry := ProjectsRegistry{Projects: map[string]ProjectMetadata{
		"/repo/a": {OpenCount: 3, LastOpenedAt: 3},
		"/repo/b": {OpenCount: 2, LastOpenedAt: 2},
	}}
	out := BuildProjectIndex(registry, ProjectIndexQuery{Offset: 1, Limit: 1})
	if len(out) != 1 || out[0].Path != "/repo/b" {
		t.Fatalf("unexpected offset index output: %+v", out)
	}
}
