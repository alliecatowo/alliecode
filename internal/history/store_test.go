package history

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestJSONLStoreRoundTrip(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}

	if err := store.AppendSessionStart("s1"); err != nil {
		t.Fatalf("AppendSessionStart() error = %v", err)
	}
	if err := store.AppendMessage("s1", types.NewTextMessage(types.RoleUser, "hello")); err != nil {
		t.Fatalf("AppendMessage(user) error = %v", err)
	}
	if err := store.AppendMessage("s1", types.NewTextMessage(types.RoleAssistant, "world")); err != nil {
		t.Fatalf("AppendMessage(assistant) error = %v", err)
	}

	events, err := store.SessionEvents("s1")
	if err != nil {
		t.Fatalf("SessionEvents() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("len(events) = %d, want 3", len(events))
	}
	if events[0].Type != EventSessionStart {
		t.Fatalf("events[0].Type = %q, want %q", events[0].Type, EventSessionStart)
	}
	if events[1].Message == nil || events[1].Message.GetText() != "hello" {
		t.Fatalf("events[1].Message text mismatch")
	}
	if events[2].Message == nil || events[2].Message.GetText() != "world" {
		t.Fatalf("events[2].Message text mismatch")
	}
}

func TestRecentSessionsSortedAndLimited(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}

	now := time.Now().UTC()
	if err := store.Append(Event{Type: EventSessionStart, SessionID: "older", Timestamp: now.Add(-2 * time.Hour)}); err != nil {
		t.Fatalf("Append(older start) error = %v", err)
	}
	if err := store.Append(Event{Type: EventMessage, SessionID: "older", Timestamp: now.Add(-90 * time.Minute), Message: ptrMessage(types.NewTextMessage(types.RoleUser, "o"))}); err != nil {
		t.Fatalf("Append(older msg) error = %v", err)
	}

	if err := store.Append(Event{Type: EventSessionStart, SessionID: "newer", Timestamp: now.Add(-30 * time.Minute)}); err != nil {
		t.Fatalf("Append(newer start) error = %v", err)
	}
	if err := store.Append(Event{Type: EventMessage, SessionID: "newer", Timestamp: now.Add(-5 * time.Minute), Message: ptrMessage(types.NewTextMessage(types.RoleAssistant, "n"))}); err != nil {
		t.Fatalf("Append(newer msg) error = %v", err)
	}

	recent, err := store.RecentSessions(1)
	if err != nil {
		t.Fatalf("RecentSessions() error = %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("len(recent) = %d, want 1", len(recent))
	}
	if recent[0].SessionID != "newer" {
		t.Fatalf("recent[0].SessionID = %q, want %q", recent[0].SessionID, "newer")
	}
	if recent[0].MessageCount != 1 {
		t.Fatalf("recent[0].MessageCount = %d, want 1", recent[0].MessageCount)
	}
	if recent[0].Title != "n" || recent[0].Summary != "n" {
		t.Fatalf("expected latest message display hydration, got title=%q summary=%q", recent[0].Title, recent[0].Summary)
	}
}

func ptrMessage(m types.Message) *types.Message {
	copyMessage := m
	return &copyMessage
}

func TestGetHistoryCurrentSessionFirstWithinWindow(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}

	events := []Event{
		{Type: EventCustom, SessionID: "s2", Project: "p1", Display: "o1", Timestamp: time.Unix(1, 0).UTC()},
		{Type: EventCustom, SessionID: "s1", Project: "p1", Display: "c1-old", Timestamp: time.Unix(2, 0).UTC()},
		{Type: EventCustom, SessionID: "s2", Project: "p1", Display: "o2", Timestamp: time.Unix(3, 0).UTC()},
		{Type: EventCustom, SessionID: "s1", Project: "p1", Display: "c1-new", Timestamp: time.Unix(4, 0).UTC()},
		{Type: EventCustom, SessionID: "s3", Project: "p2", Display: "ignore", Timestamp: time.Unix(5, 0).UTC()},
	}
	for _, ev := range events {
		if err := store.Append(ev); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	history, err := store.GetHistory(HistoryQuery{Project: "p1", CurrentSessionID: "s1", MaxItems: 4})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	got := make([]string, 0, len(history))
	for _, ev := range history {
		got = append(got, ev.Display)
	}
	want := []string{"c1-new", "c1-old", "o2", "o1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("history display order = %v, want %v", got, want)
	}
}

