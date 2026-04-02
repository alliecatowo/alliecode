package tools

import "testing"

func TestCanonicalPermissionToolNameAliases(t *testing.T) {
	cases := map[string]string{
		"WebFetch":       "webfetch",
		"PowerShell":     "powershell",
		"FileRead":       "read",
		"mcp_auth_local": "mcp_auth_local",
	}
	for in, want := range cases {
		if got := canonicalPermissionToolName(in); got != want {
			t.Fatalf("canonicalPermissionToolName(%q)=%q want %q", in, got, want)
		}
	}
}
