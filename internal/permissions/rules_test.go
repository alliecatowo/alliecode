package permissions

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRuleMatching_RegexAndGlobDSL(t *testing.T) {
	tests := []struct {
		name  string
		rule  Rule
		tool  string
		input any
		want  Decision
	}{
		{
			name: "bash regex uses true regexp semantics",
			rule: Rule{Tool: "bash", BashRegex: "^git\\s+status$", Decision: DecisionDeny},
			tool: "bash",
			input: map[string]any{
				"command": "git status --short",
			},
			want: DecisionAsk,
		},
		{
			name: "bash regex exact match denies",
			rule: Rule{Tool: "bash", BashRegex: "^git\\s+status$", Decision: DecisionDeny},
			tool: "bash",
			input: map[string]any{
				"command": "git status",
			},
			want: DecisionDeny,
		},
		{
			name: "tool regex prefix works",
			rule: Rule{Tool: "regex:^b(ash|rowser)$", Decision: DecisionDeny},
			tool: "browser",
			input: map[string]any{
				"command": "pwd",
			},
			want: DecisionDeny,
		},
		{
			name: "tool glob prefix works",
			rule: Rule{Tool: "glob:b*", Decision: DecisionDeny},
			tool: "bash",
			input: map[string]any{
				"command": "pwd",
			},
			want: DecisionDeny,
		},
		{
			name: "file glob default matches",
			rule: Rule{Tool: "read", FileGlob: "configs/*.secret", Decision: DecisionDeny},
			tool: "read",
			input: map[string]any{
				"file_path": "configs/app.secret",
			},
			want: DecisionDeny,
		},
		{
			name: "file regex prefix matches",
			rule: Rule{Tool: "read", FileGlob: `regex:^configs/.+\.secret$`, Decision: DecisionDeny},
			tool: "read",
			input: map[string]any{
				"file_path": "configs/app.secret",
			},
			want: DecisionDeny,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("marshal input: %v", err)
			}
			engine := NewEngine(ModeDefault, []Rule{tc.rule})
			got := engine.Check(tc.tool, raw)
			if got != tc.want {
				t.Fatalf("Check() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRuleValidation_MatcherDSLErrors(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr string
	}{
		{
			name:    "empty prefixed regex is rejected",
			rule:    Rule{Tool: "bash", BashRegex: "regex:   ", Decision: DecisionDeny},
			wantErr: "regex matcher cannot be empty",
		},
		{
			name:    "empty prefixed glob is rejected",
			rule:    Rule{Tool: "glob:   ", Decision: DecisionDeny},
			wantErr: "glob matcher cannot be empty",
		},
		{
			name:    "invalid glob reports context",
			rule:    Rule{Tool: "glob:[", Decision: DecisionDeny},
			wantErr: "invalid glob pattern",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRule(tc.rule)
			if err == nil {
				t.Fatalf("validateRule() error = nil, want %q", tc.wantErr)
			}
			if got := err.Error(); !strings.Contains(got, tc.wantErr) {
				t.Fatalf("validateRule() error = %q, want substring %q", got, tc.wantErr)
			}
		})
	}
}

func TestEngineRuleSourcePrecedence(t *testing.T) {
	input, err := json.Marshal(map[string]any{"command": "git status"})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	t.Run("later rule wins within same source", func(t *testing.T) {
		engine := NewEngine(ModeDefault, []Rule{
			{Source: RuleSourceProject, Tool: "bash", Decision: DecisionAllow},
			{Source: RuleSourceProject, Tool: "bash", Decision: DecisionDeny},
		})
		if got := engine.Check("bash", input); got != DecisionDeny {
			t.Fatalf("Check() = %v, want %v", got, DecisionDeny)
		}
	})

	t.Run("session overrides policy", func(t *testing.T) {
		engine := NewEngine(ModeDefault, []Rule{
			{Source: RuleSourcePolicy, Tool: "bash", Decision: DecisionDeny},
			{Source: RuleSourceSession, Tool: "bash", Decision: DecisionAllow},
		})
		if got := engine.Check("bash", input); got != DecisionAllow {
			t.Fatalf("Check() = %v, want %v", got, DecisionAllow)
		}
	})

	t.Run("project overrides user", func(t *testing.T) {
		engine := NewEngine(ModeDefault, []Rule{
			{Source: RuleSourceUser, Tool: "bash", Decision: DecisionDeny},
			{Source: RuleSourceProject, Tool: "bash", Decision: DecisionAllow},
		})
		if got := engine.Check("bash", input); got != DecisionAllow {
			t.Fatalf("Check() = %v, want %v", got, DecisionAllow)
		}
	})

	t.Run("empty source defaults to user", func(t *testing.T) {
		engine := NewEngine(ModeDefault, []Rule{
			{Source: RuleSourceProject, Tool: "bash", Decision: DecisionDeny},
			{Tool: "bash", Decision: DecisionAllow},
		})
		if got := engine.Check("bash", input); got != DecisionDeny {
			t.Fatalf("Check() = %v, want %v", got, DecisionDeny)
		}
	})
}

type sourcePrecedenceFixture struct {
	name  string
	rules []Rule
	want  Decision
}

func TestEngineRuleSourcePrecedenceFixtures(t *testing.T) {
	input, err := json.Marshal(map[string]any{"command": "git status"})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	fixtures := []sourcePrecedenceFixture{
		{
			name:  "policy baseline applies when alone",
			rules: []Rule{{Source: RuleSourcePolicy, Tool: "bash", Decision: DecisionDeny}},
			want:  DecisionDeny,
		},
		{
			name: "user overrides policy",
			rules: []Rule{
				{Source: RuleSourcePolicy, Tool: "bash", Decision: DecisionDeny},
				{Source: RuleSourceUser, Tool: "bash", Decision: DecisionAllow},
			},
			want: DecisionAllow,
		},
		{
			name: "project overrides user",
			rules: []Rule{
				{Source: RuleSourceUser, Tool: "bash", Decision: DecisionDeny},
				{Source: RuleSourceProject, Tool: "bash", Decision: DecisionAllow},
			},
			want: DecisionAllow,
		},
		{
			name: "session overrides project",
			rules: []Rule{
				{Source: RuleSourceProject, Tool: "bash", Decision: DecisionDeny},
				{Source: RuleSourceSession, Tool: "bash", Decision: DecisionAllow},
			},
			want: DecisionAllow,
		},
		{
			name: "all sources present session takes precedence",
			rules: []Rule{
				{Source: RuleSourcePolicy, Tool: "bash", Decision: DecisionDeny},
				{Source: RuleSourceUser, Tool: "bash", Decision: DecisionDeny},
				{Source: RuleSourceProject, Tool: "bash", Decision: DecisionDeny},
				{Source: RuleSourceSession, Tool: "bash", Decision: DecisionAllow},
			},
			want: DecisionAllow,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			engine := NewEngine(ModeDefault, fixture.rules)
			if got := engine.Check("bash", input); got != fixture.want {
				t.Fatalf("Check() = %v, want %v", got, fixture.want)
			}
		})
	}
}

