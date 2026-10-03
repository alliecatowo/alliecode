package commands

import (
	"context"
	"strings"
	"testing"
)

type testCommand struct {
	name    string
	aliases []string
}

func (c testCommand) Name() string        { return c.name }
func (c testCommand) Aliases() []string   { return c.aliases }
func (c testCommand) Description() string { return "test" }
func (c testCommand) Usage() string       { return "/" + c.name }
func (c testCommand) Execute(context.Context, Context, Invocation) (Result, error) {
	return Result{Handled: true, Message: c.name}, nil
}

func TestRegistryLookupByAlias(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(testCommand{name: "help", aliases: []string{"?"}}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, ok := r.Lookup("?")
	if !ok {
		t.Fatalf("expected alias lookup to succeed")
	}
}

func TestRegistryDispatch(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(testCommand{name: "model"}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	res, err := r.Dispatch(context.Background(), Context{}, "/model")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if !res.Handled {
		t.Fatalf("expected result to be handled")
	}
	if res.Message != "model" {
		t.Fatalf("expected message model, got %q", res.Message)
	}
}

func TestDefaultRegistryIncludesParityCommands(t *testing.T) {
	r := DefaultRegistry()
	expected := []string{
		"advisor",
		"btw",
		"chrome",
		"color",
		"desktop",
		"mobile",
		"fast",
		"effort",
		"plugin",
		"reload-plugins",
		"export",
		"extra-usage",
		"rate-limit-options",
		"pr-comments",
		"web-setup",
		"exit",
		"plan",
		"review",
		"session",
		"skills",
		"rewind",
		"tag",
		"remote-env",
		"security-review",
		"login",
		"logout",
		"provider",
		"branch",
		"add-dir",
		"agents",
		"buddy",
		"clear",
		"files",
		"history",
		"diff",
		"cost",
		"doctor",
		"keybindings",
		"mcp",
		"vim",
		"voice",
		"statusline",
		"ide",
		"theme",
		"output-style",
		"status",
		"stats",
		"memory",
		"privacy-settings",
		"upgrade",
		"terminal-setup",
		"release-notes",
		"install-github-app",
		"install-slack-app",
		"feedback",
		"hooks",
		"sandbox",
		"tasks",
		"config",
		"init",
		"copy",
		"version",
		"usage",
		"context",
		"permissions",
		"resume",
		"issue",
		"workflows",
		"proactive",
		"assistant",
		"share",
		"oauth-refresh",
		"bridge",
		"ant-trace",
		"autofix-pr",
		"backfill-sessions",
		"break-cache",
		"bughunter",
		"ctx-viz",
		"debug-tool-call",
		"good-claude",
		"heapdump",
		"install",
		"mock-limits",
		"onboarding",
		"perf-issue",
		"sandbox-toggle",
		"remote-setup",
		"ultraplan",
	}
	for _, name := range expected {
		if _, ok := r.Lookup(name); !ok {
			t.Fatalf("expected command /%s to be registered", name)
		}
	}
}

func TestDefaultRegistryAliasParityWaveB(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	ctx := Context{State: state}

	tests := []struct {
		input string
		want  string
	}{
		{input: "/continue latest", want: "RESUME_REQUEST\ntarget=latest\nrequested=true\ncount=1\nlast_target=latest"},
		{input: "/reset all", want: "CLEAR_RESULT\nscope=all\nclear_display=true\nclear_context=true\nclear_diff=true\ncount=1\ncontext_clears=1\ndiff_clears=1"},
		{input: "/new status", want: "CLEAR_STATUS\ncount=1\ncontext_clears=1\ndiff_clears=1\nlast_scope=all\ncompact_requested=false\nresume_requested=false\nlast_resume_target=-"},
		{input: "/allowed-tools list", want: "PERMISSIONS_MODES\ncount=4\nmode.1=plan\nmode.2=default\nmode.3=auto\nmode.4=bypass"},
		{input: "/bashes add ship wave-b", want: "TASKS_ADD\ntask=ship wave-b\ncount=1"},
		{input: "/bashes list", want: "TASKS_LIST\ncount=1\ncompleted=0\ntask.1=ship wave-b"},
		{input: "/quit", want: "EXIT_REQUEST\nrequested=true\ncount=1\ncommand=quit"},
		{input: "/remote status", want: "SESSION_INFO\nremote_mode=false\nsession_id=-\nsession_path=-\nurl=-\nqr_available=false\nviews=1\nhosting=false\nhost_addr=-\nconnected=false\nconnected_addr=-\ntoken_source=-\ntoken_prefix=-\nhost_count=0\nconnect_count=0\ndisconnect_count=0\nmode=-\nmanager_state=-\ntransport_state=-\nreconnecting=false\nreconnect_reason=-\nreconnect_error_class=-\nreconnect_count=0\nreconnect_detail=-"},
		{input: "/checkpoint before-refactor", want: "REWIND_REQUEST\ntarget=before-refactor\ncount=1\nrequested=true"},
		{input: "/app status", want: "DESKTOP_STATUS\ncount=0\nlast_target=-"},
		{input: "/ios", want: "MOBILE_QR\nplatform=ios\ncount=1"},
		{input: "/plugins list", want: "PLUGIN_LIST\ninstalled=0\nenabled=0\npending_reload=false"},
		{input: "/marketplace", want: "PLUGIN_MARKETPLACES\ncount=0"},
		{input: "/issues status", want: "ISSUE_STATUS\nprovider=-\ncount=0\nopen=0\nclosed=0\nlast_action=-"},
		{input: "/workflow status", want: "WORKFLOWS_STATUS\ncount=0\nrunning=0\nfailed=0\nlast_action=-"},
		{input: "/pro status", want: "PROACTIVE_STATUS\nenabled=false\nrules=0\nlast_action=-"},
		{input: "/assist status", want: "ASSISTANT_STATUS\nmode=chat\nsession_id=-\nlast_action=-"},
		{input: "/publish status", want: "SHARE_STATUS\ncount=0\nactive=0\nrevoked=0\nlast_action=-"},
		{input: "/oauth status", want: "OAUTH_REFRESH_STATUS\ncount=0\nlast_action=-"},
		{input: "/provider status", want: "PROVIDER_STATUS\nprovider=-\nprovider_ready=false\nquick_fix_model=/model_<provider/model>\nquick_fix_auth=/provider set ollama\nnext=use_/provider_set_<name>_or_/login_provider_<name>"},
		{input: "/login status", want: "LOGIN_STATUS\nlogged_in=false\nprovider=-\naccount=-\nprovider_ready=false\nlogin_count=0\nlogout_count=0"},
	}

	for _, tc := range tests {
		res, err := r.Dispatch(context.Background(), ctx, tc.input)
		if err != nil {
			t.Fatalf("dispatch %q failed: %v", tc.input, err)
		}
		if res.Message != tc.want {
			t.Fatalf("unexpected output for %q:\nwant: %q\n got: %q", tc.input, tc.want, res.Message)
		}
	}
}

func TestDefaultRegistryDispatchStatefulP0Flows(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	ctx := Context{State: state}

	_, err := r.Dispatch(context.Background(), ctx, "/compact now")
	if err != nil {
		t.Fatalf("compact now failed: %v", err)
	}

	res, err := r.Dispatch(context.Background(), ctx, "/compact status")
	if err != nil {
		t.Fatalf("compact status failed: %v", err)
	}
	if res.Message != "COMPACT_STATUS\nmode=auto\nrequested=true\ncount=1\nlast_target=now" {
		t.Fatalf("unexpected compact status: %q", res.Message)
	}

	_, err = r.Dispatch(context.Background(), ctx, "/resume latest")
	if err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	res, err = r.Dispatch(context.Background(), ctx, "/resume status")
	if err != nil {
		t.Fatalf("resume status failed: %v", err)
	}
	if res.Message != "RESUME_STATUS\nrequested=true\ncount=1\nlast_target=latest" {
		t.Fatalf("unexpected resume status: %q", res.Message)
	}

	_, err = r.Dispatch(context.Background(), ctx, "/branch create feature/test")
	if err != nil {
		t.Fatalf("branch create failed: %v", err)
	}
	res, err = r.Dispatch(context.Background(), ctx, "/branch status")
	if err != nil {
		t.Fatalf("branch status failed: %v", err)
	}
	if res.Message != "BRANCH_STATUS\nactive=feature/test\ncount=2\ncreated=1\nswitches=1" {
		t.Fatalf("unexpected branch status: %q", res.Message)
	}

	_, err = r.Dispatch(context.Background(), ctx, "/diff add cmd/ac/main.go 3 1")
	if err != nil {
		t.Fatalf("diff add failed: %v", err)
	}
	res, err = r.Dispatch(context.Background(), ctx, "/clear all")
	if err != nil {
		t.Fatalf("clear all failed: %v", err)
	}
	if res.Message != "CLEAR_RESULT\nscope=all\nclear_display=true\nclear_context=true\nclear_diff=true\ncount=1\ncontext_clears=1\ndiff_clears=1" {
		t.Fatalf("unexpected clear output: %q", res.Message)
	}
	if len(state.DiffEntries) != 0 || state.ResumeRequested || state.CompactRequested {
		t.Fatalf("expected state cleared, got %+v", state)
	}
}

func TestDefaultRegistryIncludesInventoryGapCommands(t *testing.T) {
	r := DefaultRegistry()
	expected := []string{
		"bridge-kick",
		"brief",
		"commit",
		"commit-push-pr",
		"init-verifiers",
		"insights",
		"passes",
		"rename",
		"stickers",
		"think-back",
		"thinkback",
		"thinkback-play",
		"teleport",
		"summary",
		"reset-limits",
		"env",
	}
	for _, name := range expected {
		if _, ok := r.Lookup(name); !ok {
			t.Fatalf("expected command /%s to be registered", name)
		}
	}
}

func TestRegistrySuggestionsIncludesDescriptionAndFiltersByQuery(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(testCommand{name: "model", aliases: []string{"m"}}); err != nil {
		t.Fatalf("register model failed: %v", err)
	}
	if err := r.Register(testCommand{name: "memory", aliases: []string{"mem"}}); err != nil {
		t.Fatalf("register memory failed: %v", err)
	}
	if err := r.Register(testCommand{name: "help", aliases: []string{"?"}}); err != nil {
		t.Fatalf("register help failed: %v", err)
	}

	all := r.Suggestions("")
	if len(all) != 3 {
		t.Fatalf("expected all commands when query empty, got %d", len(all))
	}
	for _, item := range all {
		if strings.TrimSpace(item.Description) == "" {
			t.Fatalf("expected non-empty description for %q", item.Name)
		}
	}

	filtered := r.Suggestions("mem")
	if len(filtered) != 1 {
		t.Fatalf("expected one filtered command, got %d", len(filtered))
	}
	if filtered[0].Name != "memory" {
		t.Fatalf("expected /memory suggestion, got %q", filtered[0].Name)
	}
}

func TestRegistrySuggestionsRanksPrefixBeforeContains(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(testCommand{name: "model"}); err != nil {
		t.Fatalf("register model failed: %v", err)
	}
	if err := r.Register(testCommand{name: "demolition"}); err != nil {
		t.Fatalf("register demolition failed: %v", err)
	}

	items := r.Suggestions("mo")
	if len(items) < 2 {
		t.Fatalf("expected two matches, got %d", len(items))
	}
	if items[0].Name != "model" {
		t.Fatalf("expected prefix match ranked first, got %q", items[0].Name)
	}
}

func TestRegistrySuggestionsIncludeMetadataForDropdownUX(t *testing.T) {
	r := DefaultRegistry()
	items := r.Suggestions("provider")
	if len(items) == 0 {
		t.Fatalf("expected provider suggestions")
	}
	first := items[0]
	if first.Name != "provider" {
		t.Fatalf("expected /provider first, got %q", first.Name)
	}
	if first.Category != "configuration" {
		t.Fatalf("expected configuration category, got %q", first.Category)
	}
	if strings.TrimSpace(first.ArgumentHint) == "" {
		t.Fatalf("expected argument hint")
	}
	if len(first.Keywords) == 0 {
		t.Fatalf("expected metadata keywords")
	}
	if len(first.Examples) == 0 {
		t.Fatalf("expected metadata examples")
	}
}

func TestRegistrySuggestionsSupportKeywordAndExampleMatching(t *testing.T) {
	r := DefaultRegistry()

	keywordItems := r.Suggestions("oauth")
	if len(keywordItems) == 0 || keywordItems[0].Name != "oauth-refresh" {
		t.Fatalf("expected oauth-refresh keyword match, got %+v", keywordItems)
	}

	exampleItems := r.Suggestions("hardened-linux")
	if len(exampleItems) == 0 || exampleItems[0].Name != "remote-env" {
		t.Fatalf("expected remote-env example match, got %+v", exampleItems)
	}
}
