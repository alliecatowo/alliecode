package commands

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/remote"
	"github.com/alliecatowo/alliecode/internal/types"
)

type statusRuntimeProvider struct {
	chatCalls int
}

func TestHelpCommandRendersStructuredSuggestionsWithMetadata(t *testing.T) {
	r := DefaultRegistry()
	cmd := NewHelpCommand(r)

	res, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "help", Args: []string{"provider"}})
	if err != nil {
		t.Fatalf("help execute failed: %v", err)
	}
	if !strings.Contains(res.Message, "HELP") || !strings.Contains(res.Message, "entry.1.name=/provider") {
		t.Fatalf("unexpected help message: %q", res.Message)
	}
	if !strings.Contains(res.Message, "entry.1.category=configuration") {
		t.Fatalf("expected category metadata: %q", res.Message)
	}
	if !strings.Contains(res.Message, "entry.1.argument_hint=") {
		t.Fatalf("expected argument hint metadata: %q", res.Message)
	}
	if !strings.Contains(res.Message, "entry.1.keyword_count=") {
		t.Fatalf("expected keyword metadata: %q", res.Message)
	}
	if !strings.Contains(res.Message, "entry.1.example_count=") {
		t.Fatalf("expected example metadata: %q", res.Message)
	}
}

func TestHelpCommandSupportsKeywordSearch(t *testing.T) {
	r := DefaultRegistry()
	cmd := NewHelpCommand(r)

	res, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "help", Args: []string{"oauth"}})
	if err != nil {
		t.Fatalf("help execute failed: %v", err)
	}
	if !strings.Contains(res.Message, "entry.1.name=/oauth-refresh") {
		t.Fatalf("expected oauth-refresh result, got %q", res.Message)
	}
	if !strings.Contains(res.Message, "entry.2.match_reason=keyword") {
		t.Fatalf("expected secondary keyword match reason, got %q", res.Message)
	}
}

func (p *statusRuntimeProvider) Name() string { return "status-runtime" }

func (p *statusRuntimeProvider) Chat(_ context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	p.chatCalls++
	ch := make(chan types.StreamEvent, 2)
	ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: strings.Repeat("partial ", 12)}
	ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopMaxTokens}
	close(ch)
	return ch, nil
}

func (p *statusRuntimeProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *statusRuntimeProvider) ListModels(context.Context) ([]types.Model, error) { return nil, nil }
func (p *statusRuntimeProvider) SupportsStreaming() bool                           { return true }
func (p *statusRuntimeProvider) SupportsTools() bool                               { return false }
func (p *statusRuntimeProvider) SupportsThinking() bool                            { return false }

func TestModelCommandSetAndGet(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{Model: "llama3", ProviderName: "openai"}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"gpt-4o"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if !setRes.Handled || state.Model != "gpt-4o" || state.ProviderName != "openai" {
		t.Fatalf("model was not updated")
	}

	getRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if getRes.Message != "MODEL_STATUS\nprovider=openai\nprovider_ready=false\nmodel=gpt-4o\ncapabilities=text,image,audio,tool_use,vision,attachments\nquick_fix=/model_openai/<model>\nnext=use_/model_list_to_browse_or_/model_<provider>/<model>_to_set" {
		t.Fatalf("unexpected get message: %q", getRes.Message)
	}
}

func TestModelCommandStrictProviderValidation(t *testing.T) {
	cmd := NewModelCommand()

	stateNoProvider := &RuntimeState{}
	_, err := cmd.Execute(context.Background(), Context{State: stateNoProvider}, Invocation{Name: "model", Args: []string{"gpt-4o"}})
	if err == nil || err.Error() != "model must include provider as provider/model when no provider is configured" {
		t.Fatalf("expected missing-provider error, got %v", err)
	}

	stateWithProvider := &RuntimeState{ProviderName: "openai"}
	_, err = cmd.Execute(context.Background(), Context{State: stateWithProvider}, Invocation{Name: "model", Args: []string{"not-a-model"}})
	if err == nil || err.Error() != "unknown model \"not-a-model\" for provider \"openai\" (run /model list openai, then /model openai/<model>)" {
		t.Fatalf("expected unknown model error, got %v", err)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: stateWithProvider}, Invocation{Name: "model", Args: []string{"openai/gpt-4o"}})
	if err != nil {
		t.Fatalf("set with explicit provider failed: %v", err)
	}
	if !setRes.Handled || stateWithProvider.Model != "gpt-4o" || stateWithProvider.ProviderName != "openai" {
		t.Fatalf("expected explicit provider model to be set")
	}
}

func TestIssueCommandStatefulFlow(t *testing.T) {
	cmd := NewIssueCommand()
	state := &RuntimeState{ProviderName: "openai"}

	createRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "issue", Args: []string{"create", "Fix", "parser"}})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if createRes.Message != "ISSUE_CREATE\nid=ISSUE-1\nstatus=open\nprovider=openai\ntitle=Fix parser\ncount=1" {
		t.Fatalf("unexpected create message: %q", createRes.Message)
	}

	assignRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "issue", Args: []string{"assign", "ISSUE-1", "alice"}})
	if err != nil {
		t.Fatalf("assign failed: %v", err)
	}
	if assignRes.Message != "ISSUE_ASSIGN\nid=ISSUE-1\nassignee=alice\nstatus=open" {
		t.Fatalf("unexpected assign message: %q", assignRes.Message)
	}

	closeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "issue", Args: []string{"close", "ISSUE-1"}})
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if closeRes.Message != "ISSUE_SET_STATUS\nid=ISSUE-1\nprevious=open\nstatus=closed\nchanged=true" {
		t.Fatalf("unexpected close message: %q", closeRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "issue"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "ISSUE_STATUS\nprovider=openai\ncount=1\nopen=0\nclosed=1\nlast_action=close" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}
}

func TestWorkflowsCommandRunAndComplete(t *testing.T) {
	cmd := NewWorkflowsCommand()
	state := &RuntimeState{ProviderName: "openai"}

	runRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "workflows", Args: []string{"run", "ci"}})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(runRes.Message, "WORKFLOWS_SET\nname=ci\nstatus=running") {
		t.Fatalf("unexpected run message: %q", runRes.Message)
	}

	completeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "workflows", Args: []string{"complete", "ci"}})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	if !strings.Contains(completeRes.Message, "WORKFLOWS_SET\nname=ci\nstatus=succeeded") {
		t.Fatalf("unexpected complete message: %q", completeRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "workflows", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "WORKFLOWS_STATUS\ncount=1\nrunning=0\nfailed=0\nlast_action=complete" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}
}

func TestProactiveCommandRulesAndTrigger(t *testing.T) {
	cmd := NewProactiveCommand()
	state := &RuntimeState{}

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "proactive", Args: []string{"on"}})
	if err != nil {
		t.Fatalf("on failed: %v", err)
	}

	addRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "proactive", Args: []string{"rule", "add", "on-commit"}})
	if err != nil {
		t.Fatalf("rule add failed: %v", err)
	}
	if addRes.Message != "PROACTIVE_RULE_ADD\nrule=on-commit\nadded=true\ncount=1" {
		t.Fatalf("unexpected add message: %q", addRes.Message)
	}

	triggerRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "proactive", Args: []string{"trigger", "commit"}})
	if err != nil {
		t.Fatalf("trigger failed: %v", err)
	}
	if triggerRes.Message != "PROACTIVE_TRIGGER\nevent=commit\nenabled=true\nstatus=queued" {
		t.Fatalf("unexpected trigger message: %q", triggerRes.Message)
	}
}

func TestAssistantCommandModeSessionReset(t *testing.T) {
	cmd := NewAssistantCommand()
	state := &RuntimeState{}

	modeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "assistant", Args: []string{"mode", "review"}})
	if err != nil {
		t.Fatalf("mode failed: %v", err)
	}
	if modeRes.Message != "ASSISTANT_MODE\nmode=review" {
		t.Fatalf("unexpected mode message: %q", modeRes.Message)
	}

	sessionRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "assistant", Args: []string{"session", "sess-22"}})
	if err != nil {
		t.Fatalf("session failed: %v", err)
	}
	if sessionRes.Message != "ASSISTANT_SESSION\nsession_id=sess-22\ncreated=true" {
		t.Fatalf("unexpected session message: %q", sessionRes.Message)
	}

	resetRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "assistant", Args: []string{"reset"}})
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}
	if resetRes.Message != "ASSISTANT_RESET\nmode=chat\nsession_id=-" {
		t.Fatalf("unexpected reset message: %q", resetRes.Message)
	}
}

func TestShareCommandCreateRevoke(t *testing.T) {
	cmd := NewShareCommand()
	state := &RuntimeState{}

	createRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "share", Args: []string{"create", "session", "public"}})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if createRes.Message != "SHARE_CREATE\nid=share-1\nscope=session\nvisibility=public\nrevoked=false\nurl=https://share.example.invalid/share-1" {
		t.Fatalf("unexpected create message: %q", createRes.Message)
	}

	revokeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "share", Args: []string{"revoke", "share-1"}})
	if err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	if revokeRes.Message != "SHARE_REVOKE\nid=share-1\nrevoked=true\nurl=https://share.example.invalid/share-1" {
		t.Fatalf("unexpected revoke message: %q", revokeRes.Message)
	}
}

func TestOAuthRefreshCommandRefreshAndError(t *testing.T) {
	cmd := NewOAuthRefreshCommand()
	state := &RuntimeState{}

	refreshRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "oauth-refresh", Args: []string{"refresh", "github"}})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if refreshRes.Message != "OAUTH_REFRESH_RESULT\nprovider=github\nrefreshed=true\nexpires_in=3600\ntoken_prefix=github...\nerror=-" {
		t.Fatalf("unexpected refresh message: %q", refreshRes.Message)
	}

	errorRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "oauth-refresh", Args: []string{"error", "github", "token", "expired"}})
	if err != nil {
		t.Fatalf("error failed: %v", err)
	}
	if errorRes.Message != "OAUTH_REFRESH_RESULT\nprovider=github\nrefreshed=false\nexpires_in=0\ntoken_prefix=-\nerror=token expired" {
		t.Fatalf("unexpected error message: %q", errorRes.Message)
	}
}

func TestPermissionsCommandSetMode(t *testing.T) {
	cmd := NewPermissionsCommand()
	state := &RuntimeState{PermissionMode: permissions.ModeDefault}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"auto"}})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !res.Handled {
		t.Fatalf("expected handled result")
	}
	if state.PermissionMode != permissions.ModeAuto {
		t.Fatalf("expected mode auto, got %v", state.PermissionMode)
	}
}

func TestCompactCommandRequestsCompaction(t *testing.T) {
	cmd := NewCompactCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "compact"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !res.Handled || !state.CompactRequested {
		t.Fatalf("expected compact request to be set")
	}
	if res.Message != "COMPACT_REQUEST\nmode=auto\nrequested=true\ncount=1\nlast_target=now" {
		t.Fatalf("unexpected compact request message: %q", res.Message)
	}
}

func TestCompactCommandModes(t *testing.T) {
	cmd := NewCompactCommand()
	state := &RuntimeState{}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "compact", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "COMPACT_STATUS\nmode=auto\nrequested=false\ncount=0\nlast_target=-" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	offRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "compact", Args: []string{"off"}})
	if err != nil {
		t.Fatalf("off failed: %v", err)
	}
	if offRes.Message != "COMPACT_MODE\nmode=off\nrequested=false\ncount=0" || state.CompactMode != "off" {
		t.Fatalf("unexpected off result: %+v state=%+v", offRes, state)
	}

	autoRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "compact", Args: []string{"auto"}})
	if err != nil {
		t.Fatalf("auto failed: %v", err)
	}
	if autoRes.Message != "COMPACT_MODE\nmode=auto\nrequested=false\ncount=0" || state.CompactMode != "auto" {
		t.Fatalf("unexpected auto result: %+v state=%+v", autoRes, state)
	}
}

func TestBranchCommandUsageAndMessage(t *testing.T) {
	cmd := NewBranchCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "branch"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !res.Handled || res.Message != "BRANCH_STATUS\nactive=main\ncount=1\ncreated=0\nswitches=0" {
		t.Fatalf("unexpected result: %+v", res)
	}

	createRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "branch", Args: []string{"create", "feature/p0"}})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if createRes.Message != "BRANCH_CREATE\nname=feature/p0\nactive=feature/p0\ncount=2\ncreated=1\nswitched=true\nswitches=1" {
		t.Fatalf("unexpected create result: %q", createRes.Message)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "branch", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "BRANCH_LIST\ncount=2\nactive=feature/p0\nbranch.1=feature/p0\nbranch.2=main" {
		t.Fatalf("unexpected list result: %q", listRes.Message)
	}

	_, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "branch", Args: []string{"switch", "missing"}})
	if err == nil || err.Error() != "branch not found: missing" {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestConfigCommandShowAndSet(t *testing.T) {
	cmd := NewConfigCommand()
	state := &RuntimeState{}

	showRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config"})
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if showRes.Message != "CONFIG_SHOW\ncount=0" {
		t.Fatalf("unexpected show message: %q", showRes.Message)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"set", "theme", "light"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if setRes.Message != "CONFIG_SET\nkey=theme\nupdated=false\nvalue=light" {
		t.Fatalf("unexpected set message: %q", setRes.Message)
	}

	getRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"get", "theme"}})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if getRes.Message != "CONFIG_GET\nkey=theme\nfound=true\nvalue=light" {
		t.Fatalf("unexpected get message: %q", getRes.Message)
	}

	unsetRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"unset", "theme"}})
	if err != nil {
		t.Fatalf("unset failed: %v", err)
	}
	if unsetRes.Message != "CONFIG_UNSET\nkey=theme\nremoved=true" {
		t.Fatalf("unexpected unset message: %q", unsetRes.Message)
	}
}

func TestConfigCommandEdgeCases(t *testing.T) {
	cmd := NewConfigCommand()

	getMissingRes, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "config", Args: []string{"get", "missing"}})
	if err != nil {
		t.Fatalf("get missing failed: %v", err)
	}
	if getMissingRes.Message != "CONFIG_GET\nkey=missing\nfound=false" {
		t.Fatalf("unexpected missing get message: %q", getMissingRes.Message)
	}

	state := &RuntimeState{ConfigValues: map[string]string{"tone": "formal"}}
	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"set", "tone", "very", "casual"}})
	if err != nil {
		t.Fatalf("set with spaces failed: %v", err)
	}
	if setRes.Message != "CONFIG_SET\nkey=tone\nupdated=true\nvalue=very casual" {
		t.Fatalf("unexpected set update message: %q", setRes.Message)
	}
	if state.ConfigValues["tone"] != "very casual" {
		t.Fatalf("expected merged value, got %q", state.ConfigValues["tone"])
	}

	unsetMissingRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"unset", "missing"}})
	if err != nil {
		t.Fatalf("unset missing failed: %v", err)
	}
	if unsetMissingRes.Message != "CONFIG_UNSET\nkey=missing\nremoved=false" {
		t.Fatalf("unexpected unset missing message: %q", unsetMissingRes.Message)
	}
}

func TestInitCommandParsing(t *testing.T) {
	cmd := NewInitCommand()
	res, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "init", Args: []string{"./demo"}})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Message != "Initialization requested for ./demo." {
		t.Fatalf("unexpected message: %q", res.Message)
	}
}