func TestModeSemantics_BashRiskDefaults(t *testing.T) {
	t.Run("plan mode still honors explicit rules", func(t *testing.T) {
		engine := NewEngine(ModePlan, []Rule{{Tool: "bash", Decision: DecisionAllow}})
		input, _ := json.Marshal(map[string]any{"command": "rm -rf /"})
		if got := engine.Check("bash", input); got != DecisionAllow {
			t.Fatalf("Check() = %v, want %v", got, DecisionAllow)
		}
	})

	t.Run("plan mode honors persistent decisions", func(t *testing.T) {
		engine := NewEngine(ModePlan, nil)
		input, _ := json.Marshal(map[string]any{"command": "git commit -m 'x'"})
		engine.Remember("bash", input, DecisionAllow)
		if got := engine.Check("bash", input); got != DecisionAllow {
			t.Fatalf("Check() = %v, want %v", got, DecisionAllow)
		}
	})

	t.Run("plan mode asks for non-critical bash writes", func(t *testing.T) {
		engine := NewEngine(ModePlan, nil)
		input, _ := json.Marshal(map[string]any{"command": "git commit -m 'x'"})
		if got := engine.Check("bash", input); got != DecisionAsk {
			t.Fatalf("Check() = %v, want %v", got, DecisionAsk)
		}
	})

	t.Run("plan mode preserves critical bash deny invariant", func(t *testing.T) {
		engine := NewEngine(ModePlan, nil)
		input, _ := json.Marshal(map[string]any{"command": "rm -rf /"})
		if got := engine.Check("bash", input); got != DecisionDeny {
			t.Fatalf("Check() = %v, want %v", got, DecisionDeny)
		}
	})

	t.Run("default mode denies critical bash commands", func(t *testing.T) {
		engine := NewEngine(ModeDefault, nil)
		input, _ := json.Marshal(map[string]any{"command": "rm -rf /"})
		if got := engine.Check("bash", input); got != DecisionDeny {
			t.Fatalf("Check() = %v, want %v", got, DecisionDeny)
		}
	})

	t.Run("default mode still asks for non-critical bash writes", func(t *testing.T) {
		engine := NewEngine(ModeDefault, nil)
		input, _ := json.Marshal(map[string]any{"command": "git commit -m 'x'"})
		if got := engine.Check("bash", input); got != DecisionAsk {
			t.Fatalf("Check() = %v, want %v", got, DecisionAsk)
		}
	})

	t.Run("auto mode allows low asks medium denies critical", func(t *testing.T) {
		engine := NewEngine(ModeAuto, nil)

		lowInput, _ := json.Marshal(map[string]any{"command": "git commit -m 'x'"})
		if got := engine.Check("bash", lowInput); got != DecisionAllow {
			t.Fatalf("low risk Check() = %v, want %v", got, DecisionAllow)
		}

		mediumInput, _ := json.Marshal(map[string]any{"command": "rm temp.txt"})
		if got := engine.Check("bash", mediumInput); got != DecisionAsk {
			t.Fatalf("medium risk Check() = %v, want %v", got, DecisionAsk)
		}

		criticalInput, _ := json.Marshal(map[string]any{"command": "rm -rf /"})
		if got := engine.Check("bash", criticalInput); got != DecisionDeny {
			t.Fatalf("critical risk Check() = %v, want %v", got, DecisionDeny)
		}
	})

	t.Run("accept_edits mode aliases auto semantics", func(t *testing.T) {
		engine := NewEngine(ModeAcceptEdits, nil)

		lowInput, _ := json.Marshal(map[string]any{"command": "git commit -m 'x'"})
		if got := engine.Check("bash", lowInput); got != DecisionAllow {
			t.Fatalf("low risk Check() = %v, want %v", got, DecisionAllow)
		}

		mediumInput, _ := json.Marshal(map[string]any{"command": "rm temp.txt"})
		if got := engine.Check("bash", mediumInput); got != DecisionAsk {
			t.Fatalf("medium risk Check() = %v, want %v", got, DecisionAsk)
		}

		criticalInput, _ := json.Marshal(map[string]any{"command": "rm -rf /"})
		if got := engine.Check("bash", criticalInput); got != DecisionDeny {
			t.Fatalf("critical risk Check() = %v, want %v", got, DecisionDeny)
		}
	})

	t.Run("bypass mode still respects deny rules", func(t *testing.T) {
		engine := NewEngine(ModeBypass, []Rule{{Tool: "bash", Decision: DecisionDeny}})
		input, _ := json.Marshal(map[string]any{"command": "echo hello"})
		if got := engine.Check("bash", input); got != DecisionDeny {
			t.Fatalf("Check() = %v, want %v", got, DecisionDeny)
		}
	})

	t.Run("bypass mode still respects remembered deny decisions", func(t *testing.T) {
		engine := NewEngine(ModeBypass, nil)
		input, _ := json.Marshal(map[string]any{"command": "echo hello"})
		engine.Remember("bash", input, DecisionDeny)
		if got := engine.Check("bash", input); got != DecisionDeny {
			t.Fatalf("Check() = %v, want %v", got, DecisionDeny)
		}
	})
}

