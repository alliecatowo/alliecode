package permissions

import "testing"

func TestClassifyDenialReasonCodeWave6(t *testing.T) {
	cases := map[string]string{
		"destructive command": "destructive",
		"auth required":       "auth",
		"sensitive file":      "sensitive",
		"outside workspace":   "outside_workspace",
		"":                    "unspecified",
		"other message":       "other",
	}
	for in, want := range cases {
		if got := classifyDenialReasonCode(in); got != want {
			t.Fatalf("classify(%q)=%q want %q", in, got, want)
		}
	}
}
