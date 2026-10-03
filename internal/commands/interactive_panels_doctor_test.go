package commands

import "testing"

func TestInteractivePanelDoctorIncludesHumanFixAndJSON(t *testing.T) {
	panel := requirePanel(t, "doctor", &RuntimeState{})
	want := []string{"/doctor", "/doctor fix", "/doctor json"}
	for _, apply := range want {
		found := false
		for _, item := range panel.Items {
			if item.ApplyInput == apply {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected doctor panel action %q", apply)
		}
	}
}
