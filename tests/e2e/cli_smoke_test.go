package e2e_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/session"
	"github.com/alliecatowo/alliecode/internal/types"
	"github.com/alliecatowo/alliecode/tests/integration/fixtures"
)

func TestCLISmoke_CommandDispatchAndSessionPersistence(t *testing.T) {
	state := &commands.RuntimeState{
		Model:          "llama3",
		ProviderName:   "openai",
		PermissionMode: permissions.ModeDefault,
	}
	registry := commands.DefaultRegistry()

	modelRes, err := registry.Dispatch(context.Background(), commands.Context{State: state}, "/model openai/gpt-4o-mini")
	if err != nil {
		t.Fatalf("dispatch /model error = %v", err)
	}
	if !modelRes.Handled || state.Model != "openai/gpt-4o-mini" {
		t.Fatalf("expected /model to update runtime state, got handled=%t model=%q", modelRes.Handled, state.Model)
	}

	permRes, err := registry.Dispatch(context.Background(), commands.Context{State: state}, "/permissions auto")
	if err != nil {
		t.Fatalf("dispatch /permissions error = %v", err)
	}
	if !permRes.Handled || state.PermissionMode != permissions.ModeAuto {
		t.Fatalf("expected /permissions to update mode, got handled=%t mode=%v", permRes.Handled, state.PermissionMode)
	}

	fixture := fixtures.NewSessionFixture(t, "e2e-smoke")
	if err := fixture.Store.AppendMessage(types.NewTextMessage(types.RoleUser, "/model openai/gpt-4o-mini")); err != nil {
		t.Fatalf("append user message error = %v", err)
	}
	if err := fixture.Store.AppendMessage(types.NewTextMessage(types.RoleAssistant, modelRes.Message)); err != nil {
		t.Fatalf("append assistant message error = %v", err)
	}

	replayed, err := session.LoadByID(fixture.RootDir, fixture.SessionID)
	if err != nil {
		t.Fatalf("LoadByID() error = %v", err)
	}
	if got := len(replayed.Messages); got != 2 {
		t.Fatalf("len(Messages) = %d, want 2", got)
	}
	if got := replayed.Messages[0].GetText(); got != "/model openai/gpt-4o-mini" {
		t.Fatalf("first replayed message = %q, want %q", got, "/model openai/gpt-4o-mini")
	}
	if got := replayed.Messages[1].GetText(); got != "Model set to openai/gpt-4o-mini\nCapabilities: text,image,audio,tool_use,vision,attachments\nQuick fix: /provider status\nNext: run /status to confirm runtime readiness." {
		t.Fatalf("second replayed message = %q, want %q", got, "Model set to openai/gpt-4o-mini\nCapabilities: text,image,audio,tool_use,vision,attachments\nQuick fix: /provider status\nNext: run /status to confirm runtime readiness.")
	}
}

func TestCLISmoke_UnknownSlashCommand(t *testing.T) {
	registry := commands.DefaultRegistry()
	_, err := registry.Dispatch(context.Background(), commands.Context{State: &commands.RuntimeState{}}, "/not-a-command")
	if err == nil {
		t.Fatalf("expected error for unknown slash command")
	}
	if got := err.Error(); got != "unknown slash command: /not-a-command" {
		t.Fatalf("error = %q, want %q", got, "unknown slash command: /not-a-command")
	}
}

func TestCLISmoke_StartupRepair_LoadSessionWithoutStartEvent(t *testing.T) {
	root := t.TempDir()
	sessionID := "startup-repair-42"
	path := filepath.Join(root, ".alliecode", "sessions", sessionID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	transcript := strings.Join([]string{
		"",
		`{"type":"future_event","ts":"2026-04-01T00:00:00Z","session_id":"ignored"}`,
		`{"type":"message_append","ts":"2026-04-01T00:00:01Z","message":{"role":"user","content":[{"type":"text","text":"hello"}]}}`,
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(transcript), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	replayed, err := session.LoadByPath(path)
	if err != nil {
		t.Fatalf("LoadByPath() error = %v", err)
	}
	if got := replayed.SessionID; got != sessionID {
		t.Fatalf("SessionID = %q, want %q", got, sessionID)
	}
	if got := len(replayed.Messages); got != 1 {
		t.Fatalf("len(Messages) = %d, want 1", got)
	}
	if got := replayed.Messages[0].GetText(); got != "hello" {
		t.Fatalf("first replayed message = %q, want %q", got, "hello")
	}
}

func TestCLISmoke_AutocompleteSuggestionsRankAndResolveAliases(t *testing.T) {
	registry := commands.DefaultRegistry()

	items := registry.Suggestions("mo")
	if len(items) == 0 {
		t.Fatalf("expected suggestions for query %q", "mo")
	}
	if items[0].Name != "model" {
		t.Fatalf("top suggestion = %q, want %q", items[0].Name, "model")
	}
	if items[0].Description == "" {
		t.Fatalf("expected non-empty description for %q", items[0].Name)
	}

	res, err := registry.Dispatch(context.Background(), commands.Context{State: &commands.RuntimeState{}}, "/m")
	if err != nil {
		t.Fatalf("dispatch alias /m error = %v", err)
	}
	if !res.Handled {
		t.Fatalf("dispatch alias /m returned Handled=false")
	}
	if !strings.Contains(res.Message, "Current model:") {
		t.Fatalf("alias /m response = %q, want model status payload", res.Message)
	}
}
