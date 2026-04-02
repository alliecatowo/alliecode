package commands

import "testing"

func TestWave5MetadataFamilies(t *testing.T) {
	config := commandMetadataForName("config")
	if config.Category != "configuration" || config.ArgumentHint == "" {
		t.Fatalf("unexpected config metadata: %#v", config)
	}
	mcp := commandMetadataForName("mcp")
	if mcp.ArgumentHint == "" || len(mcp.Diagnostics) == 0 {
		t.Fatalf("unexpected mcp metadata: %#v", mcp)
	}
	perm := commandMetadataForName("permissions")
	if perm.ArgumentHint == "" || len(perm.Shortcuts) == 0 {
		t.Fatalf("unexpected permissions metadata: %#v", perm)
	}
}
