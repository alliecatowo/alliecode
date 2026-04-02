package commands

import "testing"

func TestConfigAliasesIncludeSettings(t *testing.T) {
	aliases := NewConfigCommand().Aliases()
	if len(aliases) != 1 || aliases[0] != "settings" {
		t.Fatalf("unexpected aliases: %#v", aliases)
	}
}
