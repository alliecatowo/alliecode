package permissions

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRuleStoreCRUD(t *testing.T) {
	store := NewRuleStore(filepath.Join(t.TempDir(), "rules.json"))

	created, err := store.Create(Rule{
		Tool:      "bash",
		BashRegex: "^git\\s+status$",
		Decision:  DecisionDeny,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("Create returned empty ID")
	}

	loaded, ok, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected created rule to exist")
	}
	if loaded.BashRegex != "^git\\s+status$" {
		t.Fatalf("unexpected matcher in loaded rule: %q", loaded.BashRegex)
	}

	updated, ok, err := store.Update(created.ID, Rule{
		Tool:      "bash",
		BashRegex: "^git\\s+diff$",
		Decision:  DecisionAllow,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected update target to exist")
	}
	if updated.Decision != DecisionAllow {
		t.Fatalf("updated decision mismatch: got %v", updated.Decision)
	}

	all, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected exactly one stored rule, got %d", len(all))
	}
	if all[0].ID != created.ID {
		t.Fatalf("list should include created rule id, got %q want %q", all[0].ID, created.ID)
	}

	deleted, err := store.Delete(created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !deleted {
		t.Fatalf("expected delete to remove existing rule")
	}

	all, err = store.List()
	if err != nil {
		t.Fatalf("List after delete failed: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected no rules after delete, got %d", len(all))
	}
}

func TestRuleStoreRejectsInvalidRegex(t *testing.T) {
	store := NewRuleStore(filepath.Join(t.TempDir(), "rules.json"))
	_, err := store.Create(Rule{Tool: "bash", BashRegex: "(", Decision: DecisionDeny})
	if err == nil {
		t.Fatalf("expected invalid regex validation error")
	}
	if !strings.Contains(err.Error(), "invalid regex pattern") {
		t.Fatalf("expected contextual regex validation error, got %v", err)
	}
}

func TestRuleStoreRejectsDuplicateRule(t *testing.T) {
	store := NewRuleStore(filepath.Join(t.TempDir(), "rules.json"))
	rule := Rule{Tool: "bash", BashRegex: "^git\\s+status$", Decision: DecisionDeny}
	if _, err := store.Create(rule); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if _, err := store.Create(rule); err == nil {
		t.Fatal("expected duplicate create to fail")
	}
}

func TestRuleStoreRejectsBlankAndDuplicateIDs(t *testing.T) {
	store := NewRuleStore(filepath.Join(t.TempDir(), "rules.json"))
	first, err := store.Create(Rule{Tool: "bash", BashRegex: "^git\\s+status$", Decision: DecisionDeny})
	if err != nil {
		t.Fatalf("Create first failed: %v", err)
	}
	_, err = store.Create(Rule{Tool: "read", FileGlob: "configs/*.yaml", Decision: DecisionAllow})
	if err != nil {
		t.Fatalf("Create second failed: %v", err)
	}

	if _, _, err := store.Update("  ", Rule{Tool: "bash", BashRegex: "^git\\s+diff$", Decision: DecisionAllow}); err == nil {
		t.Fatal("expected empty update id to fail")
	}

	_, _, err = store.Update(first.ID, Rule{Tool: "read", FileGlob: "configs/*.yaml", Decision: DecisionAllow})
	if err == nil {
		t.Fatal("expected duplicate update to fail")
	}

	if ok, err := store.Delete("   "); err != nil || ok {
		t.Fatalf("Delete(blank) = (%v, %v), want (false, nil)", ok, err)
	}
}

func TestRuleStoreListIsDeterministic(t *testing.T) {
	store := NewRuleStore(filepath.Join(t.TempDir(), "rules.json"))
	one, err := store.Create(Rule{Tool: "glob", FileGlob: "*.go", Decision: DecisionAllow})
	if err != nil {
		t.Fatalf("Create one failed: %v", err)
	}
	two, err := store.Create(Rule{Tool: "bash", BashRegex: "^git\\s+status$", Decision: DecisionDeny})
	if err != nil {
		t.Fatalf("Create two failed: %v", err)
	}
	three, err := store.Create(Rule{Tool: "read", FileGlob: "configs/*.secret", Decision: DecisionDeny})
	if err != nil {
		t.Fatalf("Create three failed: %v", err)
	}

	_ = one
	_ = two
	_ = three

	first, err := store.List()
	if err != nil {
		t.Fatalf("first List failed: %v", err)
	}
	second, err := store.List()
	if err != nil {
		t.Fatalf("second List failed: %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("list lengths differ: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("non-deterministic ordering at index %d: %q vs %q", i, first[i].ID, second[i].ID)
		}
	}
}

func TestDenialsLedgerAppendAndList(t *testing.T) {
	ledger := NewDenialsLedger(filepath.Join(t.TempDir(), "denials.jsonl"))

	input, _ := json.Marshal(map[string]string{"command": "rm -rf /tmp/a"})
	if err := ledger.Append(DenialEntry{Tool: "bash", Input: input, Reason: "policy deny"}); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	entries, err := ledger.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Tool != "bash" {
		t.Fatalf("tool mismatch: got %q", entries[0].Tool)
	}
	if entries[0].Timestamp.IsZero() {
		t.Fatalf("timestamp should be auto-populated")
	}
}

func TestDenialsLedgerQueryFilterSortAndLimit(t *testing.T) {
	ledger := NewDenialsLedger(filepath.Join(t.TempDir(), "denials.jsonl"))
	base := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		tool    string
		command string
		reason  string
		at      time.Time
	}{
		{tool: "bash", command: "git push", reason: "network restricted", at: base.Add(2 * time.Minute)},
		{tool: "bash", command: "rm -rf /tmp/a", reason: "destructive", at: base.Add(1 * time.Minute)},
		{tool: "read", command: "cat secrets.txt", reason: "secret access", at: base.Add(3 * time.Minute)},
	} {
		input, _ := json.Marshal(map[string]string{"command": tc.command})
		if err := ledger.Append(DenialEntry{Tool: tc.tool, Input: input, Reason: tc.reason, Timestamp: tc.at}); err != nil {
			t.Fatalf("Append(%s) failed: %v", tc.command, err)
		}
	}

	defaultList, err := ledger.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(defaultList) != 3 {
		t.Fatalf("default List count = %d, want 3", len(defaultList))
	}
	if got := extractDenialCommand(defaultList[0].Input); got != "cat secrets.txt" {
		t.Fatalf("default List should be newest-first, got first command %q", got)
	}

	filtered, err := ledger.Query(DenialQuery{
		Tool:            "bash",
		CommandContains: "rm -rf",
		Since:           base,
		Until:           base.Add(2 * time.Minute),
		SortBy:          "command",
		Desc:            false,
	})
	if err != nil {
		t.Fatalf("Query filtered failed: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("filtered count = %d, want 1", len(filtered))
	}
	if got := extractDenialCommand(filtered[0].Input); got != "rm -rf /tmp/a" {
		t.Fatalf("filtered command = %q, want rm -rf /tmp/a", got)
	}

	limited, err := ledger.Query(DenialQuery{SortBy: "ts", Desc: true, Limit: 2})
	if err != nil {
		t.Fatalf("Query limited failed: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("limited count = %d, want 2", len(limited))
	}
}

func TestSandboxSettingsStoreLoadSaveRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sandbox.json")
	store := NewSandboxSettingsStore(path)

	defaults, err := store.Load()
	if err != nil {
		t.Fatalf("Load defaults failed: %v", err)
	}
	if defaults.Mode != "workspace-write" || defaults.WorkspaceLocked || len(defaults.ExcludedCommands) != 0 {
		t.Fatalf("unexpected defaults: %+v", defaults)
	}

	err = store.Save(SandboxSettings{
		Mode:             "READ-ONLY",
		WorkspaceLocked:  true,
		ExcludedCommands: []string{" npm run test:* ", "npm run test:*", "git push"},
	})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.Mode != "read-only" {
		t.Fatalf("mode mismatch: got %q", loaded.Mode)
	}
	if !loaded.WorkspaceLocked {
		t.Fatalf("expected workspace lock true")
	}
	if len(loaded.ExcludedCommands) != 2 || loaded.ExcludedCommands[0] != "git push" || loaded.ExcludedCommands[1] != "npm run test:*" {
		t.Fatalf("unexpected excluded commands: %#v", loaded.ExcludedCommands)
	}
}

func TestNormalizeSandboxSettingsFallbacks(t *testing.T) {
	settings := NormalizeSandboxSettings(SandboxSettings{Mode: "invalid", ExcludedCommands: []string{" ", "a", "a"}})
	if settings.Mode != "workspace-write" {
		t.Fatalf("expected fallback mode, got %q", settings.Mode)
	}
	if len(settings.ExcludedCommands) != 1 || settings.ExcludedCommands[0] != "a" {
		t.Fatalf("unexpected normalized excludes: %#v", settings.ExcludedCommands)
	}
}

func TestSandboxPolicySummaryAndAliases(t *testing.T) {
	if got := SandboxModeAlias("read-only"); got != "restrictive" {
		t.Fatalf("read-only alias = %q, want restrictive", got)
	}
	if got := SandboxModeAlias("danger-full-access"); got != "permissive" {
		t.Fatalf("danger-full-access alias = %q, want permissive", got)
	}
	if got := SandboxModeAlias("workspace-write"); got != "balanced" {
		t.Fatalf("workspace-write alias = %q, want balanced", got)
	}
	if RuleSourcePrecedence != "policy>user>project>session" {
		t.Fatalf("unexpected rule precedence: %q", RuleSourcePrecedence)
	}

	summary := EffectiveSandboxPolicySummary(SandboxSettings{Mode: "READ-ONLY", WorkspaceLocked: true, ExcludedCommands: []string{"git push", "git push", " "}})
	if summary.Mode != "read-only" {
		t.Fatalf("summary mode = %q, want read-only", summary.Mode)
	}
	if summary.ModeAlias != "restrictive" {
		t.Fatalf("summary alias = %q, want restrictive", summary.ModeAlias)
	}
	if !summary.WorkspaceLocked {
		t.Fatalf("expected workspace lock true")
	}
	if summary.ExcludedCount != 1 {
		t.Fatalf("excluded count = %d, want 1", summary.ExcludedCount)
	}
}
