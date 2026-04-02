package integration_test

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/session"
	"github.com/alliecatowo/alliecode/internal/types"
	"github.com/alliecatowo/alliecode/tests/integration/fixtures"
)

func TestDeterministicProviderChatSyncFixture(t *testing.T) {
	provider := fixtures.NewDeterministicProvider(map[string]string{
		"ping": "pong",
	}, "fallback")

	resp, err := provider.ChatSync(context.Background(), types.ChatRequest{
		Messages: []types.Message{types.NewTextMessage(types.RoleUser, "ping")},
	})
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	if got := resp.Message.GetText(); got != "pong" {
		t.Fatalf("response text = %q, want %q", got, "pong")
	}
}

func TestDeterministicProviderStreamingSequence(t *testing.T) {
	provider := fixtures.NewDeterministicProvider(nil, "fixture-reply")

	stream, err := provider.Chat(context.Background(), types.ChatRequest{
		Messages: []types.Message{types.NewTextMessage(types.RoleUser, "hello")},
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	var events []types.StreamEvent
	for ev := range stream {
		events = append(events, ev)
	}

	if len(events) != 3 {
		t.Fatalf("len(events) = %d, want 3", len(events))
	}
	if events[0].Type != types.StreamStart {
		t.Fatalf("events[0].Type = %q, want %q", events[0].Type, types.StreamStart)
	}
	if events[1].Type != types.StreamContentDelta || events[1].Delta != "fixture-reply" {
		t.Fatalf("unexpected content event: %#v", events[1])
	}
	if events[2].Type != types.StreamMessageDone || events[2].Message == nil {
		t.Fatalf("unexpected message_done event: %#v", events[2])
	}
	if got := events[2].Message.GetText(); got != "fixture-reply" {
		t.Fatalf("message_done text = %q, want %q", got, "fixture-reply")
	}
}

func TestSessionFixtureDeterministicRoundTrip(t *testing.T) {
	fixture := fixtures.NewSessionFixture(t, "integration-fixture")

	fixtures.SeedMessages(t, fixture.Store,
		types.NewTextMessage(types.RoleUser, "hello"),
		types.NewTextMessage(types.RoleAssistant, "world"),
	)

	transcript, err := session.LoadByID(fixture.RootDir, fixture.SessionID)
	if err != nil {
		t.Fatalf("LoadByID() error = %v", err)
	}
	if transcript.SessionID != fixture.SessionID {
		t.Fatalf("SessionID = %q, want %q", transcript.SessionID, fixture.SessionID)
	}
	if len(transcript.Messages) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(transcript.Messages))
	}
}
