package commands

import "testing"

func TestInteractivePanelPermissionsIncludesModesAndDenials(t *testing.T) {
	panel := requirePanel(t, "permissions", &RuntimeState{PermissionMode: 2, PermissionRules: []string{"project:allow read"}, PermissionDenials: []PermissionDenial{{ID: "1", Command: "rm", Reason: "blocked"}}})
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected header intents for permissions panel")
	}
	want := map[string]bool{"/permissions plan": false, "/permissions default": false, "/permissions auto": false, "/permissions bypass": false, "/permissions denials": false}
	for _, item := range panel.Items {
		if _, ok := want[item.ApplyInput]; ok {
			want[item.ApplyInput] = true
		}
	}
	for apply, ok := range want {
		if !ok {
			t.Fatalf("expected permissions panel action %q", apply)
		}
	}
	for _, item := range panel.Items {
		if item.ApplyInput == "/permissions denials" && len(item.PreviewIntents) == 0 {
			t.Fatalf("expected structured denials preview")
		}
	}
}
