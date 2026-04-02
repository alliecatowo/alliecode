package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestFileReadIncludesReadMetadataBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	in, _ := json.Marshal(fileReadInput{FilePath: path})
	res, err := (&FileReadTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("read execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected read error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "[read_metadata]") {
		t.Fatalf("expected read metadata block, got: %s", res.Content)
	}
}

func TestFileWriteRequiresReadMetadataWhenConversationStateExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	in, _ := json.Marshal(fileWriteInput{FilePath: path, Content: "new"})
	ctx := types.ToolContext{
		WorkingDir: dir,
		Messages: []types.Message{
			types.NewTextMessage(types.RoleAssistant, "something happened"),
		},
	}
	res, err := (&FileWriteTool{}).Execute(context.Background(), in, ctx)
	if err != nil {
		t.Fatalf("write execute error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected read-before-write guardrail to fail")
	}
	if !strings.Contains(res.Content, "has not been read yet") {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "[stale_read]") || !strings.Contains(res.Content, "reason: not_read") {
		t.Fatalf("expected stale_read reason metadata, got: %s", res.Content)
	}
}

func TestFileWriteRejectsPartialReadMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	meta := renderReadMetadata(path, time.Now(), 3, true, 50, 20)
	in, _ := json.Marshal(fileWriteInput{FilePath: path, Content: "new"})
	ctx := types.ToolContext{
		WorkingDir: dir,
		Messages: []types.Message{
			{Role: types.RoleUser, Content: []types.ContentBlock{{Type: types.ContentToolResult, Content: meta}}},
		},
	}
	res, err := (&FileWriteTool{}).Execute(context.Background(), in, ctx)
	if err != nil {
		t.Fatalf("write execute error: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "only partially read") {
		t.Fatalf("expected partial-read guardrail, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, "reason: partial_read") {
		t.Fatalf("expected partial_read reason metadata, got: %s", res.Content)
	}
}

func TestFileEditRejectsStaleReadMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, oldTime, oldTime); err != nil {
		t.Fatalf("set old mtime: %v", err)
	}
	meta := renderReadMetadata(path, oldTime, 9, false, 0, 0)

	newTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(path, newTime, newTime); err != nil {
		t.Fatalf("set new mtime: %v", err)
	}

	in, _ := json.Marshal(fileEditInput{FilePath: path, OldString: "old", NewString: "new"})
	ctx := types.ToolContext{
		WorkingDir: dir,
		Messages: []types.Message{
			{Role: types.RoleUser, Content: []types.ContentBlock{{Type: types.ContentToolResult, Content: meta}}},
		},
	}
	res, err := (&FileEditTool{}).Execute(context.Background(), in, ctx)
	if err != nil {
		t.Fatalf("edit execute error: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "modified since read") {
		t.Fatalf("expected stale-read guardrail, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, "reason: mtime_changed") {
		t.Fatalf("expected mtime_changed reason metadata, got: %s", res.Content)
	}
}

func TestValidateReadBeforeModifyRejectsMissingSizeBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	msg := "[read_metadata]\npath: " + path + "\nmtime_unix_ms: " + strconv.FormatInt(time.Now().UnixMilli(), 10) + "\noffset: 0\nlimit: all\npartial: false\n[/read_metadata]"
	err := validateReadBeforeModify([]types.Message{{Role: types.RoleUser, Content: []types.ContentBlock{{Type: types.ContentToolResult, Content: msg}}}}, path, time.Now())
	if err == nil || !strings.Contains(err.Error(), "stale or incomplete") {
		t.Fatalf("expected stale/incomplete metadata error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "reason: missing_size") {
		t.Fatalf("expected missing_size reason metadata, got: %v", err)
	}
}

func TestValidateReadBeforeModifyAllowsZeroByteFilesWithSizeMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat fixture: %v", err)
	}
	meta := renderReadMetadata(path, stat.ModTime(), 0, false, 0, 2000)
	err = validateReadBeforeModify([]types.Message{{Role: types.RoleUser, Content: []types.ContentBlock{{Type: types.ContentToolResult, Content: meta}}}}, path, stat.ModTime())
	if err != nil {
		t.Fatalf("expected zero-byte file to pass guardrail, got: %v", err)
	}
}

func TestRenderReadMetadataIncludesDeterministicParityFields(t *testing.T) {
	meta := renderReadMetadata("/tmp/a.txt", time.Unix(1_700_000_000, 0), 10, true, 2, 5)
	if !strings.Contains(meta, "media_kind: text") {
		t.Fatalf("expected media_kind field, got: %s", meta)
	}
	if !strings.Contains(meta, "mime_type: text/plain") {
		t.Fatalf("expected mime_type field, got: %s", meta)
	}
	if !strings.Contains(meta, "truncation_hint: use_offset_and_limit_for_remaining_content") {
		t.Fatalf("expected truncation_hint for partial reads, got: %s", meta)
	}
}

func TestRenderReadAttachmentMetadataIncludesDeterministicBlock(t *testing.T) {
	meta := renderReadAttachmentMetadata("/tmp/a.pdf", "pdf", "application/pdf", time.Unix(1_700_000_000, 0), 42)
	if !strings.Contains(meta, "[read_metadata]") {
		t.Fatalf("expected read_metadata block, got: %s", meta)
	}
	if !strings.Contains(meta, "media_kind: pdf") {
		t.Fatalf("expected media_kind pdf, got: %s", meta)
	}
	if !strings.Contains(meta, "mime_type: application/pdf") {
		t.Fatalf("expected pdf mime type, got: %s", meta)
	}
	if !strings.Contains(meta, "truncation_hint: none") {
		t.Fatalf("expected truncation_hint none, got: %s", meta)
	}
}