func TestCopyCommandRequiresText(t *testing.T) {
	cmd := NewCopyCommand()
	_, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "copy"})
	if err == nil || err.Error() != "usage: /copy <text>" {
		t.Fatalf("expected usage error, got %v", err)
	}

	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "copy", Args: []string{"hello", "world"}})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Message != "COPY_RESULT\ntext=hello world\ncount=1\nlast_text=hello world" {
		t.Fatalf("unexpected message: %q", res.Message)
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()
	res, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "version"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Message != "VERSION_INFO\nruntime=parity\nversion=v1" {
		t.Fatalf("unexpected message: %q", res.Message)
	}
}

func TestUsageCommandWithState(t *testing.T) {
	cmd := NewUsageCommand()
	state := &RuntimeState{Model: "gpt-4o", PermissionMode: permissions.ModeAuto, CompactRequested: true}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "usage"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Message != "USAGE_SNAPSHOT\nmodel=gpt-4o\npermission=auto\ncompact_requested=true" {
		t.Fatalf("unexpected message: %q", res.Message)
	}
}

func TestContextCommandShowAndClear(t *testing.T) {
	cmd := NewContextCommand()
	state := &RuntimeState{CompactRequested: true, ResumeRequested: true}

	showRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "context"})
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if showRes.Message != "CONTEXT_STATE\ncompact_requested=true\nresume_requested=true\nlast_resume_target=-\nlast_compact_target=-\ncompact_mode=-\npermission_mode=plan\nmodel=-\nprovider=-\nworkspace_dir_count=0\nbranch_active=-\ndiff_entries=0" {
		t.Fatalf("unexpected show message: %q", showRes.Message)
	}

	clearRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "context", Args: []string{"clear"}})
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if clearRes.Message != "CONTEXT_CLEAR\ncompact_requested=false\nresume_requested=false\nlast_resume_target=-\nlast_compact_target=-" {
		t.Fatalf("unexpected clear message: %q", clearRes.Message)
	}
	if state.CompactRequested || state.ResumeRequested {
		t.Fatalf("expected flags to be cleared")
	}
}

func TestPermissionsCommandEnhancedGetSetList(t *testing.T) {
	cmd := NewPermissionsCommand()
	state := &RuntimeState{
		PermissionMode: permissions.ModeDefault,
		PermissionRules: []string{
			"allow read:internal/commands/**",
			"deny write:/tmp",
		},
		PermissionDenials: []PermissionDenial{
			{ID: "d2", Command: "rm -rf /tmp", Reason: "denied by policy"},
			{ID: "d1", Command: "scp secret", Reason: "network restricted"},
		},
	}

	getRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"get"}})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if getRes.Message != "PERMISSIONS_MODE\nmode=default" {
		t.Fatalf("unexpected get message: %q", getRes.Message)
	}

	summaryRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"summary"}})
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}
	if summaryRes.Message != "PERMISSIONS_SUMMARY\nmode=default\nmode.aliases=-\nrules.count=2\nrules.precedence=policy>user>project>session\nrules.policy=0\nrules.user=0\nrules.project=0\nrules.session=2\ndenials.count=2\ndenials.groups=2" {
		t.Fatalf("unexpected summary message: %q", summaryRes.Message)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "PERMISSIONS_MODES\ncount=4\nmode.1=plan\nmode.2=default\nmode.3=auto\nmode.4=bypass" {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}

	rulesRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules"}})
	if err != nil {
		t.Fatalf("rules failed: %v", err)
	}
	if rulesRes.Message != "PERMISSIONS_RULES\ncount=2\nrule.1.source=session\nrule.1.value=allow read:internal/commands/**\nrule.2.source=session\nrule.2.value=deny write:/tmp" {
		t.Fatalf("unexpected rules message: %q", rulesRes.Message)
	}

	denialsRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"denials"}})
	if err != nil {
		t.Fatalf("denials failed: %v", err)
	}
	if denialsRes.Message != "PERMISSIONS_DENIALS\ncount=2\ngroup_count=2\ngroup.1.reason=denied by policy\ngroup.1.count=1\ndenial.1.id=d2\ndenial.1.command=rm -rf /tmp\ndenial.1.reason=denied by policy\ngroup.2.reason=network restricted\ngroup.2.count=1\ndenial.2.id=d1\ndenial.2.command=scp secret\ndenial.2.reason=network restricted" {
		t.Fatalf("unexpected denials message: %q", denialsRes.Message)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"set", "plan"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if setRes.Message != "PERMISSIONS_SET\nmode=plan" {
		t.Fatalf("unexpected set message: %q", setRes.Message)
	}
	if state.PermissionMode != permissions.ModePlan {
		t.Fatalf("expected mode plan, got %v", state.PermissionMode)
	}
}

func TestPermissionsCommandRulesAddRemoveListStability(t *testing.T) {
	cmd := NewPermissionsCommand()
	state := &RuntimeState{}

	addRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules", "add", "allow read:/repo"}})
	if err != nil {
		t.Fatalf("rules add failed: %v", err)
	}
	if addRes.Message != "PERMISSIONS_RULES_ADD\nrule=allow read:/repo\ncount=1" {
		t.Fatalf("unexpected rules add message: %q", addRes.Message)
	}

	_, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules", "add", "allow read:/repo"}})
	if err != nil {
		t.Fatalf("duplicate rules add failed: %v", err)
	}
	if len(state.PermissionRules) != 1 {
		t.Fatalf("expected deduplicated rules, got %#v", state.PermissionRules)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules", "list"}})
	if err != nil {
		t.Fatalf("rules list failed: %v", err)
	}
	if listRes.Message != "PERMISSIONS_RULES\ncount=1\nrule.1.source=session\nrule.1.value=allow read:/repo" {
		t.Fatalf("unexpected rules list message: %q", listRes.Message)
	}

	removeMissingRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules", "remove", "deny write:/tmp"}})
	if err != nil {
		t.Fatalf("rules remove missing failed: %v", err)
	}
	if removeMissingRes.Message != "PERMISSIONS_RULES_REMOVE\nrule=deny write:/tmp\nremoved=false\ncount=1" {
		t.Fatalf("unexpected remove missing message: %q", removeMissingRes.Message)
	}

	removeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules", "remove", "allow read:/repo"}})
	if err != nil {
		t.Fatalf("rules remove failed: %v", err)
	}
	if removeRes.Message != "PERMISSIONS_RULES_REMOVE\nrule=allow read:/repo\nremoved=true\ncount=0" {
		t.Fatalf("unexpected rules remove message: %q", removeRes.Message)
	}

	emptyListRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules"}})
	if err != nil {
		t.Fatalf("rules empty list failed: %v", err)
	}
	if emptyListRes.Message != "PERMISSIONS_RULES\ncount=0" {
		t.Fatalf("unexpected empty rules message: %q", emptyListRes.Message)
	}
}

func TestDiffCostDoctorCommands(t *testing.T) {
	tests := []struct {
		name string
		cmd  Command
		want string
	}{
		{name: "diff", cmd: NewDiffCommand(), want: "DIFF_LIST\nmode=working\ncount=0"},
		{name: "cost", cmd: NewCostCommand(), want: "COST_BREAKDOWN\ninput_tokens=0\noutput_tokens=0\ncache_read_tokens=0\ncache_write_tokens=0"},
	}

	for _, tc := range tests {
		res, err := tc.cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: tc.name})
		if err != nil {
			t.Fatalf("%s execute failed: %v", tc.name, err)
		}
		if res.Message != tc.want {
			t.Fatalf("%s unexpected message: %q", tc.name, res.Message)
		}
	}
}

func TestCostCommandBreakdownWithState(t *testing.T) {
	cmd := NewCostCommand()
	state := &RuntimeState{
		CostInputTokens:  123,
		CostOutputTokens: 45,
		CostCacheRead:    67,
		CostCacheWrite:   89,
	}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "cost"})
	if err != nil {
		t.Fatalf("cost execute failed: %v", err)
	}
	if res.Message != "COST_BREAKDOWN\ninput_tokens=123\noutput_tokens=45\ncache_read_tokens=67\ncache_write_tokens=89" {
		t.Fatalf("unexpected cost message: %q", res.Message)
	}
}

