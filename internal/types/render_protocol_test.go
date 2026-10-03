package types

import "testing"

func TestLooksLikeStructuredContract(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "empty", text: "", want: false},
		{name: "plain sentence", text: "Status looks good.", want: false},
		{name: "single line header only", text: "STATUS_REPORT", want: false},
		{name: "kv contract", text: "STATUS_REPORT\nprovider=openai\nready=true", want: true},
		{name: "header with mixed separators but kv present", text: "DOCTOR_REPORT\nsection: auth\nstatus=warn", want: true},
		{name: "lowercase header", text: "status_report\nprovider=openai", want: false},
	}
	for _, tc := range tests {
		if got := LooksLikeStructuredContract(tc.text); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}
