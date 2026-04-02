package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestMemoryToolSetGetListDelete(t *testing.T) {
	dir := t.TempDir()
	tool := &MemoryTool{}
	ctx := types.ToolContext{WorkingDir: dir}

	setIn, _ := json.Marshal(memoryInput{Action: "set", Key: "foo", Value: "bar"})
	res, err := tool.Execute(context.Background(), setIn, ctx)
	if err != nil || res.IsError {
		t.Fatalf("set failed: err=%v content=%q", err, res.Content)
	}

	getIn, _ := json.Marshal(memoryInput{Action: "get", Key: "foo"})
	res, err = tool.Execute(context.Background(), getIn, ctx)
	if err != nil || res.IsError {
		t.Fatalf("get failed: err=%v content=%q", err, res.Content)
	}
	if res.Content != "bar" {
		t.Fatalf("unexpected get content: %q", res.Content)
	}

	listIn, _ := json.Marshal(memoryInput{Action: "list"})
	res, err = tool.Execute(context.Background(), listIn, ctx)
	if err != nil || res.IsError {
		t.Fatalf("list failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "foo=bar") {
		t.Fatalf("unexpected list content: %q", res.Content)
	}

	delIn, _ := json.Marshal(memoryInput{Action: "delete", Key: "foo"})
	res, err = tool.Execute(context.Background(), delIn, ctx)
	if err != nil || res.IsError {
		t.Fatalf("delete failed: err=%v content=%q", err, res.Content)
	}

	res, err = tool.Execute(context.Background(), getIn, ctx)
	if err != nil {
		t.Fatalf("get after delete returned error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected missing key to be an error")
	}

	storePath := filepath.Join(dir, memoryFileName)
	store, err := loadMemoryStore(storePath)
	if err != nil {
		t.Fatalf("failed loading store from disk: %v", err)
	}
	if len(store) != 0 {
		t.Fatalf("expected empty store, got: %#v", store)
	}
}

func TestMemoryToolReadOnlyAndDestructive(t *testing.T) {
	tool := &MemoryTool{}

	getIn, _ := json.Marshal(memoryInput{Action: "get", Key: "k"})
	if !tool.IsReadOnly(getIn) {
		t.Fatalf("get should be read-only")
	}

	listIn, _ := json.Marshal(memoryInput{Action: "list"})
	if !tool.IsReadOnly(listIn) {
		t.Fatalf("list should be read-only")
	}

	setIn, _ := json.Marshal(memoryInput{Action: "set", Key: "k", Value: "v"})
	if tool.IsReadOnly(setIn) {
		t.Fatalf("set should not be read-only")
	}

	delIn, _ := json.Marshal(memoryInput{Action: "delete", Key: "k"})
	if !tool.IsDestructive(delIn) {
		t.Fatalf("delete should be destructive")
	}
}
