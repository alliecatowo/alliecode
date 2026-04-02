package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/tools"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestToolContractAskUserQuestionEdgeResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty questions",
			input:   `{"questions":[]}`,
			wantErr: "questions must not be empty",
		},
		{
			name:    "single-select with multiple answers",
			input:   `{"questions":[{"id":"q1","question":"pick","options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]}],"answers":[{"question_id":"q1","option_ids":["a","b"]}]}`,
			wantErr: `answers[0] has multiple options for single-select question "q1"`,
		},
		{
			name:    "unknown option id",
			input:   `{"questions":[{"id":"q1","question":"pick","options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]}],"answers":[{"question_id":"q1","option_ids":["c"]}]}`,
			wantErr: `answers[0] contains unknown option id "c"`,
		},
	}

	tool := &tools.AskUserQuestionTool{}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := tool.Execute(context.Background(), []byte(tc.input), types.ToolContext{})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !res.IsError {
				t.Fatalf("expected IsError=true, got false with content=%q", res.Content)
			}
			if res.Content != tc.wantErr {
				t.Fatalf("error content mismatch\n got: %q\nwant: %q", res.Content, tc.wantErr)
			}
		})
	}
}

func TestToolContractTaskOutputEdgeResponses(t *testing.T) {
	t.Parallel()

	tool := &tools.TaskOutputTool{}

	missingTask, err := tool.Execute(context.Background(), []byte(`{"task_id":"missing","block":false}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute(missing) error = %v", err)
	}
	if missingTask.IsError {
		t.Fatalf("missing task should not be error, content=%q", missingTask.Content)
	}
	if missingTask.Content != `{"retrieval_status":"not_found","output_status":"empty","task":null,"runtime":{"source":"adapter","blocked":false,"duration_ms":0,"found":false,"terminal":false,"requested_task_id":"missing","poll_used_ms":200}}` {
		t.Fatalf("unexpected missing task payload = %q", missingTask.Content)
	}

	negativePoll, err := tool.Execute(context.Background(), []byte(`{"task_id":"missing","poll_ms":-1}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute(negative poll) error = %v", err)
	}
	if !negativePoll.IsError {
		t.Fatalf("negative poll should be IsError=true")
	}
	if negativePoll.Content != "poll_ms must be >= 0" {
		t.Fatalf("unexpected negative poll error = %q", negativePoll.Content)
	}

	missingID, err := tool.Execute(context.Background(), []byte(`{"block":false}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute(missing id) error = %v", err)
	}
	if !missingID.IsError {
		t.Fatalf("missing task_id should be IsError=true")
	}
	if missingID.Content != "task_id is required" {
		t.Fatalf("unexpected missing task_id error = %q", missingID.Content)
	}
}

func TestToolContractSleepEdgeResponses(t *testing.T) {
	t.Parallel()

	tool := &tools.SleepTool{}

	zeroDuration, err := tool.Execute(context.Background(), []byte(`{"duration_ms":0}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute(zero duration) error = %v", err)
	}
	if zeroDuration.IsError {
		t.Fatalf("zero duration should be bounded success")
	}
	if zeroDuration.Content != `{"status":"completed","requested_ms":0,"slept_ms":1,"bounded":true}` {
		t.Fatalf("unexpected zero-duration error = %q", zeroDuration.Content)
	}

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	tooLong, err := tool.Execute(cancelledCtx, []byte(`{"duration_ms":600001}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute(too long) error = %v", err)
	}
	if !tooLong.IsError {
		t.Fatalf("cancelled context should report IsError=true")
	}
	if tooLong.Content != `{"status":"cancelled","requested_ms":600001,"slept_ms":600000,"bounded":true}` {
		t.Fatalf("unexpected too-long error = %q", tooLong.Content)
	}
}

func TestToolContractBashRiskPayloadDenied(t *testing.T) {
	t.Parallel()

	in, err := json.Marshal(map[string]any{"command": "rm -rf /tmp/risky"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	res, err := (&tools.BashTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got false")
	}
	if !strings.Contains(res.Content, "Command denied: potential destructive operation detected") {
		t.Fatalf("unexpected denial content = %q", res.Content)
	}
}

func TestToolContractBashRiskPayloadAmbiguousParseDenied(t *testing.T) {
	t.Parallel()

	in, err := json.Marshal(map[string]any{"command": "echo ok # ' \"\nwhoami"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	res, err := (&tools.BashTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got false")
	}
	if !strings.Contains(res.Content, "[bash_preflight]") {
		t.Fatalf("expected precheck block in content = %q", res.Content)
	}
	if !strings.Contains(res.Content, "rule_id: comment_quote_desync") {
		t.Fatalf("expected comment_quote_desync rule in content = %q", res.Content)
	}
}

func TestToolContractWriteFailsOnStaleReadMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "stale.txt")
	if err := os.WriteFile(path, []byte("initial"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldMtime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, oldMtime, oldMtime); err != nil {
		t.Fatalf("Chtimes(old) error = %v", err)
	}

	meta := strings.Join([]string{
		"[read_metadata]",
		"path: " + path,
		"mtime_unix_ms: " + strconv.FormatInt(oldMtime.UnixMilli(), 10),
		"size_bytes: 7",
		"offset: 0",
		"limit: all",
		"partial: false",
		"[/read_metadata]",
	}, "\n")

	newMtime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(path, newMtime, newMtime); err != nil {
		t.Fatalf("Chtimes(new) error = %v", err)
	}

	in, err := json.Marshal(map[string]any{"file_path": path, "content": "next"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	res, err := (&tools.FileWriteTool{}).Execute(context.Background(), in, types.ToolContext{
		WorkingDir: dir,
		Messages: []types.Message{{
			Role:    types.RoleAssistant,
			Content: []types.ContentBlock{{Type: types.ContentToolResult, Content: meta}},
		}},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got false")
	}
	if !strings.Contains(res.Content, "file has been modified since read") {
		t.Fatalf("unexpected stale-read content = %q", res.Content)
	}
	if !strings.Contains(res.Content, "[stale_read]") {
		t.Fatalf("expected stale_read block in content = %q", res.Content)
	}
}

func TestToolContractGrepTruncationMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	var b strings.Builder
	for i := 0; i < 260; i++ {
		b.WriteString("hit line\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	in, err := json.Marshal(map[string]any{"pattern": "hit", "path": dir, "output_mode": "content"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	res, err := (&tools.GrepTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if res.IsError {
		t.Fatalf("expected IsError=false, got content=%q", res.Content)
	}
	if !strings.Contains(res.Content, "[truncation_metadata]") {
		t.Fatalf("expected truncation metadata, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "line_limit: 200") || !strings.Contains(res.Content, "char_limit: 100000") {
		t.Fatalf("expected line/char limits in content = %q", res.Content)
	}
}

func TestToolContractGlobTruncationMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for i := 0; i < 520; i++ {
		name := filepath.Join(dir, "f"+strconv.Itoa(i)+".txt")
		if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	in, err := json.Marshal(map[string]any{"pattern": "*.txt", "path": dir})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	res, err := (&tools.GlobTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if res.IsError {
		t.Fatalf("expected IsError=false, got content=%q", res.Content)
	}
	if !strings.Contains(res.Content, "[truncation_metadata]") {
		t.Fatalf("expected truncation metadata, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "line_limit: 500") || !strings.Contains(res.Content, "char_limit: 100000") {
		t.Fatalf("expected line/char limits in content = %q", res.Content)
	}
	if !strings.Contains(res.Content, "[glob_result]") || !strings.Contains(res.Content, "count: 520") {
		t.Fatalf("expected glob_result metadata in content = %q", res.Content)
	}
}
