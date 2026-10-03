package commands

import "testing"

func TestInteractivePanelPluginIncludesToggleAndReloadActions(t *testing.T) {
	panel := requirePanel(t, "plugin", &RuntimeState{PluginsInstalled: []string{"alpha"}, PluginsEnabled: []string{"alpha"}, PluginReloadPending: true})
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected header intents for plugin panel")
	}
	want := map[string]bool{"/reload-plugins": false, "/plugin disable alpha": false, "/plugin remove alpha": false}
	for _, item := range panel.Items {
		if _, ok := want[item.ApplyInput]; ok {
			want[item.ApplyInput] = true
		}
	}
	for apply, ok := range want {
		if !ok {
			t.Fatalf("expected plugin panel action %q", apply)
		}
	}
	for _, item := range panel.Items {
		if item.ApplyInput == "/plugin disable alpha" && len(item.PreviewIntents) == 0 {
			t.Fatalf("expected structured plugin preview")
		}
	}
}