func TestModeSemantics_ParseModeCompatibilityAliases(t *testing.T) {
	tests := []struct {
		input string
		want  Mode
		ok    bool
	}{
		{input: "plan", want: ModePlan, ok: true},
		{input: "default", want: ModeDefault, ok: true},
		{input: "auto", want: ModeAuto, ok: true},
		{input: "accept_edits", want: ModeAcceptEdits, ok: true},
		{input: "accept-edits", want: ModeAcceptEdits, ok: true},
		{input: "acceptedits", want: ModeAcceptEdits, ok: true},
		{input: "bypass_permissions", want: ModeBypass, ok: true},
		{input: "unknown", want: ModeDefault, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := ParseMode(tc.input)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("ParseMode(%q) = (%v, %v), want (%v, %v)", tc.input, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestRuleMatching_ShellPrefixWildcardAndExactSemantics(t *testing.T) {
	tests := []struct {
		name    string
		matcher string
		cmd     string
		want    Decision
	}{
		{name: "legacy prefix matches bare command", matcher: "git:*", cmd: "git", want: DecisionDeny},
		{name: "legacy prefix matches subcommand", matcher: "git:*", cmd: "git status", want: DecisionDeny},
		{name: "legacy prefix does not match longer token", matcher: "git:*", cmd: "gitx status", want: DecisionAsk},
		{name: "wildcard trailing args optional", matcher: "git *", cmd: "git", want: DecisionDeny},
		{name: "wildcard trailing args present", matcher: "git *", cmd: "git add .", want: DecisionDeny},
		{name: "escaped wildcard treated literally", matcher: `echo \*`, cmd: "echo *", want: DecisionDeny},
		{name: "exact non-regex shell rule", matcher: "npm run build", cmd: "npm run build", want: DecisionDeny},
		{name: "exact non-regex shell rule not prefix", matcher: "npm run build", cmd: "npm run build --watch", want: DecisionAsk},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEngine(ModeDefault, []Rule{{Tool: "bash", BashRegex: tc.matcher, Decision: DecisionDeny}})
			raw, err := json.Marshal(map[string]any{"command": tc.cmd})
			if err != nil {
				t.Fatalf("marshal input: %v", err)
			}
			if got := engine.Check("bash", raw); got != tc.want {
				t.Fatalf("Check() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestShellPathConstraintsAndDecisionReasons(t *testing.T) {
	t.Run("path outside workdir asks with reason", func(t *testing.T) {
		engine := NewEngine(ModeDefault, nil)
		raw, _ := json.Marshal(map[string]any{
			"command": "cat ../secrets.txt",
			"workdir": "/workspace/project",
		})
		got := engine.CheckDetailed("bash", raw)
		if got.Decision != DecisionAsk {
			t.Fatalf("Decision = %v, want %v", got.Decision, DecisionAsk)
		}
		if !strings.Contains(got.Reason, "outside allowed working directory") {
			t.Fatalf("Reason = %q, want outside working directory explanation", got.Reason)
		}
	})

	t.Run("dynamic redirection asks with reason", func(t *testing.T) {
		engine := NewEngine(ModeDefault, nil)
		raw, _ := json.Marshal(map[string]any{
			"command": "echo hi > $OUT",
			"workdir": "/workspace/project",
		})
		got := engine.CheckDetailed("bash", raw)
		if got.Decision != DecisionAsk {
			t.Fatalf("Decision = %v, want %v", got.Decision, DecisionAsk)
		}
		if !strings.Contains(got.Reason, "dynamic shell expansion") {
			t.Fatalf("Reason = %q, want dynamic redirection explanation", got.Reason)
		}
	})

	t.Run("sensitive redirection asks with reason", func(t *testing.T) {
		engine := NewEngine(ModeDefault, nil)
		raw, _ := json.Marshal(map[string]any{
			"command": "echo x > /etc/hosts",
			"workdir": "/workspace/project",
		})
		got := engine.CheckDetailed("bash", raw)
		if got.Decision != DecisionAsk {
			t.Fatalf("Decision = %v, want %v", got.Decision, DecisionAsk)
		}
		if !strings.Contains(got.Reason, "redirection target requires approval") {
			t.Fatalf("Reason = %q, want redirection constraint explanation", got.Reason)
		}
	})
}

func TestMatcherGlobDoubleStarSupport(t *testing.T) {
	engine := NewEngine(ModeDefault, []Rule{{Tool: "read", FileGlob: "**/*.secret", Decision: DecisionDeny}})
	raw, _ := json.Marshal(map[string]any{"file_path": "configs/nested/app.secret"})
	if got := engine.Check("read", raw); got != DecisionDeny {
		t.Fatalf("Check() = %v, want %v", got, DecisionDeny)
	}
}

func TestCheckDetailedIncludesDecisionCode(t *testing.T) {
	engine := NewEngine(ModeAuto, nil)
	raw, _ := json.Marshal(map[string]any{"command": "git commit -m 'x'"})
	result := engine.CheckDetailed("bash", raw)
	if result.Decision != DecisionAllow {
		t.Fatalf("Decision = %v, want %v", result.Decision, DecisionAllow)
	}
	if result.Code == "" {
		t.Fatalf("expected non-empty decision code")
	}
}
