package voice

import "testing"

func TestStartSessionDisabledFails(t *testing.T) {
	rt := NewLocalRuntime(nil)
	if rt.StartSession("local") {
		t.Fatalf("expected disabled runtime to reject session start")
	}
}
