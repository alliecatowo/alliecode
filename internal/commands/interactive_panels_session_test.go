package commands

import "testing"

func TestInteractivePanelSessionIncludesStatusTokenAndDisconnect(t *testing.T) {
	panel := requirePanel(t, "session", &RuntimeState{SessionConnected: true, SessionToken: "abcdef1234", SessionTokenSource: "manual", SessionTokenPrefix: "abcdef..."})
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected header intents for session panel")
	}
	want := map[string]bool{"/session status": false, "/session token show": false, "/session disconnect": false}
	for _, item := range panel.Items {
		if _, ok := want[item.ApplyInput]; ok {
			want[item.ApplyInput] = true
		}
	}
	for apply, ok := range want {
		if !ok {
			t.Fatalf("expected session panel action %q", apply)
		}
	}
	for _, item := range panel.Items {
		if item.ApplyInput == "/session token show" && len(item.PreviewIntents) == 0 {
			t.Fatalf("expected structured token preview for session panel")
		}
	}
}
