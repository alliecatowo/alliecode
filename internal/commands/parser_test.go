package commands

import "testing"

func TestParseSlashCommand(t *testing.T) {
	inv, err := Parse(`/model "gpt-4o mini"`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if inv.Name != "model" {
		t.Fatalf("expected name model, got %q", inv.Name)
	}
	if len(inv.Args) != 1 || inv.Args[0] != "gpt-4o mini" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}

func TestParseRejectsNonSlash(t *testing.T) {
	_, err := Parse("hello")
	if err != ErrNotSlashCommand {
		t.Fatalf("expected ErrNotSlashCommand, got %v", err)
	}
}

func TestParseUnclosedQuote(t *testing.T) {
	_, err := Parse(`/model "gpt-4o`)
	if err != ErrUnclosedQuote {
		t.Fatalf("expected ErrUnclosedQuote, got %v", err)
	}
}

func TestParseRejectsEmptySlash(t *testing.T) {
	_, err := Parse("/")
	if err != ErrEmptyCommand {
		t.Fatalf("expected ErrEmptyCommand, got %v", err)
	}
}

func TestParseSingleQuotedArg(t *testing.T) {
	inv, err := Parse(`/copy 'hello world'`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if inv.Name != "copy" {
		t.Fatalf("expected name copy, got %q", inv.Name)
	}
	if len(inv.Args) != 1 || inv.Args[0] != "hello world" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}

func TestParseEscapedWhitespaceInToken(t *testing.T) {
	inv, err := Parse(`/init my\ project`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(inv.Args) != 1 || inv.Args[0] != "my project" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}

func TestParseCommandNormalizedToLower(t *testing.T) {
	inv, err := Parse(`/PeRmIsSiOnS auto`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if inv.Name != "permissions" {
		t.Fatalf("expected normalized name permissions, got %q", inv.Name)
	}
}

func TestParseSubcommandWithQuotedArg(t *testing.T) {
	inv, err := Parse(`/mcp connect "local dev"`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if inv.Name != "mcp" {
		t.Fatalf("expected name mcp, got %q", inv.Name)
	}
	if len(inv.Args) != 2 || inv.Args[0] != "connect" || inv.Args[1] != "local dev" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}

func TestParseDashedCommandName(t *testing.T) {
	inv, err := Parse(`/add-dir ./workspace`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if inv.Name != "add-dir" {
		t.Fatalf("expected name add-dir, got %q", inv.Name)
	}
	if len(inv.Args) != 1 || inv.Args[0] != "./workspace" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}

func TestParseDoctorJSONMode(t *testing.T) {
	inv, err := Parse(`/doctor --json`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if inv.Name != "doctor" {
		t.Fatalf("expected name doctor, got %q", inv.Name)
	}
	if len(inv.Args) != 1 || inv.Args[0] != "--json" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}

func TestParseDashedPrivacySettingsCommand(t *testing.T) {
	inv, err := Parse(`/privacy-settings set telemetry off`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if inv.Name != "privacy-settings" {
		t.Fatalf("expected name privacy-settings, got %q", inv.Name)
	}
	if len(inv.Args) != 3 || inv.Args[0] != "set" || inv.Args[1] != "telemetry" || inv.Args[2] != "off" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}
