package mcp

import "testing"

func TestNormalizeResourceServerName(t *testing.T) {
	items := normalizeResourceServerName("srv", []Resource{{URI: "u1"}})
	if len(items) != 1 || items[0].ServerName != "srv" {
		t.Fatalf("unexpected normalized resources: %+v", items)
	}
}