func TestDoctorCommandHumanAndJSON(t *testing.T) {
	cmd := NewDoctorCommand()
	state := &RuntimeState{
		ProviderName:  "openai",
		ProviderReady: true,
		Model:         "openai/gpt-4o",
		ConfigPath:    "/tmp/.alliecode/config.yaml",
		TransportMode: "remote",
		WorkspaceDirs: []string{"/tmp"},
	}

	humanRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "doctor"})
	if err != nil {
		t.Fatalf("human execute failed: %v", err)
	}
	if !strings.Contains(humanRes.Message, "DOCTOR_REPORT") {
		t.Fatalf("unexpected human output: %q", humanRes.Message)
	}
	if !strings.Contains(humanRes.Message, "section_count=5") {
		t.Fatalf("expected sections in human output: %q", humanRes.Message)
	}
	if !strings.Contains(humanRes.Message, "status=warn") {
		t.Fatalf("expected status in human output: %q", humanRes.Message)
	}
	if !strings.Contains(humanRes.Message, "section.2.name=provider") || !strings.Contains(humanRes.Message, "section.2.status=warn") {
		t.Fatalf("expected provider check in human output: %q", humanRes.Message)
	}

	jsonRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "doctor", Args: []string{"json"}})
	if err != nil {
		t.Fatalf("json execute failed: %v", err)
	}
	var payload struct {
		Status   string `json:"status"`
		Sections []struct {
			Name string `json:"name"`
		} `json:"sections"`
		Checks []struct {
			ID string `json:"id"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(jsonRes.Message), &payload); err != nil {
		t.Fatalf("expected valid json payload: %v", err)
	}
	if payload.Status != "warn" {
		t.Fatalf("unexpected json status: %q", payload.Status)
	}
	if len(payload.Sections) != 5 {
		t.Fatalf("expected section payload in json output: %q", jsonRes.Message)
	}
	foundTransport := false
	for _, check := range payload.Checks {
		if check.ID == "transport" {
			foundTransport = true
			break
		}
	}
	if !foundTransport {
		t.Fatalf("expected transport check in json output: %q", jsonRes.Message)
	}
}

func TestAddDirCommand(t *testing.T) {
	cmd := NewAddDirCommand()
	dir := t.TempDir()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "add-dir", Args: []string{dir}})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	want := "Workspace directory added: " + filepath.Clean(dir)
	if res.Message != want {
		t.Fatalf("unexpected message: %q", res.Message)
	}
	if len(state.WorkspaceDirs) != 1 || state.WorkspaceDirs[0] != filepath.Clean(dir) {
		t.Fatalf("workspace dirs not updated deterministically: %#v", state.WorkspaceDirs)
	}

	_, err = cmd.Execute(context.Background(), Context{}, Invocation{Name: "add-dir"})
	if err == nil || err.Error() != "usage: /add-dir <path>" {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestAgentsCommandListCreateStatus(t *testing.T) {
	cmd := NewAgentsCommand()

	listRes, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "agents"})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "AGENTS_LIST\ncount=1\nagent.1.name=local\nagent.1.status=available" {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}

	createRes, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "agents", Args: []string{"create", "worker-1"}})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if createRes.Message != "AGENTS_CREATE\nname=worker-1\nstatus=created" {
		t.Fatalf("unexpected create message: %q", createRes.Message)
	}

	statusNoneRes, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "agents", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusNoneRes.Message != "AGENTS_STATUS\nstatus=none" {
		t.Fatalf("unexpected status message: %q", statusNoneRes.Message)
	}

	state := &RuntimeState{Agent: &agent.Agent{}}
	statusCfgRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "agents", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status with state failed: %v", err)
	}
	if statusCfgRes.Message != "AGENTS_STATUS\nstatus=configured" {
		t.Fatalf("unexpected configured status: %q", statusCfgRes.Message)
	}
}

func TestClearCommand(t *testing.T) {
	cmd := NewClearCommand()
	state := &RuntimeState{CompactRequested: true, ResumeRequested: true, LastResumeTarget: "abc", DiffEntries: []DiffEntry{{Path: "x.go", Added: 1}}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "clear", Args: []string{"all"}})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Message != "CLEAR_RESULT\nscope=all\nclear_display=true\nclear_context=true\nclear_diff=true\ncount=1\ncontext_clears=1\ndiff_clears=1" {
		t.Fatalf("unexpected message: %q", res.Message)
	}
	if state.CompactRequested || state.ResumeRequested || state.LastResumeTarget != "" || len(state.DiffEntries) != 0 {
		t.Fatalf("expected clear all to reset stateful flags and diff entries: %+v", state)
	}
}

func TestResumeCommandStatefulStatusAndTarget(t *testing.T) {
	cmd := NewResumeCommand()
	state := &RuntimeState{}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "resume", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "RESUME_STATUS\nrequested=false\ncount=0\nlast_target=-" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "resume", Args: []string{"sess-123"}})
	if err != nil {
		t.Fatalf("resume target failed: %v", err)
	}
	if res.Message != "RESUME_REQUEST\ntarget=sess-123\nrequested=true\ncount=1\nlast_target=sess-123" {
		t.Fatalf("unexpected resume message: %q", res.Message)
	}
}

func TestClearCommandStatusTracksCounters(t *testing.T) {
	cmd := NewClearCommand()
	state := &RuntimeState{CompactRequested: true, ResumeRequested: true, LastResumeTarget: "sess-9"}

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "clear", Args: []string{"context"}})
	if err != nil {
		t.Fatalf("context clear failed: %v", err)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "clear", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "CLEAR_STATUS\ncount=0\ncontext_clears=1\ndiff_clears=0\nlast_scope=context\ncompact_requested=false\nresume_requested=false\nlast_resume_target=-" {
		t.Fatalf("unexpected clear status message: %q", statusRes.Message)
	}
}

func TestDiffCommandStatefulFlow(t *testing.T) {
	cmd := NewDiffCommand()
	state := &RuntimeState{}

	modeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "diff", Args: []string{"mode", "all"}})
	if err != nil {
		t.Fatalf("mode failed: %v", err)
	}
	if modeRes.Message != "DIFF_MODE\nmode=all" {
		t.Fatalf("unexpected mode message: %q", modeRes.Message)
	}

	addRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "diff", Args: []string{"add", "internal/commands/handlers.go", "10", "2", "1"}})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if addRes.Message != "DIFF_ADD\npath=internal/commands/handlers.go\nadded=10\nremoved=2\nmodified=1\nentries=1\nupdates=1" {
		t.Fatalf("unexpected add message: %q", addRes.Message)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "diff", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "DIFF_LIST\nmode=all\ncount=1\nentry.1.path=internal/commands/handlers.go\nentry.1.added=10\nentry.1.removed=2\nentry.1.modified=1" {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}

	clearRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "diff", Args: []string{"clear"}})
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if clearRes.Message != "DIFF_CLEAR\nentries=0\nupdates=1\nclears=1" {
		t.Fatalf("unexpected clear message: %q", clearRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "diff", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "DIFF_STATUS\nmode=all\nentries=0\nupdates=1\nclears=1" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}
}

func TestBranchCommandSwitchReportsDeterministicSideEffects(t *testing.T) {
	cmd := NewBranchCommand()
	state := &RuntimeState{}

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "branch", Args: []string{"create", "feature/s1"}})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	switchRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "branch", Args: []string{"switch", "main"}})
	if err != nil {
		t.Fatalf("switch failed: %v", err)
	}
	if switchRes.Message != "BRANCH_SWITCH\nactive=main\nprevious=feature/s1\nchanged=true\nswitches=2" {
		t.Fatalf("unexpected switch message: %q", switchRes.Message)
	}
}

func TestHistoryCommandListShow(t *testing.T) {
	cmd := NewHistoryCommand()
	state := &RuntimeState{HistoryEntries: []HistoryEntry{
		{ID: "b2", Path: "/tmp/.alliecode/history/b2.jsonl", CreatedAt: "2026-01-02T00:00:00Z", Model: "gpt-4o", Turns: 4, Title: "beta", Summary: "second"},
		{ID: "a1", Path: "/tmp/.alliecode/history/a1.jsonl", CreatedAt: "2026-01-01T00:00:00Z", Model: "gpt-4o-mini", Turns: 2, Title: "alpha", Summary: "first"},
	}}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.HasPrefix(listRes.Message, "HISTORY_LIST\nfilter.type=none\nfilter.value=-\nlimit=0\ncount=2") {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}
	if !strings.Contains(listRes.Message, "entry.1.id=a1") || !strings.Contains(listRes.Message, "entry.1.path=/tmp/.alliecode/history/a1.jsonl") || !strings.Contains(listRes.Message, "entry.2.id=b2") || !strings.Contains(listRes.Message, "entry.2.path=/tmp/.alliecode/history/b2.jsonl") {
		t.Fatalf("expected stable id lines in list message: %q", listRes.Message)
	}

	modelRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list", "model", "gpt-4o"}})
	if err != nil {
		t.Fatalf("model filter failed: %v", err)
	}
	if !strings.Contains(modelRes.Message, "filter.type=model") || !strings.Contains(modelRes.Message, "count=1") || !strings.Contains(modelRes.Message, "entry.1.id=b2") {
		t.Fatalf("unexpected model-filtered list message: %q", modelRes.Message)
	}

	textRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list", "text", "alpha"}})
	if err != nil {
		t.Fatalf("text filter failed: %v", err)
	}
	if !strings.Contains(textRes.Message, "filter.type=text") || !strings.Contains(textRes.Message, "count=1") || !strings.Contains(textRes.Message, "entry.1.id=a1") {
		t.Fatalf("unexpected text-filtered list message: %q", textRes.Message)
	}

	showRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"show", "a1"}})
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.HasPrefix(showRes.Message, "HISTORY_SHOW\nid=a1\nsection_count=2") || !strings.Contains(showRes.Message, "section.1.name=meta") || !strings.Contains(showRes.Message, "section.1.path=/tmp/.alliecode/history/a1.jsonl") || !strings.Contains(showRes.Message, "section.2.name=content") {
		t.Fatalf("unexpected show message: %q", showRes.Message)
	}

	_, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"show", "missing"}})
	if err == nil || err.Error() != "history entry not found: missing" {
		t.Fatalf("expected missing history error, got %v", err)
	}
}

func TestHistoryCommandInvalidFilterUsage(t *testing.T) {
	cmd := NewHistoryCommand()
	_, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "history", Args: []string{"list", "model"}})
	if err == nil || err.Error() != "usage: /history [list [model <model>|text <query>|limit <n>|latest]|latest|show <id>]" {
		t.Fatalf("expected filter usage error, got %v", err)
	}
}

func TestMCPCommandLifecycleLegacyFallback(t *testing.T) {
	cmd := NewMCPCommand()
	state := &RuntimeState{}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "MCP_LIST\ncount=0" {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}

	connectRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"connect", "local"}})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if connectRes.Message != "MCP_CONNECT\nname=local\nstatus=connected\nconnected=true" {
		t.Fatalf("unexpected connect message: %q", connectRes.Message)
	}

	listConnectedRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list after connect failed: %v", err)
	}
	if listConnectedRes.Message != "MCP_LIST\ncount=1\nserver.1.name=local\nserver.1.status=connected\nserver.1.connected=true" {
		t.Fatalf("unexpected list after connect message: %q", listConnectedRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"status", "local"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "MCP_STATUS\nname=local\nstatus=connected\nconnected=true" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	disconnectRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"disconnect", "local"}})
	if err != nil {
		t.Fatalf("disconnect failed: %v", err)
	}
	if disconnectRes.Message != "MCP_DISCONNECT\nname=local\nstatus=disconnected\nconnected=false" {
		t.Fatalf("unexpected disconnect message: %q", disconnectRes.Message)
	}

	statusDisconnectedRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"status", "local"}})
	if err != nil {
		t.Fatalf("status after disconnect failed: %v", err)
	}
	if statusDisconnectedRes.Message != "MCP_STATUS\nname=local\nstatus=disconnected\nconnected=false" {
		t.Fatalf("unexpected disconnected status message: %q", statusDisconnectedRes.Message)
	}
}

func TestMCPCommandManagerBackedSubcommands(t *testing.T) {
	cmd := NewMCPCommand()
	manager, err := mcp.NewManager(context.Background(), []mcp.ServerConfig{{Name: "alpha", Disabled: true}})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	state := &RuntimeState{}
	cmdCtx := Context{State: state, MCPManager: manager}

	listRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "MCP_LIST\ncount=1\nserver.1.name=alpha\nserver.1.transport=stdio\nserver.1.connection_state=disabled\nserver.1.auth_status=unauthenticated\nserver.1.authenticated=false" {
		t.Fatalf("unexpected manager list message: %q", listRes.Message)
	}

	addRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"add", "beta", "stdio", "npx", "-y", "beta-mcp"}})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if addRes.Message != "MCP_ADD\nname=beta\ntransport=stdio\nadded=true" {
		t.Fatalf("unexpected add message: %q", addRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"status", "beta"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "MCP_STATUS\nname=beta\nconnection_state=pending\ntransport=stdio\nauth_status=unauthenticated\nauthenticated=false\ncached_resources=0\ncached_contents=0" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	authStatusRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"auth-status", "beta"}})
	if err != nil {
		t.Fatalf("auth-status failed: %v", err)
	}
	if authStatusRes.Message != "MCP_AUTH_STATUS\nname=beta\nauth_status=unauthenticated\nauthenticated=false\nconnection_state=pending" {
		t.Fatalf("unexpected auth-status message: %q", authStatusRes.Message)
	}

	listToolsRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"list-tools"}})
	if err != nil {
		t.Fatalf("list-tools failed: %v", err)
	}
	if listToolsRes.Message != "MCP_TOOLS\ncount=0" {
		t.Fatalf("unexpected list-tools message: %q", listToolsRes.Message)
	}

	listResourcesRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"list-resources"}})
	if err != nil {
		t.Fatalf("list-resources failed: %v", err)
	}
	if listResourcesRes.Message != "MCP_RESOURCES\ncount=0" {
		t.Fatalf("unexpected list-resources message: %q", listResourcesRes.Message)
	}

	removeRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"remove", "beta"}})
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if removeRes.Message != "MCP_REMOVE\nname=beta\nremoved=true" {
		t.Fatalf("unexpected remove message: %q", removeRes.Message)
	}

	removeMissingRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"remove", "missing"}})
	if err != nil {
		t.Fatalf("remove missing failed: %v", err)
	}
	if removeMissingRes.Message != "MCP_REMOVE\nname=missing\nremoved=false" {
		t.Fatalf("unexpected remove missing message: %q", removeMissingRes.Message)
	}
}

func TestVimCommandEnableDisableStatus(t *testing.T) {
	cmd := NewVimCommand()
	state := &RuntimeState{}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "vim", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "VIM_STATUS\nenabled=false" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	enableRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "vim", Args: []string{"enable"}})
	if err != nil {
		t.Fatalf("enable failed: %v", err)
	}
	if enableRes.Message != "VIM_SET\nenabled=true" {
		t.Fatalf("unexpected enable message: %q", enableRes.Message)
	}

	disableRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "vim", Args: []string{"disable"}})
	if err != nil {
		t.Fatalf("disable failed: %v", err)
	}
	if disableRes.Message != "VIM_SET\nenabled=false" {
		t.Fatalf("unexpected disable message: %q", disableRes.Message)
	}
}

func TestVoiceCommandEnableDisableStatus(t *testing.T) {
	cmd := NewVoiceCommand()
	state := &RuntimeState{}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "voice", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "VOICE_STATUS\nmode=local-placeholder\nenabled=false" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	enableRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "voice", Args: []string{"enable"}})
	if err != nil {
		t.Fatalf("enable failed: %v", err)
	}
	if enableRes.Message != "VOICE_SET\nmode=local-placeholder\nenabled=true" {
		t.Fatalf("unexpected enable message: %q", enableRes.Message)
	}

	disableRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "voice", Args: []string{"disable"}})
	if err != nil {
		t.Fatalf("disable failed: %v", err)
	}
	if disableRes.Message != "VOICE_SET\nmode=local-placeholder\nenabled=false" {
		t.Fatalf("unexpected disable message: %q", disableRes.Message)
	}
}

func TestBuddyCommandStatus(t *testing.T) {
	cmd := NewBuddyCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if res.Message != "BUDDY_STATUS\nstate=egg\nhatched=false\nmuted=false\npet_count=0" {
		t.Fatalf("unexpected status message: %q", res.Message)
	}
}

func TestBuddyCommandHatch(t *testing.T) {
	cmd := NewBuddyCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"hatch"}})
	if err != nil {
		t.Fatalf("hatch failed: %v", err)
	}
	if res.Message != "BUDDY_HATCH\nstatus=hatched\nhatched=true" {
		t.Fatalf("unexpected hatch message: %q", res.Message)
	}
	if !state.BuddyHatched || state.BuddyPetCount != 0 {
		t.Fatalf("expected hatched state with zero pets: %+v", state)
	}

	res, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"hatch"}})
	if err != nil {
		t.Fatalf("second hatch failed: %v", err)
	}
	if res.Message != "BUDDY_HATCH\nstatus=already_hatched\nhatched=true" {
		t.Fatalf("unexpected second hatch message: %q", res.Message)
	}
}

func TestBuddyCommandPet(t *testing.T) {
	cmd := NewBuddyCommand()
	state := &RuntimeState{}

	notHatchedRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"pet"}})
	if err != nil {
		t.Fatalf("pet before hatch failed: %v", err)
	}
	if notHatchedRes.Message != "BUDDY_PET\nstatus=not_hatched\nhatched=false\npet_count=0" {
		t.Fatalf("unexpected pet before hatch message: %q", notHatchedRes.Message)
	}

	_, _ = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"hatch"}})
	petRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"pet"}})
	if err != nil {
		t.Fatalf("pet after hatch failed: %v", err)
	}
	if petRes.Message != "BUDDY_PET\nstatus=pet\nhatched=true\nmuted=false\npet_count=1" {
		t.Fatalf("unexpected pet message: %q", petRes.Message)
	}

	_, _ = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"mute"}})
	petMutedRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"pet"}})
	if err != nil {
		t.Fatalf("pet while muted failed: %v", err)
	}
	if petMutedRes.Message != "BUDDY_PET\nstatus=pet_muted\nhatched=true\nmuted=true\npet_count=2" {
		t.Fatalf("unexpected pet muted message: %q", petMutedRes.Message)
	}
}

func TestBuddyCommandMuteUnmute(t *testing.T) {
	cmd := NewBuddyCommand()
	state := &RuntimeState{}

	muteRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"mute"}})
	if err != nil {
		t.Fatalf("mute failed: %v", err)
	}
	if muteRes.Message != "BUDDY_MUTE\nmuted=true" || !state.BuddyMuted {
		t.Fatalf("unexpected mute result: %q state=%+v", muteRes.Message, state)
	}

	unmuteRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"unmute"}})
	if err != nil {
		t.Fatalf("unmute failed: %v", err)
	}
	if unmuteRes.Message != "BUDDY_UNMUTE\nmuted=false" || state.BuddyMuted {
		t.Fatalf("unexpected unmute result: %q state=%+v", unmuteRes.Message, state)
	}
}

func TestBuddyCommandHelp(t *testing.T) {
	cmd := NewBuddyCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"help"}})
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if res.Message != "BUDDY_HELP\nusage=/buddy [status|hatch|pet|mute|unmute|help]\nsubcommands=status,hatch,pet,mute,unmute,help" {
		t.Fatalf("unexpected help message: %q", res.Message)
	}
}

func TestBuddyCommandInvalidUsage(t *testing.T) {
	cmd := NewBuddyCommand()
	state := &RuntimeState{}

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"unknown"}})
	if err == nil || err.Error() != "usage: /buddy [status|hatch|pet|mute|unmute|help]" {
		t.Fatalf("expected usage error for unknown subcommand, got %v", err)
	}

	_, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "buddy", Args: []string{"status", "extra"}})
	if err == nil || err.Error() != "usage: /buddy [status|hatch|pet|mute|unmute|help]" {
		t.Fatalf("expected usage error for too many args, got %v", err)
	}
}

func TestThemeCommand(t *testing.T) {
	cmd := NewThemeCommand()
	state := &RuntimeState{}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "theme"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "THEME_STATUS\ntheme=system" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "theme", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "THEME_LIST\ncount=3\ntheme.1=dark\ntheme.2=light\ntheme.3=system" {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "theme", Args: []string{"set", "dark"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if setRes.Message != "THEME_SET\ntheme=dark\nset_count=1" {
		t.Fatalf("unexpected set message: %q", setRes.Message)
	}
	if state.ThemeSetCount != 1 {
		t.Fatalf("expected theme set count increment, got %d", state.ThemeSetCount)
	}
}

func TestFilesCommandFlow(t *testing.T) {
	cmd := NewFilesCommand()
	state := &RuntimeState{}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "files"})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "FILES_LIST\ncount=0" {
		t.Fatalf("unexpected empty list message: %q", listRes.Message)
	}

	addRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "files", Args: []string{"add", "internal/commands/handlers.go"}})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if addRes.Message != "FILES_ADD\npath=internal/commands/handlers.go\nadded=true\ncount=1" {
		t.Fatalf("unexpected add message: %q", addRes.Message)
	}

	dupRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "files", Args: []string{"add", "internal/commands/handlers.go"}})
	if err != nil {
		t.Fatalf("duplicate add failed: %v", err)
	}
	if dupRes.Message != "FILES_ADD\npath=internal/commands/handlers.go\nadded=false\ncount=1" {
		t.Fatalf("unexpected duplicate add message: %q", dupRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "files", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "FILES_STATUS\ncount=1\nproject_paths=0\nadds=1\nremoves=0\nclears=0\nlast_action=add\nlast_path=internal/commands/handlers.go" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	removeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "files", Args: []string{"remove", "internal/commands/handlers.go"}})
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if removeRes.Message != "FILES_REMOVE\npath=internal/commands/handlers.go\nremoved=true\ncount=0" {
		t.Fatalf("unexpected remove message: %q", removeRes.Message)
	}
}

func TestFilesCommandUsesPersistedProjectPaths(t *testing.T) {
	cmd := NewFilesCommand()
	state := &RuntimeState{ProjectPaths: []string{"/repo/a.go", "/repo/b.go"}}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "files", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "FILES_LIST\ncount=2\nfile.1=/repo/a.go\nfile.2=/repo/b.go" {
		t.Fatalf("unexpected persisted list message: %q", listRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "files", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "FILES_STATUS\ncount=2\nproject_paths=2\nadds=0\nremoves=0\nclears=0\nlast_action=-\nlast_path=-" {
		t.Fatalf("unexpected persisted status message: %q", statusRes.Message)
	}
}

func TestStatuslineCommandFlow(t *testing.T) {
	cmd := NewStatuslineCommand()
	state := &RuntimeState{}

	setupRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "statusline", Args: []string{"setup", "Use", "my", "PS1"}})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if setupRes.Message != "STATUSLINE_SETUP\nsubagent_type=statusline-setup\nprompt=Use my PS1\ncount=1" {
		t.Fatalf("unexpected setup message: %q", setupRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "statusline", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	for _, want := range []string{"STATUSLINE_STATUS", "count=1", "provider=-", "model=-", "model_ref=-", "logged_in=false", "provider_ready=false", "working_dir=.", "workspace_root=.", "path_scope=workspace", "turns=0", "last_stop_reason=-"} {
		if !strings.Contains(statusRes.Message, want) {
			t.Fatalf("statusline status missing %q: %q", want, statusRes.Message)
		}
	}
}

func TestKeybindingsCommandFlow(t *testing.T) {
	cmd := NewKeybindingsCommand()
	state := &RuntimeState{}

	disabledRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "keybindings"})
	if err != nil {
		t.Fatalf("disabled open failed: %v", err)
	}
	if disabledRes.Message != "Keybinding customization is not enabled. This feature is currently in preview." {
		t.Fatalf("unexpected disabled message: %q", disabledRes.Message)
	}

	_, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "keybindings", Args: []string{"enable"}})
	if err != nil {
		t.Fatalf("enable failed: %v", err)
	}
	state.KeybindingsPath = "~/.claude/keybindings.json"

	createRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "keybindings", Args: []string{"open"}})
	if err != nil {
		t.Fatalf("open(create) failed: %v", err)
	}
	if createRes.Message != "Created ~/.claude/keybindings.json with template. Opened in your editor." {
		t.Fatalf("unexpected create message: %q", createRes.Message)
	}

	openRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "keybindings"})
	if err != nil {
		t.Fatalf("open(existing) failed: %v", err)
	}
	if openRes.Message != "Opened ~/.claude/keybindings.json in your editor." {
		t.Fatalf("unexpected open message: %q", openRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "keybindings", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "KEYBINDINGS_STATUS\nenabled=true\npath=~/.claude/keybindings.json\nexists=true\nopens=2" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}
}

func TestStatusAndStatsCommands(t *testing.T) {
	statusCmd := NewStatusCommand()
	statsCmd := NewStatsCommand()
	state := &RuntimeState{Model: "gpt-4o", ProviderName: "openai", PermissionMode: permissions.ModeAuto, CompactMode: "off", CompactRequested: true, ResumeRequested: true, CompactCount: 3, ResumeCount: 2, CopyCount: 4, ClearCount: 5, BranchCount: 1, DiffCount: 7, MemoryWrites: 8, TasksCompleted: 2}

	statusRes, err := statusCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	for _, want := range []string{"STATUS_REPORT", "model=gpt-4o", "model_capabilities=text,image,audio,tool_use,vision,attachments", "provider=openai", "permission_mode=auto", "compact_mode=off", "working_dir=.", "workspace_root=.", "path_scope=workspace", "tasks_completed=2", "next=/login provider openai"} {
		if !strings.Contains(statusRes.Message, want) {
			t.Fatalf("status output missing %q: %q", want, statusRes.Message)
		}
	}

	statsRes, err := statsCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "stats"})
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if statsRes.Message != "STATS_REPORT\ncompact_count=3\nresume_count=2\ncopy_count=4\nclear_count=5\nbranch_creates=1\ndiff_updates=7\nmemory_writes=8\ntasks_completed=2" {
		t.Fatalf("unexpected stats message: %q", statsRes.Message)
	}
}

func TestStatusAndStatuslineIncludeAgentRuntimeSignals(t *testing.T) {
	provider := &statusRuntimeProvider{}
	ag := agent.New(agent.Config{Provider: provider, MaxTurns: 5})
	if err := ag.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("agent run failed: %v", err)
	}

	state := &RuntimeState{Agent: ag, Model: "gpt-4o", ProviderName: "openai", PermissionMode: permissions.ModeAuto}

	statusRes, err := NewStatusCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	for _, want := range []string{
		"model=gpt-4o",
		"turns=3",
		"tool_inflight=0",
		"tasks_total=0",
		"teams_total=0",
		"last_stop_reason=max_tokens_recovery_exhausted",
	} {
		if !strings.Contains(statusRes.Message, want) {
			t.Fatalf("status output missing %q: %q", want, statusRes.Message)
		}
	}

	statuslineRes, err := NewStatuslineCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "statusline", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("statusline failed: %v", err)
	}
	for _, want := range []string{
		"turns=3",
		"tool_inflight=0",
		"tasks_total=0",
		"teams_total=0",
		"last_stop_reason=max_tokens_recovery_exhausted",
	} {
		if !strings.Contains(statuslineRes.Message, want) {
			t.Fatalf("statusline output missing %q: %q", want, statuslineRes.Message)
		}
	}
}

func TestModelCommandSetAndStatusExposeCapabilitySummary(t *testing.T) {
	modelCmd := NewModelCommand()
	statusCmd := NewStatusCommand()
	state := &RuntimeState{ProviderName: "openai"}

	setRes, err := modelCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"gpt-4o-mini"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if setRes.Message != "Model set to gpt-4o-mini\nCapabilities: text,image,audio,tool_use,vision,attachments\nQuick fix: /provider status\nNext: run /status to confirm runtime readiness." {
		t.Fatalf("unexpected set message: %q", setRes.Message)
	}

	statusRes, err := statusCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "model_capabilities=text,image,audio,tool_use,vision,attachments") {
		t.Fatalf("expected status to include model capability summary, got: %q", statusRes.Message)
	}
}

func TestProviderCommandStatusListAndSet(t *testing.T) {
	cmd := NewProviderCommand()
	state := &RuntimeState{}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "PROVIDER_STATUS\nprovider=-\nprovider_ready=false\nquick_fix_model=/model_<provider/model>\nquick_fix_auth=/provider set ollama\nnext=use_/provider_set_<name>_or_/login_provider_<name>" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "PROVIDER_LIST\ncount=4\nprovider.1=anthropic\nprovider.2=gemini\nprovider.3=ollama\nprovider.4=openai\nnext=use_/provider_set_<name>_to_switch" {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"set", "openai"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if setRes.Message != "PROVIDER_SET\nprovider=openai\nmodel=gpt-4o-mini\nprovider_ready=false\nquick_fix_model=/model_openai/<model>\nquick_fix_auth=/login provider openai\nnext=run_/status_to_confirm_runtime_for_openai" {
		t.Fatalf("unexpected set message: %q", setRes.Message)
	}

	_, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"set", "unknown"}})
	if err == nil || err.Error() != "unknown provider \"unknown\" (run /provider list, then /provider set <name>)" {
		t.Fatalf("expected unknown provider error, got %v", err)
	}
}

func TestMemoryCommand(t *testing.T) {
	cmd := NewMemoryCommand()
	state := &RuntimeState{}

	addRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "memory", Args: []string{"add", "remember", "this"}})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if addRes.Message != "MEMORY_ADD\nentry=remember this\ncount=1\nwrites=1" {
		t.Fatalf("unexpected add message: %q", addRes.Message)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "memory", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listRes.Message != "MEMORY_LIST\ncount=1\nentry.1=remember this" {
		t.Fatalf("unexpected list message: %q", listRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "memory", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "MEMORY_STATUS\ncount=1\nwrites=1\nremoves=0\nclears=0\nlast=remember this" {
		t.Fatalf("unexpected status message: %q", statusRes.Message)
	}

	clearRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "memory", Args: []string{"clear"}})
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if clearRes.Message != "MEMORY_CLEAR\ncount=0\nclears=1" {
		t.Fatalf("unexpected clear message: %q", clearRes.Message)
	}
}

func TestPrivacySettingsCommand(t *testing.T) {
	cmd := NewPrivacySettingsCommand()
	state := &RuntimeState{}

	showRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "privacy-settings"})
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if showRes.Message != "PRIVACY_SETTINGS\ntelemetry=false\ntraining=false\nupdates=0" {
		t.Fatalf("unexpected show message: %q", showRes.Message)
	}

	setTelemetryRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "privacy-settings", Args: []string{"set", "telemetry", "on"}})
	if err != nil {
		t.Fatalf("set telemetry failed: %v", err)
	}
	if setTelemetryRes.Message != "PRIVACY_SET\nfield=telemetry\nenabled=true\nupdated=true\nupdates=1" {
		t.Fatalf("unexpected set telemetry message: %q", setTelemetryRes.Message)
	}

	setTrainingRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "privacy-settings", Args: []string{"set", "training", "off"}})
	if err != nil {
		t.Fatalf("set training failed: %v", err)
	}
	if setTrainingRes.Message != "PRIVACY_SET\nfield=training\nenabled=false\nupdated=false\nupdates=1" {
		t.Fatalf("unexpected set training message: %q", setTrainingRes.Message)
	}
}

func TestLoginLogoutCommandFlow(t *testing.T) {
	login := NewLoginCommand()
	logout := NewLogoutCommand()
	state := &RuntimeState{}

	loginRes, err := login.Execute(context.Background(), Context{State: state}, Invocation{Name: "login", Args: []string{"provider", "openai"}})
	if err != nil {
		t.Fatalf("login provider failed: %v", err)
	}
	if loginRes.Message != "LOGIN_PROVIDER\nprovider=openai\nlogged_in=true\nprovider_ready=true\nlogin_count=1\nstatus=authenticated\nquick_fix_model=/model_openai/<model>\nquick_fix_verify=/provider_status\nnext=run_/model_list_openai_or_/model_openai/<model>" {
		t.Fatalf("unexpected login provider message: %q", loginRes.Message)
	}

	accountRes, err := login.Execute(context.Background(), Context{State: state}, Invocation{Name: "login", Args: []string{"account", "dev@acme"}})
	if err != nil {
		t.Fatalf("login account failed: %v", err)
	}
	if accountRes.Message != "LOGIN_ACCOUNT\naccount=dev@acme\nprovider=openai\nlogged_in=true\nprovider_ready=true\nlogin_count=2\nstatus=authenticated\nquick_fix_model=/model_openai/<model>\nquick_fix_verify=/provider_status\nnext=run_/model_list_openai_or_/model_openai/<model>" {
		t.Fatalf("unexpected login account message: %q", accountRes.Message)
	}

	statusRes, err := login.Execute(context.Background(), Context{State: state}, Invocation{Name: "login", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("login status failed: %v", err)
	}
	if statusRes.Message != "LOGIN_STATUS\nlogged_in=true\nprovider=openai\naccount=dev@acme\nprovider_ready=true\nlogin_count=2\nlogout_count=0" {
		t.Fatalf("unexpected login status message: %q", statusRes.Message)
	}

	logoutRes, err := logout.Execute(context.Background(), Context{State: state}, Invocation{Name: "logout"})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if logoutRes.Message != "LOGOUT_RESULT\nwas_logged_in=true\nlogged_in=false\nprovider_ready=false\nlogout_count=1\nquick_fix_auth=/login provider openai\nnext=run_/login_provider_<name>_to_reauthenticate" {
		t.Fatalf("unexpected logout message: %q", logoutRes.Message)
	}
}

func TestOutputStyleCommandFlow(t *testing.T) {
	cmd := NewOutputStyleCommand()
	state := &RuntimeState{}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "output-style", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if statusRes.Message != "OUTPUT_STYLE_STATUS\nstyle=default\nset_count=0" {
		t.Fatalf("unexpected output-style status: %q", statusRes.Message)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "output-style", Args: []string{"set", "concise"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if setRes.Message != "OUTPUT_STYLE_SET\nstyle=concise\nset_count=1" {
		t.Fatalf("unexpected output-style set: %q", setRes.Message)
	}
}

func TestUpgradeTerminalReleaseInstallFeedbackCommands(t *testing.T) {
	upgrade := NewUpgradeCommand()
	terminal := NewTerminalSetupCommand()
	releaseNotes := NewReleaseNotesCommand()
	github := NewInstallGitHubAppCommand()
	slack := NewInstallSlackAppCommand()
	feedback := NewFeedbackCommand()
	state := &RuntimeState{}

	upgradeRes, err := upgrade.Execute(context.Background(), Context{State: state}, Invocation{Name: "upgrade", Args: []string{"team"}})
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if upgradeRes.Message != "UPGRADE_REQUEST\nplan=team\nrequested=true\ncount=1\nnext=complete_upgrade_in_provider_portal" {
		t.Fatalf("unexpected upgrade message: %q", upgradeRes.Message)
	}

	t.Setenv("TERM_PROGRAM", "ghostty")
	terminalRes, err := terminal.Execute(context.Background(), Context{State: state}, Invocation{Name: "terminal-setup", Args: []string{"detect"}})
	if err != nil {
		t.Fatalf("terminal detect failed: %v", err)
	}
	if !strings.Contains(terminalRes.Message, "TERMINAL_SETUP_DETECT\nprofile=ghostty") {
		t.Fatalf("unexpected terminal detect message: %q", terminalRes.Message)
	}

	applyRes, err := terminal.Execute(context.Background(), Context{State: state}, Invocation{Name: "terminal-setup", Args: []string{"apply"}})
	if err != nil {
		t.Fatalf("terminal setup failed: %v", err)
	}
	if applyRes.Message != "TERMINAL_SETUP_APPLY\nprofile=ghostty\nconfigured=true\ncount=1" {
		t.Fatalf("unexpected terminal setup message: %q", applyRes.Message)
	}

	statusRes, err := terminal.Execute(context.Background(), Context{State: state}, Invocation{Name: "terminal-setup", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("terminal setup status failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "TERMINAL_SETUP_STATUS\nconfigured=true\ncount=1\nprofile=ghostty\ndetected_profile=ghostty\ndetect_source=term_program") {
		t.Fatalf("unexpected terminal setup status message: %q", statusRes.Message)
	}

	releaseRes, err := releaseNotes.Execute(context.Background(), Context{State: state}, Invocation{Name: "release-notes", Args: []string{"latest"}})
	if err != nil {
		t.Fatalf("release notes failed: %v", err)
	}
	if !strings.HasPrefix(releaseRes.Message, "RELEASE_NOTES_LATEST\nversion=v1.0.0") {
		t.Fatalf("unexpected release notes message: %q", releaseRes.Message)
	}

	githubRes, err := github.Execute(context.Background(), Context{State: state}, Invocation{Name: "install-github-app", Args: []string{"start", "acme/repo"}})
	if err != nil {
		t.Fatalf("github install failed: %v", err)
	}
	if githubRes.Message != "INSTALL_GITHUB_APP_START\nrepo=acme/repo\ncount=1\nnext=run_gh_auth_and_repo_setup" {
		t.Fatalf("unexpected github install message: %q", githubRes.Message)
	}

	slackRes, err := slack.Execute(context.Background(), Context{State: state}, Invocation{Name: "install-slack-app"})
	if err != nil {
		t.Fatalf("slack install failed: %v", err)
	}
	if slackRes.Message != "INSTALL_SLACK_APP_START\ncount=1\nnext=open_slack_marketplace_link" {
		t.Fatalf("unexpected slack install message: %q", slackRes.Message)
	}

	feedbackRes, err := feedback.Execute(context.Background(), Context{State: state}, Invocation{Name: "feedback", Args: []string{"submit", "parity", "looks", "good"}})
	if err != nil {
		t.Fatalf("feedback submit failed: %v", err)
	}
	if feedbackRes.Message != "FEEDBACK_SUBMIT\nmessage=parity looks good\ncount=1" {
		t.Fatalf("unexpected feedback message: %q", feedbackRes.Message)
	}
}

func TestIDECommandFlow(t *testing.T) {
	cmd := NewIDECommand()
	state := &RuntimeState{}

	t.Setenv("VISUAL", "code")
	detectRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "ide", Args: []string{"detect"}})
	if err != nil {
		t.Fatalf("ide detect failed: %v", err)
	}
	if !strings.Contains(detectRes.Message, "IDE_DETECT\neditor=vscode\nsource=visual") {
		t.Fatalf("unexpected ide detect message: %q", detectRes.Message)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "ide", Args: []string{"set-editor", "cursor"}})
	if err != nil {
		t.Fatalf("ide set-editor failed: %v", err)
	}
	if setRes.Message != "IDE_SET_EDITOR\neditor=cursor\nconfig_count=1" {
		t.Fatalf("unexpected ide set-editor message: %q", setRes.Message)
	}

	openRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "ide", Args: []string{"open", "./repo"}})
	if err != nil {
		t.Fatalf("ide open failed: %v", err)
	}
	if openRes.Message != "IDE_OPEN_HINT\neditor=cursor\ntarget=./repo\ncommand=cursor ./repo\nopen_count=1" {
		t.Fatalf("unexpected ide open message: %q", openRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "ide", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("ide status failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "IDE_STATUS\neditor=cursor\ndetected=vscode\nsource=visual\nopen_count=1\nconfig_count=1\nhint_count=2") {
		t.Fatalf("unexpected ide status message: %q", statusRes.Message)
	}
}

func TestHooksSandboxTasksCommands(t *testing.T) {
	hooksCmd := NewHooksCommand()
	sandboxCmd := NewSandboxCommand()
	tasksCmd := NewTasksCommand()
	state := &RuntimeState{}

	hooksSetRes, err := hooksCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "hooks", Args: []string{"enable", "all"}})
	if err != nil {
		t.Fatalf("hooks enable failed: %v", err)
	}
	if hooksSetRes.Message != "HOOKS_SET\naction=enable\ntarget=all\npre=true\npost=true" {
		t.Fatalf("unexpected hooks set message: %q", hooksSetRes.Message)
	}

	hooksStatusRes, err := hooksCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "hooks", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("hooks status failed: %v", err)
	}
	if hooksStatusRes.Message != "HOOKS_STATUS\npre=true\npost=true" {
		t.Fatalf("unexpected hooks status message: %q", hooksStatusRes.Message)
	}

	sandboxStatusRes, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox"})
	if err != nil {
		t.Fatalf("sandbox status failed: %v", err)
	}
	if !strings.Contains(sandboxStatusRes.Message, "SANDBOX_STATUS\nmode=workspace-write\nworkspace_locked=false\nexcluded_count=0\navailability.shell=") || !strings.Contains(sandboxStatusRes.Message, "scope.path_scope=workspace") || !strings.Contains(sandboxStatusRes.Message, "scope.restrictions=workspace_write") {
		t.Fatalf("unexpected sandbox status message: %q", sandboxStatusRes.Message)
	}

	sandboxSetRes, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"set", "read-only"}})
	if err != nil {
		t.Fatalf("sandbox set failed: %v", err)
	}
	if sandboxSetRes.Message != "SANDBOX_SET\nmode=read-only" {
		t.Fatalf("unexpected sandbox set message: %q", sandboxSetRes.Message)
	}

	lockRes, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"lock", "on"}})
	if err != nil {
		t.Fatalf("sandbox lock failed: %v", err)
	}
	if lockRes.Message != "SANDBOX_LOCK_SET\nworkspace_locked=true" {
		t.Fatalf("unexpected sandbox lock message: %q", lockRes.Message)
	}

	addExcludeRes, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"exclude", "add", "npm run test:*"}})
	if err != nil {
		t.Fatalf("sandbox exclude add failed: %v", err)
	}
	if addExcludeRes.Message != "SANDBOX_EXCLUDE_ADD\npattern=npm run test:*\nadded=true\ncount=1" {
		t.Fatalf("unexpected sandbox exclude add message: %q", addExcludeRes.Message)
	}

	listExcludeRes, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"exclude", "list"}})
	if err != nil {
		t.Fatalf("sandbox exclude list failed: %v", err)
	}
	if listExcludeRes.Message != "SANDBOX_EXCLUDE_LIST\ncount=1\npattern.1=npm run test:*" {
		t.Fatalf("unexpected sandbox exclude list message: %q", listExcludeRes.Message)
	}

	addTaskRes, err := tasksCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "tasks", Args: []string{"add", "ship", "parity"}})
	if err != nil {
		t.Fatalf("task add failed: %v", err)
	}
	if addTaskRes.Message != "TASKS_ADD\ntask=ship parity\ncount=1" {
		t.Fatalf("unexpected tasks add message: %q", addTaskRes.Message)
	}

	listTaskRes, err := tasksCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "tasks", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("task list failed: %v", err)
	}
	if listTaskRes.Message != "TASKS_LIST\ncount=1\ncompleted=0\ntask.1=ship parity" {
		t.Fatalf("unexpected tasks list message: %q", listTaskRes.Message)
	}

	doneTaskRes, err := tasksCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "tasks", Args: []string{"done", "1"}})
	if err != nil {
		t.Fatalf("task done failed: %v", err)
	}
	if doneTaskRes.Message != "TASKS_DONE\nindex=1\ntask=ship parity\ncount=0\ncompleted=1" {
		t.Fatalf("unexpected tasks done message: %q", doneTaskRes.Message)
	}

	clearTaskRes, err := tasksCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "tasks", Args: []string{"clear"}})
	if err != nil {
		t.Fatalf("task clear failed: %v", err)
	}
	if clearTaskRes.Message != "TASKS_CLEAR\ncount=0\ncompleted=1" {
		t.Fatalf("unexpected tasks clear message: %q", clearTaskRes.Message)
	}
}

func TestSandboxCommandPersistsSettingsWhenPathConfigured(t *testing.T) {
	sandboxCmd := NewSandboxCommand()
	state := &RuntimeState{SandboxSettingsPath: filepath.Join(t.TempDir(), "sandbox-settings.json")}

	_, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"mode", "danger-full-access"}})
	if err != nil {
		t.Fatalf("sandbox mode failed: %v", err)
	}
	_, err = sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"lock", "on"}})
	if err != nil {
		t.Fatalf("sandbox lock failed: %v", err)
	}
	_, err = sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"exclude", "add", "git push"}})
	if err != nil {
		t.Fatalf("sandbox exclude failed: %v", err)
	}

	fresh := &RuntimeState{SandboxSettingsPath: state.SandboxSettingsPath}
	statusRes, err := sandboxCmd.Execute(context.Background(), Context{State: fresh}, Invocation{Name: "sandbox", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("sandbox status failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "SANDBOX_STATUS\nmode=danger-full-access\nworkspace_locked=true\nexcluded_count=1\navailability.shell=") || !strings.Contains(statusRes.Message, "scope.restrictions=workspace_locked") {
		t.Fatalf("unexpected persisted sandbox status: %q", statusRes.Message)
	}

	listRes, err := sandboxCmd.Execute(context.Background(), Context{State: fresh}, Invocation{Name: "sandbox", Args: []string{"exclude", "list"}})
	if err != nil {
		t.Fatalf("sandbox exclude list failed: %v", err)
	}
	if listRes.Message != "SANDBOX_EXCLUDE_LIST\ncount=1\npattern.1=git push" {
		t.Fatalf("unexpected persisted sandbox exclude list: %q", listRes.Message)
	}
}

func TestSandboxCommandCheckReportsDiagnostics(t *testing.T) {
	sandboxCmd := NewSandboxCommand()
	state := &RuntimeState{}

	origLookPath := sandboxLookPath
	t.Cleanup(func() { sandboxLookPath = origLookPath })
	sandboxLookPath = func(file string) (string, error) {
		switch file {
		case "sh":
			return "/bin/sh", nil
		case "bash":
			return "/bin/bash", nil
		case "git":
			return "", exec.ErrNotFound
		default:
			return "", exec.ErrNotFound
		}
	}

	res, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"check"}})
	if err != nil {
		t.Fatalf("sandbox check failed: %v", err)
	}
	want := "SANDBOX_CHECK\nstatus=warn\ncheck_count=5\ncheck.1.id=shell_available\ncheck.1.status=ok\ncheck.1.detail=sh available at /bin/sh\ncheck.2.id=dependency_git\ncheck.2.status=warn\ncheck.2.detail=git binary not found\ncheck.3.id=policy_mode\ncheck.3.status=ok\ncheck.3.detail=mode=workspace-write alias=balanced\ncheck.4.id=policy_workspace_lock\ncheck.4.status=ok\ncheck.4.detail=workspace lock disabled\ncheck.5.id=policy_excludes\ncheck.5.status=ok\ncheck.5.detail=excluded command patterns=0"
	if res.Message != want {
		t.Fatalf("unexpected sandbox check message: %q", res.Message)
	}
}

func TestPermissionsSummaryIncludesAliasesAndPrecedence(t *testing.T) {
	cmd := NewPermissionsCommand()
	state := &RuntimeState{PermissionMode: permissions.ModeAuto, PermissionRules: []string{"session:allow read:/repo"}}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"summary"}})
	if err != nil {
		t.Fatalf("permissions summary failed: %v", err)
	}
	want := "PERMISSIONS_SUMMARY\nmode=auto\nmode.aliases=accept_edits,accept-edits,acceptedits\nrules.count=1\nrules.precedence=policy>user>project>session\nrules.policy=0\nrules.user=0\nrules.project=0\nrules.session=1\ndenials.count=0\ndenials.groups=0"
	if res.Message != want {
		t.Fatalf("unexpected permissions summary: %q", res.Message)
	}
}

func TestExitCommandAndAlias(t *testing.T) {
	cmd := NewExitCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "exit"})
	if err != nil {
		t.Fatalf("exit failed: %v", err)
	}
	if res.Message != "EXIT_REQUEST\nrequested=true\ncount=1\ncommand=exit" {
		t.Fatalf("unexpected exit message: %q", res.Message)
	}

	res, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "quit"})
	if err != nil {
		t.Fatalf("quit failed: %v", err)
	}
	if res.Message != "EXIT_REQUEST\nrequested=true\ncount=2\ncommand=quit" {
		t.Fatalf("unexpected quit message: %q", res.Message)
	}
	if !state.ExitRequested || state.ExitCount != 2 {
		t.Fatalf("unexpected exit state: %+v", state)
	}
}

func TestPlanCommandFlow(t *testing.T) {
	cmd := NewPlanCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "plan", Args: []string{"fix", "lint"}})
	if err != nil {
		t.Fatalf("plan enable failed: %v", err)
	}
	if res.Message != "PLAN_ENABLE\nenabled=true\nenable_count=1\ndescription=fix lint\nquery_hint=true" {
		t.Fatalf("unexpected plan enable message: %q", res.Message)
	}

	openRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "plan", Args: []string{"open"}})
	if err != nil {
		t.Fatalf("plan open failed: %v", err)
	}
	if openRes.Message != "PLAN_OPEN\npath=.claude/plan.md\nopen_count=1" {
		t.Fatalf("unexpected plan open message: %q", openRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "plan", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("plan status failed: %v", err)
	}
	if statusRes.Message != "PLAN_STATUS\nenabled=true\nenable_count=1\nopen_count=1\nlast_description=fix lint" {
		t.Fatalf("unexpected plan status message: %q", statusRes.Message)
	}
}

func TestReviewCommandFlow(t *testing.T) {
	cmd := NewReviewCommand()
	state := &RuntimeState{}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "review"})
	if err != nil {
		t.Fatalf("review list failed: %v", err)
	}
	if listRes.Message != "REVIEW_REQUEST\nmode=list\ntarget=-\ncount=1\nchecklist.count=6\nchecklist.1=Summarize intent and changed scope\nchecklist.2=Call out assumptions and non-goals\nchecklist.3=Validate correctness and edge cases\nchecklist.4=Verify test coverage and missing cases\nchecklist.5=Assess security and data handling\nchecklist.6=Capture follow-up questions for clarification" {
		t.Fatalf("unexpected review list message: %q", listRes.Message)
	}

	prRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "review", Args: []string{"#123"}})
	if err != nil {
		t.Fatalf("review pr failed: %v", err)
	}
	if prRes.Message != "REVIEW_REQUEST\nmode=pr\ntarget=123\ncount=2\nchecklist.count=6\nchecklist.1=Summarize intent and changed scope\nchecklist.2=Call out assumptions and non-goals\nchecklist.3=Validate correctness and edge cases\nchecklist.4=Verify test coverage and missing cases\nchecklist.5=Assess security and data handling\nchecklist.6=Check merge readiness and unresolved feedback" {
		t.Fatalf("unexpected review pr message: %q", prRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "review", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("review status failed: %v", err)
	}
	if !strings.HasPrefix(statusRes.Message, "REVIEW_STATUS\ncount=2\nlast_target=123\nlast_mode=pr") {
		t.Fatalf("unexpected review status message: %q", statusRes.Message)
	}
}

func TestSessionCommandFlow(t *testing.T) {
	cmd := NewSessionCommand()
	state := &RuntimeState{TransportMode: "remote", SessionID: "session-abc", SessionPath: "/tmp/.alliecode/history/session-abc.jsonl"}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"set-url", "https://claude.ai/remote/session/abc"}})
	if err != nil {
		t.Fatalf("session set-url failed: %v", err)
	}
	if setRes.Message != "SESSION_URL_SET\nurl=https://claude.ai/remote/session/abc" {
		t.Fatalf("unexpected set-url message: %q", setRes.Message)
	}

	infoRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session"})
	if err != nil {
		t.Fatalf("session info failed: %v", err)
	}
	if infoRes.Message != "SESSION_INFO\nremote_mode=true\nsession_id=session-abc\nsession_path=/tmp/.alliecode/history/session-abc.jsonl\nurl=https://claude.ai/remote/session/abc\nqr_available=true\nviews=1\nhosting=false\nhost_addr=-\nconnected=false\nconnected_addr=-\ntoken_source=-\ntoken_prefix=-\nhost_count=0\nconnect_count=0\ndisconnect_count=0\nmode=-\nmanager_state=-\ntransport_state=-\nreconnecting=false\nreconnect_reason=-\nreconnect_error_class=-\nreconnect_count=0\nreconnect_detail=-" {
		t.Fatalf("unexpected session info message: %q", infoRes.Message)
	}
}

func TestSessionCommandStatusWithoutPersistedURL(t *testing.T) {
	cmd := NewSessionCommand()
	state := &RuntimeState{TransportMode: "remote", SessionID: "session-z", SessionPath: "/tmp/.alliecode/history/session-z.jsonl"}

	infoRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("session status failed: %v", err)
	}
	if infoRes.Message != "SESSION_INFO\nremote_mode=true\nsession_id=session-z\nsession_path=/tmp/.alliecode/history/session-z.jsonl\nurl=-\nqr_available=false\nviews=1\nhosting=false\nhost_addr=-\nconnected=false\nconnected_addr=-\ntoken_source=-\ntoken_prefix=-\nhost_count=0\nconnect_count=0\ndisconnect_count=0\nmode=-\nmanager_state=-\ntransport_state=-\nreconnecting=false\nreconnect_reason=-\nreconnect_error_class=-\nreconnect_count=0\nreconnect_detail=-" {
		t.Fatalf("unexpected session status message: %q", infoRes.Message)
	}
}

func TestSessionCommandHostDisconnectFlow(t *testing.T) {
	cmd := NewSessionCommand()
	state := &RuntimeState{TransportMode: "remote", SessionID: "session-hosted", SessionPath: "/tmp/.alliecode/history/session-hosted.jsonl"}

	hostRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"host", "127.0.0.1:0"}})
	if err != nil {
		t.Fatalf("session host failed: %v", err)
	}
	if !strings.HasPrefix(hostRes.Message, "SESSION_HOST\nhosting=true\naddr=127.0.0.1:") {
		t.Fatalf("unexpected host message: %q", hostRes.Message)
	}
	if !strings.Contains(hostRes.Message, "\ntoken_prefix=host-") {
		t.Fatalf("expected host token prefix in message: %q", hostRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("session status after host failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "\nhosting=true\n") {
		t.Fatalf("unexpected hosted status message: %q", statusRes.Message)
	}
	if !strings.Contains(statusRes.Message, "\nmode=host\n") || !strings.Contains(statusRes.Message, "\nmanager_state=closed\n") || !strings.Contains(statusRes.Message, "\ntransport_state=closed\n") {
		t.Fatalf("expected host lifecycle details in status: %q", statusRes.Message)
	}

	disconnectRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"disconnect"}})
	if err != nil {
		t.Fatalf("session disconnect failed: %v", err)
	}
	if !strings.Contains(disconnectRes.Message, "SESSION_DISCONNECT\nconnected=false\nwas_connected=false") {
		t.Fatalf("unexpected disconnect message: %q", disconnectRes.Message)
	}
	if state.SessionManager != nil || state.SessionTCPTransport != nil || state.SessionTransport != nil {
		t.Fatalf("expected disconnect to clear runtime manager/transport state")
	}

	finalStatus, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("session final status failed: %v", err)
	}
	if !strings.Contains(finalStatus.Message, "\nmode=-\nmanager_state=-\ntransport_state=-\n") {
		t.Fatalf("unexpected final status message: %q", finalStatus.Message)
	}
}

func TestSessionCommandConnectDisconnectFlow(t *testing.T) {
	cmd := NewSessionCommand()
	token := "supersecrettoken"

	hostTransport, err := remote.NewTCPTransport(remote.TCPTransportConfig{
		Mode:  remote.TCPTransportModeHost,
		Addr:  "127.0.0.1:0",
		Token: token,
	})
	if err != nil {
		t.Fatalf("host transport setup failed: %v", err)
	}
	defer func() { _ = hostTransport.Shutdown() }()

	hostMgr, err := remote.NewSessionManager(hostTransport, "session-client", remote.DefaultRetryPolicy(), time.Now().UTC())
	if err != nil {
		t.Fatalf("host session manager setup failed: %v", err)
	}
	hostCtx, cancelHost := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelHost()
	hostErrCh := make(chan error, 1)
	go func() {
		hostErrCh <- hostMgr.Connect(hostCtx)
	}()

	state := &RuntimeState{TransportMode: "remote", SessionID: "session-client", SessionPath: "/tmp/.alliecode/history/session-client.jsonl"}
	connectRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"connect", hostTransport.Addr(), token}})
	if err != nil {
		t.Fatalf("session connect failed: %v", err)
	}
	if err := <-hostErrCh; err != nil {
		t.Fatalf("host connect failed: %v", err)
	}
	wantConnect := "SESSION_CONNECT\nconnected=true\naddr=" + hostTransport.Addr() + "\ntoken_prefix=supers...\nconnect_count=1"
	if connectRes.Message != wantConnect {
		t.Fatalf("unexpected connect message: %q", connectRes.Message)
	}
	if strings.Contains(connectRes.Message, token) {
		t.Fatalf("connect output leaked full token: %q", connectRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("session status after connect failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "\nconnected=true\n") || !strings.Contains(statusRes.Message, "\nmode=client\n") || !strings.Contains(statusRes.Message, "\nmanager_state=active\n") || !strings.Contains(statusRes.Message, "\ntransport_state=connected\n") {
		t.Fatalf("unexpected connected status message: %q", statusRes.Message)
	}

	disconnectRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"disconnect"}})
	if err != nil {
		t.Fatalf("session disconnect failed: %v", err)
	}
	wantDisconnect := "SESSION_DISCONNECT\nconnected=false\nwas_connected=true\nprevious_addr=" + hostTransport.Addr() + "\ndisconnect_count=1"
	if disconnectRes.Message != wantDisconnect {
		t.Fatalf("unexpected disconnect message: %q", disconnectRes.Message)
	}

	finalStatus, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("session final status failed: %v", err)
	}
	if finalStatus.Message != "SESSION_INFO\nremote_mode=true\nsession_id=session-client\nsession_path=/tmp/.alliecode/history/session-client.jsonl\nurl=-\nqr_available=false\nviews=2\nhosting=false\nhost_addr=-\nconnected=false\nconnected_addr=-\ntoken_source=-\ntoken_prefix=-\nhost_count=0\nconnect_count=1\ndisconnect_count=1\nmode=-\nmanager_state=-\ntransport_state=-\nreconnecting=false\nreconnect_reason=-\nreconnect_error_class=-\nreconnect_count=0\nreconnect_detail=-" {
		t.Fatalf("unexpected final status message: %q", finalStatus.Message)
	}
}

func TestSessionCommandStatusIncludesManagerReconnectDiagnostics(t *testing.T) {
	cmd := NewSessionCommand()
	transport := remote.NewInMemoryTransport()
	mgr, err := remote.NewSessionManager(transport, "session-diag", remote.DefaultRetryPolicy(), time.Now().UTC())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}
	if err := mgr.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := mgr.HandleDisconnectWithCause(remote.ReconnectCause{
		Reason:     remote.ReconnectReasonDisconnect,
		ErrorClass: remote.TransportErrorClassAuth,
		Detail:     "token expired",
	}); err != nil {
		t.Fatalf("HandleDisconnectWithCause() error = %v", err)
	}

	state := &RuntimeState{
		TransportMode:  "remote",
		SessionID:      "session-diag",
		SessionPath:    "/tmp/.alliecode/history/session-diag.jsonl",
		SessionMode:    "client",
		SessionManager: mgr,
	}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("session status failed: %v", err)
	}
	if !strings.Contains(res.Message, "\nmanager_state=reconnecting\n") || !strings.Contains(res.Message, "\ntransport_state=reconnecting\n") {
		t.Fatalf("missing manager transport diagnostics: %q", res.Message)
	}
	if !strings.Contains(res.Message, "\nreconnecting=false\n") || !strings.Contains(res.Message, "\nreconnect_reason=disconnect\n") {
		t.Fatalf("missing reconnect reason diagnostics: %q", res.Message)
	}
	if !strings.Contains(res.Message, "\nreconnect_error_class=auth\n") || !strings.Contains(res.Message, "\nreconnect_detail=token expired") {
		t.Fatalf("missing reconnect error class diagnostics: %q", res.Message)
	}
}

func TestMCPCommandListAndStatusReflectManagerAuthTransitions(t *testing.T) {
	cmd := NewMCPCommand()
	manager, err := mcp.NewManager(context.Background(), []mcp.ServerConfig{{Name: "alpha", Disabled: true}})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if err := manager.AddServerConfig(mcp.ServerConfig{Name: "beta", Transport: mcp.TransportWebSocket, URL: "ws://localhost:4242"}); err != nil {
		t.Fatalf("AddServerConfig(beta) error = %v", err)
	}

	cmdCtx := Context{State: &RuntimeState{}, MCPManager: manager}
	listRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("mcp list failed: %v", err)
	}
	if !strings.Contains(listRes.Message, "server.2.name=beta") || !strings.Contains(listRes.Message, "server.2.transport=websocket") {
		t.Fatalf("missing beta transport in list: %q", listRes.Message)
	}
	if !strings.Contains(listRes.Message, "server.2.connection_state=pending") || !strings.Contains(listRes.Message, "server.2.auth_status=unauthenticated") {
		t.Fatalf("unexpected initial beta state in list: %q", listRes.Message)
	}

	manager.InvalidateServerCache("beta")
	statusRes, err := cmd.Execute(context.Background(), cmdCtx, Invocation{Name: "mcp", Args: []string{"status", "beta"}})
	if err != nil {
		t.Fatalf("mcp status beta failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "\nconnection_state=pending\n") || !strings.Contains(statusRes.Message, "\ntransport=websocket\n") {
		t.Fatalf("unexpected status after cache invalidation: %q", statusRes.Message)
	}
}

func TestSkillsCommandFlow(t *testing.T) {
	cmd := NewSkillsCommand()
	state := &RuntimeState{}

	addRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"add", "openclaw-status"}})
	if err != nil {
		t.Fatalf("skills add failed: %v", err)
	}
	if !strings.Contains(addRes.Message, "SKILLS_ADD") || !strings.Contains(addRes.Message, "source=state") || !strings.Contains(addRes.Message, "state=enabled") {
		t.Fatalf("unexpected skills add message: %q", addRes.Message)
	}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("skills list failed: %v", err)
	}
	if !strings.Contains(listRes.Message, "SKILLS_LIST") || !strings.Contains(listRes.Message, "skill.1.name=openclaw-status") || !strings.Contains(listRes.Message, "skill.1.source=state") || !strings.Contains(listRes.Message, "skill.1.state=enabled") {
		t.Fatalf("unexpected skills list message: %q", listRes.Message)
	}

	removeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"remove", "openclaw-status"}})
	if err != nil {
		t.Fatalf("skills remove failed: %v", err)
	}
	if removeRes.Message != "SKILLS_REMOVE\nname=openclaw-status\nremoved=true\ncount=0" {
		t.Fatalf("unexpected skills remove message: %q", removeRes.Message)
	}
}

func TestSkillsCommandDoctorIncludesDiagnosticsFields(t *testing.T) {
	cmd := NewSkillsCommand()
	state := &RuntimeState{
		Skills:              []string{"alpha", "beta"},
		SkillsSources:       map[string]string{"alpha": "file", "beta": "plugin"},
		SkillsOrigins:       map[string]string{"alpha": "/workspace/.alliecode/skills/alpha.md", "beta": "plugin:ops"},
		SkillsEnabled:       map[string]bool{"alpha": true, "beta": false},
		SkillsConflictCount: 2,
		SkillsSyncCount:     3,
	}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("skills doctor failed: %v", err)
	}
	for _, token := range []string{"SKILLS_DOCTOR", "enabled=1", "disabled=1", "conflicts=2", "quick_fix=/skills sync"} {
		if !strings.Contains(res.Message, token) {
			t.Fatalf("expected %q in doctor message: %q", token, res.Message)
		}
	}
}

func TestSkillsCommandUsageIncludesDoctorAndRepairModes(t *testing.T) {
	usage := NewSkillsCommand().Usage()
	for _, token := range []string{"doctor", "repair [sync|auto|dedupe]", "sync"} {
		if !strings.Contains(usage, token) {
			t.Fatalf("expected usage token %q in %q", token, usage)
		}
	}
}

func TestRewindCommandAndAlias(t *testing.T) {
	cmd := NewRewindCommand()
	state := &RuntimeState{CompactRequested: true, ResumeRequested: true}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "rewind", Args: []string{"head~1"}})
	if err != nil {
		t.Fatalf("rewind failed: %v", err)
	}
	if res.Message != "REWIND_REQUEST\ntarget=head~1\ncount=1\nrequested=true" {
		t.Fatalf("unexpected rewind message: %q", res.Message)
	}
	if state.CompactRequested || state.ResumeRequested {
		t.Fatalf("expected rewind to clear context flags")
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "checkpoint", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("checkpoint status failed: %v", err)
	}
	if statusRes.Message != "REWIND_STATUS\ncount=1\nlast_target=head~1" {
		t.Fatalf("unexpected rewind status: %q", statusRes.Message)
	}
}

func TestTagCommandToggleFlow(t *testing.T) {
	cmd := NewTagCommand()
	state := &RuntimeState{}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "tag", Args: []string{"bugfix"}})
	if err != nil {
		t.Fatalf("tag set failed: %v", err)
	}
	if setRes.Message != "TAG_SET\ntag=bugfix\nupdates=1" {
		t.Fatalf("unexpected tag set message: %q", setRes.Message)
	}

	removeRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "tag", Args: []string{"bugfix"}})
	if err != nil {
		t.Fatalf("tag remove failed: %v", err)
	}
	if removeRes.Message != "TAG_REMOVE\ntag=bugfix\nupdates=2" {
		t.Fatalf("unexpected tag remove message: %q", removeRes.Message)
	}
}

func TestRemoteEnvCommandFlow(t *testing.T) {
	cmd := NewRemoteEnvCommand()
	state := &RuntimeState{}

	listRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "remote-env", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("remote-env list failed: %v", err)
	}
	if listRes.Message != "REMOTE_ENV_LIST\ncount=3\nenv.1=default\nenv.2=code-review\nenv.3=hardened-linux" {
		t.Fatalf("unexpected remote-env list message: %q", listRes.Message)
	}

	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "remote-env", Args: []string{"set", "code-review"}})
	if err != nil {
		t.Fatalf("remote-env set failed: %v", err)
	}
	if setRes.Message != "REMOTE_ENV_SET\nenvironment=code-review\nupdates=1" {
		t.Fatalf("unexpected remote-env set message: %q", setRes.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "remote-env", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("remote-env status failed: %v", err)
	}
	if statusRes.Message != "REMOTE_ENV_STATUS\nenvironment=code-review\nupdates=1" {
		t.Fatalf("unexpected remote-env status message: %q", statusRes.Message)
	}
}

func TestSecurityReviewCommandFlow(t *testing.T) {
	cmd := NewSecurityReviewCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "security-review"})
	if err != nil {
		t.Fatalf("security-review failed: %v", err)
	}
	if res.Message != "SECURITY_REVIEW_REQUEST\ntarget=current-branch\ncount=1\nmode=focused" {
		t.Fatalf("unexpected security-review message: %q", res.Message)
	}

	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "security-review", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("security-review status failed: %v", err)
	}
	if statusRes.Message != "SECURITY_REVIEW_STATUS\ncount=1\nlast_target=current-branch" {
		t.Fatalf("unexpected security-review status message: %q", statusRes.Message)
	}
}

func TestLogoutStatusEnhancement(t *testing.T) {
	cmd := NewLogoutCommand()
	state := &RuntimeState{LoggedIn: true, ProviderName: "openai", AuthAccount: "dev@acme"}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "logout", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("logout status failed: %v", err)
	}
	if res.Message != "LOGOUT_STATUS\nlogged_in=true\nprovider=openai\naccount=dev@acme\nlogout_count=0" {
		t.Fatalf("unexpected logout status message: %q", res.Message)
	}
}

func TestLoginFlowFirstLoginIncludesNextActions(t *testing.T) {
	cmd := NewLoginCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "login"})
	if err != nil {
		t.Fatalf("login flow failed: %v", err)
	}
	if !strings.Contains(res.Message, "LOGIN_FLOW") || !strings.Contains(res.Message, "next=run_/model_to_select_a_model_then_/status") {
		t.Fatalf("expected coherent first-login next steps, got %q", res.Message)
	}
}

func TestProviderSwitchClearsLoginForFirstLoginCoherency(t *testing.T) {
	providerCmd := NewProviderCommand()
	statusCmd := NewStatusCommand()
	state := &RuntimeState{ProviderName: "openai", LoggedIn: true, AuthAccount: "dev@acme"}

	res, err := providerCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"set", "anthropic"}})
	if err != nil {
		t.Fatalf("provider set failed: %v", err)
	}
	if !strings.Contains(res.Message, "provider_ready=false") {
		t.Fatalf("expected provider readiness reset, got %q", res.Message)
	}
	if state.LoggedIn {
		t.Fatalf("expected logged_in cleared after provider switch")
	}
	if state.AuthAccount != "" {
		t.Fatalf("expected account cleared after provider switch, got %q", state.AuthAccount)
	}

	statusRes, err := statusCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "provider=anthropic") || !strings.Contains(statusRes.Message, "logged_in=false") || !strings.Contains(statusRes.Message, "provider_ready=false") {
		t.Fatalf("unexpected status message after provider switch: %q", statusRes.Message)
	}
}

func TestModelAliasMapsToCanonicalAnthropicModel(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "anthropic", LoggedIn: true}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/claude-opus-4-6"}})
	if err != nil {
		t.Fatalf("model alias set failed: %v", err)
	}
	if !strings.Contains(res.Message, "Model set to claude-opus-4-20250514") {
		t.Fatalf("expected canonical model in response, got %q", res.Message)
	}
	if state.Model != "claude-opus-4-20250514" {
		t.Fatalf("expected canonical model stored, got %q", state.Model)
	}
	if state.ProviderName != "anthropic" {
		t.Fatalf("expected anthropic provider retained, got %q", state.ProviderName)
	}
}

func TestModelUnknownFriendlyErrorSuggestsList(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "anthropic", LoggedIn: true}

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/not-a-real-model"}})
	if err == nil {
		t.Fatalf("expected unknown model error")
	}
	if !strings.Contains(err.Error(), "run /model list anthropic") {
		t.Fatalf("expected actionable unknown model error, got %v", err)
	}
}

func TestModelAliasSuggestionForOpus46Shorthand(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "anthropic", LoggedIn: true}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/opus-4-6"}})
	if err != nil {
		t.Fatalf("expected shorthand alias to map, got %v", err)
	}
	if !strings.Contains(res.Message, "Model set to claude-opus-4-20250514") {
		t.Fatalf("expected shorthand alias to map to canonical model, got %q", res.Message)
	}
}

func TestModelAliasMapsCommonAnthropicShorthands(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "anthropic", LoggedIn: true}

	tests := []struct {
		input string
		want  string
	}{
		{input: "anthropic/opus46", want: "claude-opus-4-20250514"},
		{input: "anthropic/claude46sonnet", want: "claude-sonnet-4-20250514"},
		{input: "anthropic/haiku35", want: "claude-haiku-3-5-20241022"},
	}

	for _, tc := range tests {
		res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{tc.input}})
		if err != nil {
			t.Fatalf("expected alias %q to map, got %v", tc.input, err)
		}
		if !strings.Contains(res.Message, "Model set to "+tc.want) {
			t.Fatalf("expected alias %q to resolve to %q, got %q", tc.input, tc.want, res.Message)
		}
	}
}

func TestModelUnknownAnthropicErrorSuggestsAliases(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "anthropic", LoggedIn: true}

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/claude-unknown"}})
	if err == nil {
		t.Fatalf("expected unknown anthropic model error")
	}
	if !strings.Contains(err.Error(), "run /model list anthropic") || !strings.Contains(err.Error(), "/model anthropic/opus") {
		t.Fatalf("expected actionable anthropic aliases in error, got %v", err)
	}
}

func TestProviderStatusAndLoginStatusAgreeForOllama(t *testing.T) {
	providerCmd := NewProviderCommand()
	loginCmd := NewLoginCommand()
	state := &RuntimeState{ProviderName: "ollama", LoggedIn: false}

	providerRes, err := providerCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("provider status failed: %v", err)
	}
	if !strings.Contains(providerRes.Message, "provider_ready=true") {
		t.Fatalf("expected ollama provider ready, got %q", providerRes.Message)
	}

	loginRes, err := loginCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "login", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("login status failed: %v", err)
	}
	if !strings.Contains(loginRes.Message, "logged_in=false") || !strings.Contains(loginRes.Message, "provider_ready=true") {
		t.Fatalf("expected login status to agree with provider readiness, got %q", loginRes.Message)
	}
}

func TestAdvisorBtwColorFastEffortCommands(t *testing.T) {
	state := &RuntimeState{}

	advisor := NewAdvisorCommand()
	res, err := advisor.Execute(context.Background(), Context{State: state}, Invocation{Name: "advisor", Args: []string{"opus"}})
	if err != nil {
		t.Fatalf("advisor set failed: %v", err)
	}
	if res.Message != "ADVISOR_SET\nmodel=opus\nactive=true\nupdates=1" {
		t.Fatalf("unexpected advisor set message: %q", res.Message)
	}

	btw := NewBtwCommand()
	res, err = btw.Execute(context.Background(), Context{State: state}, Invocation{Name: "btw", Args: []string{"quick", "question"}})
	if err != nil {
		t.Fatalf("btw failed: %v", err)
	}
	if res.Message != "BTW_RESULT\nquestion=quick question\ncount=1\nstatus=queued" {
		t.Fatalf("unexpected btw message: %q", res.Message)
	}

	color := NewColorCommand()
	res, err = color.Execute(context.Background(), Context{State: state}, Invocation{Name: "color", Args: []string{"blue"}})
	if err != nil {
		t.Fatalf("color set failed: %v", err)
	}
	if res.Message != "COLOR_SET\ncolor=blue\nset_count=1" {
		t.Fatalf("unexpected color set message: %q", res.Message)
	}

	fast := NewFastCommand()
	res, err = fast.Execute(context.Background(), Context{State: state}, Invocation{Name: "fast", Args: []string{"on"}})
	if err != nil {
		t.Fatalf("fast on failed: %v", err)
	}
	if res.Message != "FAST_SET\nenabled=true\ntoggles=1" {
		t.Fatalf("unexpected fast set message: %q", res.Message)
	}

	effort := NewEffortCommand()
	res, err = effort.Execute(context.Background(), Context{State: state}, Invocation{Name: "effort", Args: []string{"max"}})
	if err != nil {
		t.Fatalf("effort set failed: %v", err)
	}
	if res.Message != "EFFORT_SET\nvalue=max\nupdates=1" {
		t.Fatalf("unexpected effort set message: %q", res.Message)
	}
}

func TestChromeDesktopMobileCommands(t *testing.T) {
	state := &RuntimeState{}

	chrome := NewChromeCommand()
	res, err := chrome.Execute(context.Background(), Context{State: state}, Invocation{Name: "chrome", Args: []string{"install-extension"}})
	if err != nil {
		t.Fatalf("chrome install failed: %v", err)
	}
	if res.Message != "CHROME_EXTENSION\ninstalled=true\nactions=1" {
		t.Fatalf("unexpected chrome install message: %q", res.Message)
	}

	desktop := NewDesktopCommand()
	res, err = desktop.Execute(context.Background(), Context{State: state}, Invocation{Name: "app"})
	if err != nil {
		t.Fatalf("desktop alias failed: %v", err)
	}
	if res.Message != "DESKTOP_HANDOFF\ntarget=claude-desktop\ncount=1" {
		t.Fatalf("unexpected desktop handoff message: %q", res.Message)
	}

	mobile := NewMobileCommand()
	res, err = mobile.Execute(context.Background(), Context{State: state}, Invocation{Name: "android"})
	if err != nil {
		t.Fatalf("mobile alias failed: %v", err)
	}
	if res.Message != "MOBILE_QR\nplatform=android\ncount=1" {
		t.Fatalf("unexpected mobile qr message: %q", res.Message)
	}
}

func TestPluginReloadExportExtraUsageRateLimitPRWebSetupCommands(t *testing.T) {
	state := &RuntimeState{}
	workingDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	pluginsRoot := filepath.Join(workingDir, ".alliecode", "plugins")
	if err := os.MkdirAll(filepath.Join(pluginsRoot, "registry", "pr-comments"), 0o755); err != nil {
		t.Fatalf("mkdir plugin registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginsRoot, "registry", "pr-comments", "plugin.yaml"), []byte("id: pr-comments\nversion: 1.0.0\ncommands:\n  - name: pr-comments\n"), 0o644); err != nil {
		t.Fatalf("write plugin manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginsRoot, "index.yaml"), []byte("plugins:\n  - id: pr-comments\n    version: 1.0.0\n    source_path: registry/pr-comments\n"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workingDir, ".alliecode", "skills"), 0o755); err != nil {
		t.Fatalf("mkdir skills root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDir, ".alliecode", "skills", "pr-comments.md"), []byte("---\nname: pr-comments\ndescription: local skill\n---\nUse local skill\n"), 0o644); err != nil {
		t.Fatalf("write local skill: %v", err)
	}

	plugin := NewPluginCommand()
	res, err := plugin.Execute(context.Background(), Context{State: state}, Invocation{Name: "plugin", Args: []string{"install", "pr-comments"}})
	if err != nil {
		t.Fatalf("plugin install failed: %v", err)
	}
	if !strings.Contains(res.Message, "PLUGIN_INSTALL\nname=pr-comments\ninstalled=1\nenabled=1\npending_reload=true\nmutations=1") {
		t.Fatalf("unexpected plugin install message: %q", res.Message)
	}

	res, err = plugin.Execute(context.Background(), Context{State: state}, Invocation{Name: "plugin", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("plugin list failed: %v", err)
	}
	if !strings.Contains(res.Message, "PLUGIN_LIST") || !strings.Contains(res.Message, "conflicts=1") {
		t.Fatalf("expected plugin list conflict diagnostics, got %q", res.Message)
	}
	if !strings.Contains(res.Message, "conflict.1.name=pr-comments") {
		t.Fatalf("expected pr-comments conflict entry, got %q", res.Message)
	}

	reload := NewReloadPluginsCommand()
	res, err = reload.Execute(context.Background(), Context{State: state}, Invocation{Name: "reload-plugins"})
	if err != nil {
		t.Fatalf("reload-plugins failed: %v", err)
	}
	if res.Message != "RELOAD_PLUGINS_RESULT\npending_before=true\ninstalled=1\nenabled=1\nmarketplaces=0\nreload_count=1" {
		t.Fatalf("unexpected reload-plugins message: %q", res.Message)
	}

	exportCmd := NewExportCommand()
	res, err = exportCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "export", Args: []string{"notes"}})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if res.Message != "EXPORT_RESULT\npath=notes.txt\nformat=txt\ncount=1" {
		t.Fatalf("unexpected export message: %q", res.Message)
	}

	extra := NewExtraUsageCommand()
	res, err = extra.Execute(context.Background(), Context{State: state}, Invocation{Name: "extra-usage", Args: []string{"request"}})
	if err != nil {
		t.Fatalf("extra-usage request failed: %v", err)
	}
	if res.Message != "EXTRA_USAGE_REQUEST\nrequests=1\nenabled=false" {
		t.Fatalf("unexpected extra-usage message: %q", res.Message)
	}

	rate := NewRateLimitOptionsCommand()
	res, err = rate.Execute(context.Background(), Context{State: state}, Invocation{Name: "rate-limit-options", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("rate-limit-options status failed: %v", err)
	}
	if res.Message != "RATE_LIMIT_OPTIONS\ncount=3\nprompts=1\nlast_action=-\noption.1=cancel\noption.2=extra-usage\noption.3=upgrade" {
		t.Fatalf("unexpected rate-limit-options status message: %q", res.Message)
	}

	pr := NewPRCommentsCommand()
	res, err = pr.Execute(context.Background(), Context{State: state}, Invocation{Name: "pr-comments", Args: []string{"123"}})
	if err != nil {
		t.Fatalf("pr-comments failed: %v", err)
	}
	if res.Message != "PR_COMMENTS_SUMMARY\nref=123\nsource=ref\nfetches=1\nthreads=0" {
		t.Fatalf("unexpected pr-comments message: %q", res.Message)
	}

	web := NewWebSetupCommand()
	res, err = web.Execute(context.Background(), Context{State: state}, Invocation{Name: "web-setup"})
	if err != nil {
		t.Fatalf("web-setup failed: %v", err)
	}
	if res.Message != "WEB_SETUP_CONNECT\nconnected=true\nurl=https://claude.ai/code\ncount=1" {
		t.Fatalf("unexpected web-setup message: %q", res.Message)
	}
}

func TestInventoryGapCommandsDeterministicFlows(t *testing.T) {
	state := &RuntimeState{}
	ctx := Context{State: state}

	bridge := NewBridgeKickCommand()
	res, err := bridge.Execute(context.Background(), ctx, Invocation{Name: "bridge-kick", Args: []string{"close", "1002"}})
	if err != nil {
		t.Fatalf("bridge-kick close failed: %v", err)
	}
	if res.Message != "BRIDGE_KICK_APPLY\naction=close\ncode=1002\ncount=1" {
		t.Fatalf("unexpected bridge-kick message: %q", res.Message)
	}

	brief := NewBriefCommand()
	res, err = brief.Execute(context.Background(), ctx, Invocation{Name: "brief", Args: []string{"on"}})
	if err != nil {
		t.Fatalf("brief on failed: %v", err)
	}
	if res.Message != "BRIEF_SET\nenabled=true\ntoggles=1" {
		t.Fatalf("unexpected brief message: %q", res.Message)
	}

	commit := NewCommitCommand()
	res, err = commit.Execute(context.Background(), ctx, Invocation{Name: "commit", Args: []string{"feat:", "deterministic", "flow"}})
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	if !strings.Contains(res.Message, "COMMIT_PLAN") || !strings.Contains(res.Message, "suggested_message=feat: deterministic flow") {
		t.Fatalf("unexpected commit message: %q", res.Message)
	}

	pr := NewCommitPushPRCommand()
	res, err = pr.Execute(context.Background(), ctx, Invocation{Name: "commit-push-pr", Args: []string{"inventory", "gap"}})
	if err != nil {
		t.Fatalf("commit-push-pr failed: %v", err)
	}
	if res.Message != "COMMIT_PUSH_PR_PLAN\nmode=dry-run\nbranch=main\nbase=main\ntracked=true\nupstream=origin/main\ntitle=inventory gap\nstep.1.validate_git_state=blocked:no_changes\nstep.2.validate_staged_changes=ok\nstep.3.validate_branch_tracking=ok\nstep.4.commit=blocked\nstep.5.push=blocked\nstep.6.open_pr=blocked\nurl=-" {
		t.Fatalf("unexpected commit-push-pr message: %q", res.Message)
	}

	initVerifiers := NewInitVerifiersCommand()
	res, err = initVerifiers.Execute(context.Background(), ctx, Invocation{Name: "init-verifiers", Args: []string{"verifier-api"}})
	if err != nil {
		t.Fatalf("init-verifiers failed: %v", err)
	}
	if res.Message != "INIT_VERIFIERS_CREATED\nname=verifier-api\ncount=1" {
		t.Fatalf("unexpected init-verifiers message: %q", res.Message)
	}

	insights := NewInsightsCommand()
	res, err = insights.Execute(context.Background(), ctx, Invocation{Name: "insights", Args: []string{"report", "weekly"}})
	if err != nil {
		t.Fatalf("insights failed: %v", err)
	}
	if res.Message != "INSIGHTS_REPORT\nscope=weekly\ncount=1" {
		t.Fatalf("unexpected insights message: %q", res.Message)
	}

	passes := NewPassesCommand()
	res, err = passes.Execute(context.Background(), ctx, Invocation{Name: "passes", Args: []string{"claim", "2"}})
	if err != nil {
		t.Fatalf("passes claim failed: %v", err)
	}
	if res.Message != "PASSES_CLAIM\nclaimed=2\nremaining=3\nvisits=1" {
		t.Fatalf("unexpected passes message: %q", res.Message)
	}

	rename := NewRenameCommand()
	res, err = rename.Execute(context.Background(), ctx, Invocation{Name: "rename", Args: []string{"Sprint", "Review"}})
	if err != nil {
		t.Fatalf("rename failed: %v", err)
	}
	if res.Message != "RENAME_SET\nname=Sprint Review\ncount=1\nhistory_updated=0" {
		t.Fatalf("unexpected rename message: %q", res.Message)
	}

	stickers := NewStickersCommand()
	res, err = stickers.Execute(context.Background(), ctx, Invocation{Name: "stickers"})
	if err != nil {
		t.Fatalf("stickers failed: %v", err)
	}
	if res.Message != "STICKERS_ORDER\ncount=1\nurl=https://www.stickermule.com/claudecode" {
		t.Fatalf("unexpected stickers message: %q", res.Message)
	}

	thinkback := NewThinkbackCommand()
	res, err = thinkback.Execute(context.Background(), ctx, Invocation{Name: "thinkback", Args: []string{"generate"}})
	if err != nil {
		t.Fatalf("think-back generate failed: %v", err)
	}
	if res.Message != "THINKBACK_ACTION\naction=generate\ncount=1" {
		t.Fatalf("unexpected think-back message: %q", res.Message)
	}

	thinkbackPlay := NewThinkbackPlayCommand()
	res, err = thinkbackPlay.Execute(context.Background(), ctx, Invocation{Name: "thinkback-play"})
	if err != nil {
		t.Fatalf("thinkback-play failed: %v", err)
	}
	if res.Message != "THINKBACK_PLAY\ncount=1" {
		t.Fatalf("unexpected thinkback-play message: %q", res.Message)
	}

	teleport := NewTeleportCommand()
	res, err = teleport.Execute(context.Background(), ctx, Invocation{Name: "teleport", Args: []string{"agent://sandbox"}})
	if err != nil {
		t.Fatalf("teleport failed: %v", err)
	}
	if res.Message != "TELEPORT_SET\ntarget=agent://sandbox\ncount=1" {
		t.Fatalf("unexpected teleport message: %q", res.Message)
	}

	summary := NewSummaryCommand()
	res, err = summary.Execute(context.Background(), ctx, Invocation{Name: "summary", Args: []string{"refresh"}})
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}
	if res.Message != "SUMMARY_REFRESH\ncount=1" {
		t.Fatalf("unexpected summary message: %q", res.Message)
	}

	resetLimits := NewResetLimitsCommand()
	state.CostInputTokens = 99
	state.RateLimitPrompts = 7
	res, err = resetLimits.Execute(context.Background(), ctx, Invocation{Name: "reset-limits", Args: []string{"all"}})
	if err != nil {
		t.Fatalf("reset-limits failed: %v", err)
	}
	if res.Message != "RESET_LIMITS\ncount=1" {
		t.Fatalf("unexpected reset-limits message: %q", res.Message)
	}
	if state.CostInputTokens != 0 || state.RateLimitPrompts != 0 {
		t.Fatalf("expected limits reset, got state=%+v", state)
	}

	env := NewEnvCommand()
	res, err = env.Execute(context.Background(), ctx, Invocation{Name: "env", Args: []string{"set", "API_MODE", "offline"}})
	if err != nil {
		t.Fatalf("env set failed: %v", err)
	}
	if res.Message != "ENV_SET\nkey=API_MODE\nvalue=offline\nsets=1" {
		t.Fatalf("unexpected env set message: %q", res.Message)
	}

	res, err = env.Execute(context.Background(), ctx, Invocation{Name: "env", Args: []string{"get", "API_MODE"}})
	if err != nil {
		t.Fatalf("env get failed: %v", err)
	}
	if res.Message != "ENV_GET\nkey=API_MODE\nfound=true\nvalue=offline" {
		t.Fatalf("unexpected env get message: %q", res.Message)
	}
}

func TestCommitCommandBuildsActionablePlanFromDiffState(t *testing.T) {
	cmd := NewCommitCommand()
	state := &RuntimeState{
		ActiveBranch: "feature/commit-plan",
		Branches:     []string{"main", "feature/commit-plan"},
		DiffEntries: []DiffEntry{
			{Path: "internal/commands/handlers.go", Added: 40, Removed: 5, Modified: 3, Staged: true},
			{Path: "internal/commands/handlers_test.go", Added: 30, Removed: 10, Modified: 2},
			{Path: "docs/notes.md", Added: 5, Removed: 0, Modified: 0, Untracked: true},
		},
	}
	withWorkingDir(t, t.TempDir(), func() {
		res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "commit"})
		if err != nil {
			t.Fatalf("commit plan failed: %v", err)
		}
		if !strings.Contains(res.Message, "COMMIT_PLAN") || !strings.Contains(res.Message, "source=state") || !strings.Contains(res.Message, "step.1.action=Stage intended files with git add") || !strings.Contains(res.Message, "step.4.action=Run tests before finalizing commit") {
			t.Fatalf("unexpected commit plan message: %q", res.Message)
		}
		if !strings.Contains(res.Message, "suggested_message=feat:") {
			t.Fatalf("commit plan missing suggested message: %q", res.Message)
		}
	})
}

func TestCommitPushPRCommandDryRunAndForce(t *testing.T) {
	cmd := NewCommitPushPRCommand()
	state := &RuntimeState{
		ActiveBranch: "feature/parity",
		Branches:     []string{"main", "feature/parity"},
		ConfigValues: map[string]string{
			"git.upstream.feature/parity": "origin/feature/parity",
		},
		DiffEntries: []DiffEntry{{Path: "internal/commands/handlers.go", Added: 12, Removed: 2, Modified: 1, Staged: true}},
	}

	dryRun, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "commit-push-pr", Args: []string{"improve", "parity"}})
	if err != nil {
		t.Fatalf("commit-push-pr dry-run failed: %v", err)
	}
	if !strings.Contains(dryRun.Message, "mode=dry-run") || !strings.Contains(dryRun.Message, "step.4.commit=dry-run") || !strings.Contains(dryRun.Message, "step.6.open_pr=dry-run") {
		t.Fatalf("unexpected dry-run message: %q", dryRun.Message)
	}

	forced, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "commit-push-pr", Args: []string{"--force", "improve", "parity"}})
	if err != nil {
		t.Fatalf("commit-push-pr force failed: %v", err)
	}
	if !strings.Contains(forced.Message, "mode=execute") || !strings.Contains(forced.Message, "step.6.open_pr=executed") || !strings.Contains(forced.Message, "url=https://example.invalid/feature-parity/pull/1") {
		t.Fatalf("unexpected force message: %q", forced.Message)
	}
}

func TestPRCommentsCommandParsesMockFileAndThreads(t *testing.T) {
	cmd := NewPRCommentsCommand()
	dir := t.TempDir()
	mockPath := filepath.Join(dir, "comments.json")
	payload := `{"ref":"42","threads":[{"path":"internal/commands/handlers.go","line":100,"comments":[{"author":"alice","body":"Please rename this variable."},{"author":"bob","body":"Done in latest patch.","reply_to":"1"}]}]}`
	if err := os.WriteFile(mockPath, []byte(payload), 0o600); err != nil {
		t.Fatalf("write mock payload: %v", err)
	}

	res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "pr-comments", Args: []string{"file=" + mockPath}})
	if err != nil {
		t.Fatalf("pr-comments parse failed: %v", err)
	}
	if !strings.Contains(res.Message, "PR_COMMENTS_SUMMARY") || !strings.Contains(res.Message, "threads=1") || !strings.Contains(res.Message, "thread.1.comment.2.depth=1") {
		t.Fatalf("unexpected pr-comments summary: %q", res.Message)
	}
}

func TestPRCommentsCommandParsesStateMockPayload(t *testing.T) {
	cmd := NewPRCommentsCommand()
	state := &RuntimeState{ConfigValues: map[string]string{
		"pr-comments.mock": `{"ref":"state-7","threads":[{"path":"internal/commands/review.go","line":21,"comments":[{"author":"maintainer","body":"nit: clarify phrasing"}]}]}`,
	}}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "pr-comments", Args: []string{"state"}})
	if err != nil {
		t.Fatalf("pr-comments state payload failed: %v", err)
	}
	if !strings.Contains(res.Message, "ref=state-7") || !strings.Contains(res.Message, "source=state") || !strings.Contains(res.Message, "thread.1.path=internal/commands/review.go") {
		t.Fatalf("unexpected state payload summary: %q", res.Message)
	}
}

func TestRenameCommandPersistsTitleToMatchingHistoryEntry(t *testing.T) {
	cmd := NewRenameCommand()
	state := &RuntimeState{
		SessionID: "session-2",
		HistoryEntries: []HistoryEntry{
			{ID: "session-1", Title: "Old One", CreatedAt: "2026-01-01T00:00:00Z"},
			{ID: "session-2", Title: "Old Two", CreatedAt: "2026-01-02T00:00:00Z"},
		},
	}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "rename", Args: []string{"New", "Title"}})
	if err != nil {
		t.Fatalf("rename failed: %v", err)
	}
	if !strings.Contains(res.Message, "history_updated=1") {
		t.Fatalf("expected history update in output, got: %q", res.Message)
	}
	if got := state.HistoryEntries[1].Title; got != "New Title" {
		t.Fatalf("expected matching history title update, got %q", got)
	}
}

func TestReviewCommandClassifiesFileModeChecklist(t *testing.T) {
	cmd := NewReviewCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "review", Args: []string{"internal/commands/handlers.go"}})
	if err != nil {
		t.Fatalf("review file mode failed: %v", err)
	}
	if !strings.Contains(res.Message, "mode=file") || !strings.Contains(res.Message, "checklist.6=Confirm file-level conventions and ownership") {
		t.Fatalf("unexpected file mode review output: %q", res.Message)
	}
}

func TestCommitStatusUsesGitPorcelainCounts(t *testing.T) {
	repo := initGitRepoWithRemote(t)
	withWorkingDir(t, repo, func() {
		mustRunGit(t, repo, "checkout", "-b", "feature/porcelain")

		if err := os.WriteFile(filepath.Join(repo, "staged.txt"), []byte("staged\n"), 0o600); err != nil {
			t.Fatalf("write staged file: %v", err)
		}
		mustRunGit(t, repo, "add", "staged.txt")

		if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("base\nchange\n"), 0o600); err != nil {
			t.Fatalf("write tracked file: %v", err)
		}

		if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("draft\n"), 0o600); err != nil {
			t.Fatalf("write untracked file: %v", err)
		}

		cmd := NewCommitCommand()
		res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "commit", Args: []string{"status"}})
		if err != nil {
			t.Fatalf("commit status failed: %v", err)
		}
		if !strings.Contains(res.Message, "COMMIT_STATUS") || !strings.Contains(res.Message, "source=git") || !strings.Contains(res.Message, "staged=1") || !strings.Contains(res.Message, "unstaged=1") || !strings.Contains(res.Message, "untracked=1") {
			t.Fatalf("unexpected git-backed commit status: %q", res.Message)
		}
	})
}

func TestCommitPushPRStatusDetectsTrackingAndRemoteReadiness(t *testing.T) {
	repo := initGitRepoWithRemote(t)
	withWorkingDir(t, repo, func() {
		cmd := NewCommitPushPRCommand()
		res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "commit-push-pr", Args: []string{"status"}})
		if err != nil {
			t.Fatalf("commit-push-pr status failed: %v", err)
		}
		if !strings.Contains(res.Message, "COMMIT_PUSH_PR_STATUS") || !strings.Contains(res.Message, "source=git") || !strings.Contains(res.Message, "tracked=true") || !strings.Contains(res.Message, "upstream=origin/main") || !strings.Contains(res.Message, "remote_ready=true") {
			t.Fatalf("unexpected git-backed commit-push-pr status: %q", res.Message)
		}

		mustRunGit(t, repo, "checkout", "-b", "feature/no-upstream")
		res, err = cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "commit-push-pr", Args: []string{"status"}})
		if err != nil {
			t.Fatalf("commit-push-pr status on untracked branch failed: %v", err)
		}
		if !strings.Contains(res.Message, "tracked=false") || !strings.Contains(res.Message, "remote_ready=false") {
			t.Fatalf("expected untracked/unready status, got: %q", res.Message)
		}
	})
}

func TestReviewStatusInfersModeAndSectionsFromTargetType(t *testing.T) {
	cmd := NewReviewCommand()
	state := &RuntimeState{}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "review", Args: []string{"status", "#321"}})
	if err != nil {
		t.Fatalf("review status with target failed: %v", err)
	}
	if !strings.Contains(res.Message, "REVIEW_STATUS") || !strings.Contains(res.Message, "mode=pr") || !strings.Contains(res.Message, "target=321") || !strings.Contains(res.Message, "section.1.title=Intent") || !strings.Contains(res.Message, "section.4.title=Mode-specific") {
		t.Fatalf("unexpected review status output: %q", res.Message)
	}
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	fn()
}

func mustRunGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func initGitRepoWithRemote(t *testing.T) string {
	t.Helper()
	remoteParent := t.TempDir()
	remoteDir := filepath.Join(remoteParent, "origin.git")
	mustRunGit(t, remoteParent, "init", "--bare", remoteDir)

	repo := t.TempDir()
	mustRunGit(t, repo, "init", "-b", "main")
	mustRunGit(t, repo, "config", "user.name", "Test User")
	mustRunGit(t, repo, "config", "user.email", "test@example.com")
	mustRunGit(t, repo, "remote", "add", "origin", remoteDir)

	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("base\n"), 0o600); err != nil {
		t.Fatalf("write base file: %v", err)
	}
	mustRunGit(t, repo, "add", "tracked.txt")
	mustRunGit(t, repo, "commit", "-m", "init")
	mustRunGit(t, repo, "push", "-u", "origin", "main")
	return repo
}

func TestInventoryGapCommandsUsageValidation(t *testing.T) {
	ctx := Context{State: &RuntimeState{}}

	_, err := NewBridgeKickCommand().Execute(context.Background(), ctx, Invocation{Name: "bridge-kick", Args: []string{"poll"}})
	if err == nil || err.Error() != "usage: /bridge-kick [status|close <code>|poll <status>|reconnect]" {
		t.Fatalf("expected bridge-kick usage error, got %v", err)
	}

	_, err = NewInsightsCommand().Execute(context.Background(), ctx, Invocation{Name: "insights", Args: []string{"export"}})
	if err == nil || err.Error() != "usage: /insights [status|report [scope]]" {
		t.Fatalf("expected insights usage error, got %v", err)
	}

	_, err = NewEnvCommand().Execute(context.Background(), ctx, Invocation{Name: "env", Args: []string{"set", "ONLY_KEY"}})
	if err == nil || err.Error() != "usage: /env [status|set <key> <value>|get <key>]" {
		t.Fatalf("expected env usage error, got %v", err)
	}
}

func TestSandboxStatusScopeRestrictionsReflectPolicy(t *testing.T) {
	sandboxCmd := NewSandboxCommand()
	state := &RuntimeState{SandboxMode: "read-only", SandboxWorkspaceLocked: false}
	res, err := sandboxCmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("sandbox status failed: %v", err)
	}
	if !strings.Contains(res.Message, "scope.restrictions=read_only_filesystem") {
		t.Fatalf("expected read_only_filesystem restriction in %q", res.Message)
	}
}
