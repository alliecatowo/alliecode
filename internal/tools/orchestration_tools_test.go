package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestSkillToolListInvokeHistory(t *testing.T) {
	resetOrchestrationStateForTests()
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".claude", "skills", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Demo\nUse this skill."), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	tool := &SkillTool{}
	toolCtx := types.ToolContext{WorkingDir: dir}

	res, err := tool.Execute(context.Background(), []byte(`{"action":"list"}`), toolCtx)
	if err != nil || res.IsError {
		t.Fatalf("skill list failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, `"demo"`) {
		t.Fatalf("expected demo skill in list output: %s", res.Content)
	}
	if !strings.Contains(res.Content, `"plugin_origins"`) {
		t.Fatalf("expected plugin_origins field in list output: %s", res.Content)
	}
	if !strings.Contains(res.Content, `"conflict_diagnostics"`) {
		t.Fatalf("expected conflict diagnostics field in list output: %s", res.Content)
	}

	res, err = tool.Execute(context.Background(), []byte(`{"action":"invoke","skill":"demo","args":"--check"}`), toolCtx)
	if err != nil || res.IsError {
		t.Fatalf("skill invoke failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, `"success":true`) {
		t.Fatalf("expected success invoke output: %s", res.Content)
	}

	res, err = tool.Execute(context.Background(), []byte(`{"action":"history"}`), toolCtx)
	if err != nil || res.IsError {
		t.Fatalf("skill history failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, `"invocations":1`) {
		t.Fatalf("expected invocation count in history: %s", res.Content)
	}
}

func TestConfigToolSetGetListUnset(t *testing.T) {
	resetOrchestrationStateForTests()
	dir := t.TempDir()
	tool := &ConfigTool{}
	ctx := types.ToolContext{WorkingDir: dir}

	setRes, err := tool.Execute(context.Background(), []byte(`{"operation":"set","setting":"theme","value":"solarized"}`), ctx)
	if err != nil || setRes.IsError {
		t.Fatalf("config set failed: err=%v content=%q", err, setRes.Content)
	}

	getRes, err := tool.Execute(context.Background(), []byte(`{"operation":"get","setting":"theme"}`), ctx)
	if err != nil || getRes.IsError {
		t.Fatalf("config get failed: err=%v content=%q", err, getRes.Content)
	}
	if !strings.Contains(getRes.Content, `"value":"solarized"`) {
		t.Fatalf("expected stored value in get output: %s", getRes.Content)
	}

	listRes, err := tool.Execute(context.Background(), []byte(`{"operation":"list"}`), ctx)
	if err != nil || listRes.IsError {
		t.Fatalf("config list failed: err=%v content=%q", err, listRes.Content)
	}
	if !strings.Contains(listRes.Content, `"setting":"theme"`) {
		t.Fatalf("expected theme setting in list output: %s", listRes.Content)
	}

	unsetRes, err := tool.Execute(context.Background(), []byte(`{"operation":"unset","setting":"theme"}`), ctx)
	if err != nil || unsetRes.IsError {
		t.Fatalf("config unset failed: err=%v content=%q", err, unsetRes.Content)
	}
}

func TestLSPToolHoverAndDefinition(t *testing.T) {
	resetOrchestrationStateForTests()
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	src := "package main\n\nfunc Demo() {}\n\nfunc use() { Demo() }\n"
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	tool := &LSPTool{}
	ctx := types.ToolContext{WorkingDir: dir}

	hover, err := tool.Execute(context.Background(), []byte(`{"operation":"hover","file_path":"main.go","line":5,"character":14}`), ctx)
	if err != nil || hover.IsError {
		t.Fatalf("hover failed: err=%v content=%q", err, hover.Content)
	}
	if !strings.Contains(hover.Content, `"symbol":"Demo"`) {
		t.Fatalf("expected symbol Demo in hover output: %s", hover.Content)
	}

	def, err := tool.Execute(context.Background(), []byte(`{"operation":"goToDefinition","file_path":"main.go","line":5,"character":14}`), ctx)
	if err != nil || def.IsError {
		t.Fatalf("definition failed: err=%v content=%q", err, def.Content)
	}
	if !strings.Contains(def.Content, `"result_count":1`) {
		t.Fatalf("expected at least one definition: %s", def.Content)
	}
}

func TestWorktreeEnterExitLifecycle(t *testing.T) {
	resetOrchestrationStateForTests()
	dir := t.TempDir()
	ctx := types.ToolContext{WorkingDir: dir}

	enterRes, err := (&EnterWorktreeTool{}).Execute(context.Background(), []byte(`{"name":"feature-x"}`), ctx)
	if err != nil || enterRes.IsError {
		t.Fatalf("worktree enter failed: err=%v content=%q", err, enterRes.Content)
	}

	var enterOut map[string]any
	if err := json.Unmarshal([]byte(enterRes.Content), &enterOut); err != nil {
		t.Fatalf("invalid enter json: %v", err)
	}
	path, _ := enterOut["worktree_path"].(string)
	if path == "" {
		t.Fatalf("expected worktree path in enter output")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected worktree directory to exist: %v", err)
	}

	exitRes, err := (&ExitWorktreeTool{}).Execute(context.Background(), []byte(`{"action":"remove","discard_changes":true}`), ctx)
	if err != nil || exitRes.IsError {
		t.Fatalf("worktree exit failed: err=%v content=%q", err, exitRes.Content)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected worktree directory removed, stat err=%v", err)
	}
}

func TestTeamAndSendMessageTools(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{WorkingDir: t.TempDir()}

	createRes, err := (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"alpha","members":["dev1","dev2"]}`), ctx)
	if err != nil || createRes.IsError {
		t.Fatalf("team create failed: err=%v content=%q", err, createRes.Content)
	}

	sendRes, err := (&SendMessageTool{}).Execute(context.Background(), []byte(`{"to":"*","summary":"sync","message":"standup now"}`), ctx)
	if err != nil || sendRes.IsError {
		t.Fatalf("send_message failed: err=%v content=%q", err, sendRes.Content)
	}
	if !strings.Contains(sendRes.Content, `"stored":2`) {
		t.Fatalf("expected two stored broadcast messages: %s", sendRes.Content)
	}

	statusBeforeDelete, err := (&TeamStatusTool{}).Execute(context.Background(), []byte(`{"team_name":"alpha"}`), ctx)
	if err != nil || statusBeforeDelete.IsError {
		t.Fatalf("team status before delete failed: err=%v content=%q", err, statusBeforeDelete.Content)
	}
	if !strings.Contains(statusBeforeDelete.Content, `"messages":[`) {
		t.Fatalf("expected message history in team_status output: %s", statusBeforeDelete.Content)
	}

	deleteRes, err := (&TeamDeleteTool{}).Execute(context.Background(), []byte(`{"team_name":"alpha"}`), ctx)
	if err != nil || deleteRes.IsError {
		t.Fatalf("team delete failed: err=%v content=%q", err, deleteRes.Content)
	}

	statusRes, err := (&TeamStatusTool{}).Execute(context.Background(), []byte(`{"team_name":"alpha"}`), ctx)
	if err != nil || statusRes.IsError {
		t.Fatalf("team status failed: err=%v content=%q", err, statusRes.Content)
	}
	if !strings.Contains(statusRes.Content, `"status":"canceled"`) {
		t.Fatalf("expected canceled status in team_status output: %s", statusRes.Content)
	}
}

func TestTeamDeleteArchiveMode(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{WorkingDir: t.TempDir()}

	createRes, err := (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"gamma","members":["dev1"]}`), ctx)
	if err != nil || createRes.IsError {
		t.Fatalf("team create failed: err=%v content=%q", err, createRes.Content)
	}

	archiveRes, err := (&TeamDeleteTool{}).Execute(context.Background(), []byte(`{"team_name":"gamma","archive":true,"reason":"done"}`), ctx)
	if err != nil || archiveRes.IsError {
		t.Fatalf("team archive failed: err=%v content=%q", err, archiveRes.Content)
	}
	if !strings.Contains(archiveRes.Content, `"status":"archived"`) {
		t.Fatalf("expected archived status in delete output: %s", archiveRes.Content)
	}

	statusRes, err := (&TeamStatusTool{}).Execute(context.Background(), []byte(`{"team_name":"gamma"}`), ctx)
	if err != nil || statusRes.IsError {
		t.Fatalf("team status failed: err=%v content=%q", err, statusRes.Content)
	}
	if !strings.Contains(statusRes.Content, `"status":"archived"`) {
		t.Fatalf("expected archived status in team_status output: %s", statusRes.Content)
	}
}

func TestTeamToolsListAndUpdate(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{WorkingDir: t.TempDir()}

	createRes, err := (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"beta","members":["dev1"]}`), ctx)
	if err != nil || createRes.IsError {
		t.Fatalf("team create failed: err=%v content=%q", err, createRes.Content)
	}

	updateRes, err := (&TeamUpdateTool{}).Execute(context.Background(), []byte(`{"team_name":"beta","description":"updated","members":["dev1","dev2"]}`), ctx)
	if err != nil || updateRes.IsError {
		t.Fatalf("team update failed: err=%v content=%q", err, updateRes.Content)
	}
	if !strings.Contains(updateRes.Content, `"description":"updated"`) {
		t.Fatalf("expected updated description: %s", updateRes.Content)
	}

	listRes, err := (&TeamListTool{}).Execute(context.Background(), []byte(`{}`), ctx)
	if err != nil || listRes.IsError {
		t.Fatalf("team list failed: err=%v content=%q", err, listRes.Content)
	}
	if !strings.Contains(listRes.Content, `"total":1`) {
		t.Fatalf("expected one team in list: %s", listRes.Content)
	}
}

func TestCronToolsCreateListDelete(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{WorkingDir: t.TempDir()}

	createRes, err := (&CronCreateTool{}).Execute(context.Background(), []byte(`{"cron":"*/5 * * * *","prompt":"check status"}`), ctx)
	if err != nil || createRes.IsError {
		t.Fatalf("cron create failed: err=%v content=%q", err, createRes.Content)
	}
	if !strings.Contains(createRes.Content, `"id":"cron-0001"`) {
		t.Fatalf("expected deterministic cron id: %s", createRes.Content)
	}

	listRes, err := (&CronListTool{}).Execute(context.Background(), []byte(`{}`), ctx)
	if err != nil || listRes.IsError {
		t.Fatalf("cron list failed: err=%v content=%q", err, listRes.Content)
	}
	if !strings.Contains(listRes.Content, `"cron-0001"`) {
		t.Fatalf("expected cron in list output: %s", listRes.Content)
	}

	deleteRes, err := (&CronDeleteTool{}).Execute(context.Background(), []byte(`{"id":"cron-0001"}`), ctx)
	if err != nil || deleteRes.IsError {
		t.Fatalf("cron delete failed: err=%v content=%q", err, deleteRes.Content)
	}
}

func TestRemoteTriggerToolLifecycle(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{WorkingDir: t.TempDir()}
	tool := &RemoteTriggerTool{}

	createRes, err := tool.Execute(context.Background(), []byte(`{"action":"create","body":{"cron":"0 9 * * *"}}`), ctx)
	if err != nil || createRes.IsError {
		t.Fatalf("remote trigger create failed: err=%v content=%q", err, createRes.Content)
	}
	if !strings.Contains(createRes.Content, `"status":201`) {
		t.Fatalf("expected create status 201: %s", createRes.Content)
	}

	runRes, err := tool.Execute(context.Background(), []byte(`{"action":"run","trigger_id":"trigger-0001"}`), ctx)
	if err != nil || runRes.IsError {
		t.Fatalf("remote trigger run failed: err=%v content=%q", err, runRes.Content)
	}
	if !strings.Contains(runRes.Content, `"run_count":1`) {
		t.Fatalf("expected run_count increment: %s", runRes.Content)
	}

	listRes, err := tool.Execute(context.Background(), []byte(`{"action":"list"}`), ctx)
	if err != nil || listRes.IsError {
		t.Fatalf("remote trigger list failed: err=%v content=%q", err, listRes.Content)
	}
	if !strings.Contains(listRes.Content, `"trigger-0001"`) {
		t.Fatalf("expected trigger in list: %s", listRes.Content)
	}
}

func TestBriefToolWithAttachment(t *testing.T) {
	resetOrchestrationStateForTests()
	dir := t.TempDir()
	attachment := filepath.Join(dir, "artifact.txt")
	if err := os.WriteFile(attachment, []byte("artifact"), 0o644); err != nil {
		t.Fatalf("write attachment: %v", err)
	}

	tool := &BriefTool{}
	ctx := types.ToolContext{WorkingDir: dir}
	res, err := tool.Execute(context.Background(), []byte(`{"message":"done","status":"normal","attachments":["artifact.txt"]}`), ctx)
	if err != nil || res.IsError {
		t.Fatalf("brief failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, `"attachment_count":1`) {
		t.Fatalf("expected one attachment in output: %s", res.Content)
	}
}