func TestGetHistoryMaxItemsUsesNewestMatchingWindow(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}

	events := []Event{
		{Type: EventCustom, SessionID: "s1", Project: "p1", Display: "older-current", Timestamp: time.Unix(1, 0).UTC()},
		{Type: EventCustom, SessionID: "s2", Project: "p1", Display: "newer-other-1", Timestamp: time.Unix(2, 0).UTC()},
		{Type: EventCustom, SessionID: "s3", Project: "p1", Display: "newer-other-2", Timestamp: time.Unix(3, 0).UTC()},
	}
	for _, ev := range events {
		if err := store.Append(ev); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	history, err := store.GetHistory(HistoryQuery{Project: "p1", CurrentSessionID: "s1", MaxItems: 2})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(history))
	}
	if history[0].Display != "newer-other-2" || history[1].Display != "newer-other-1" {
		t.Fatalf("history display order = [%q %q], want [%q %q]", history[0].Display, history[1].Display, "newer-other-2", "newer-other-1")
	}
}

func TestGetTimestampedHistoryDedupesByDisplayNewestFirst(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}

	events := []Event{
		{Type: EventCustom, SessionID: "s1", Project: "p1", Display: "dup", Timestamp: time.Unix(1, 0).UTC()},
		{Type: EventCustom, SessionID: "s2", Project: "p1", Display: "x", Timestamp: time.Unix(2, 0).UTC()},
		{Type: EventCustom, SessionID: "s3", Project: "p1", Display: "dup", Timestamp: time.Unix(3, 0).UTC()},
		{Type: EventCustom, SessionID: "s1", Project: "p1", Display: "y", Timestamp: time.Unix(4, 0).UTC()},
	}
	for _, ev := range events {
		if err := store.Append(ev); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	items, err := store.GetTimestampedHistory(HistoryQuery{Project: "p1", MaxItems: 2})
	if err != nil {
		t.Fatalf("GetTimestampedHistory() error = %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Display != "y" || items[1].Display != "dup" {
		t.Fatalf("timestamped displays = [%q %q], want [%q %q]", items[0].Display, items[1].Display, "y", "dup")
	}
}

func TestGlobalHistoryRetrievalSkipsMalformedLines(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}

	if err := store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "p1", Display: "first", Timestamp: time.Unix(1, 0).UTC()}); err != nil {
		t.Fatalf("Append(first) error = %v", err)
	}

	globalPath := filepath.Join(root, defaultHistoryDir, globalHistoryFileName)
	f, err := os.OpenFile(globalPath, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatalf("OpenFile(global) error = %v", err)
	}
	if _, err := f.WriteString("{not-json}\n"); err != nil {
		_ = f.Close()
		t.Fatalf("WriteString(malformed) error = %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close(global) error = %v", err)
	}

	if err := store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "p1", Display: "second", Timestamp: time.Unix(2, 0).UTC()}); err != nil {
		t.Fatalf("Append(second) error = %v", err)
	}

	history, err := store.GetHistory(HistoryQuery{Project: "p1", CurrentSessionID: "s1", MaxItems: 10})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(history))
	}
	if history[0].Display != "second" || history[1].Display != "first" {
		t.Fatalf("history display order = [%q %q], want [%q %q]", history[0].Display, history[1].Display, "second", "first")
	}
}

func TestQueryRecentSessionsAndHistoryFilters(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}

	if err := store.Append(Event{Type: EventCustom, SessionID: "alpha", Project: "/repo/a", Display: "deploy app", Timestamp: time.Unix(10, 0).UTC(), RuntimeSurface: "cli", CommandSurface: "repl", Provider: "openai", Model: "gpt-4o-mini", Tags: []string{"deploy", "prod"}}); err != nil {
		t.Fatalf("Append(alpha) error = %v", err)
	}
	if err := store.Append(Event{Type: EventCustom, SessionID: "beta", Project: "/repo/b", Display: "test suite", Timestamp: time.Unix(20, 0).UTC(), RuntimeSurface: "cli", CommandSurface: "batch", Provider: "anthropic", Model: "claude", Tags: []string{"test"}}); err != nil {
		t.Fatalf("Append(beta) error = %v", err)
	}

	sessions, err := store.QueryRecentSessions(SessionQuery{ProjectContains: "/repo/b", Limit: 5})
	if err != nil {
		t.Fatalf("QueryRecentSessions() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].SessionID != "beta" {
		t.Fatalf("unexpected sessions result: %+v", sessions)
	}

	history, err := store.GetHistory(HistoryQuery{Project: "/repo/a", CurrentSessionID: "alpha", ContainsDisplay: "deploy", Type: EventCustom, MaxItems: 5})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if len(history) != 1 || history[0].Display != "deploy app" {
		t.Fatalf("unexpected filtered history: %+v", history)
	}
}
