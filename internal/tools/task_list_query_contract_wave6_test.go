package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskListQuerySummaryRichMetadataWave6(t *testing.T) {
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("wave6-a", "running", "", "")
	taskAdapterSetFields("wave6-a", func(rec *taskAdapterRecord) { rec.Owner = "dev1" })
	taskAdapterUpdate("wave6-b", "completed", "ok", "")
	taskAdapterSetFields("wave6-b", func(rec *taskAdapterRecord) { rec.Owner = "dev1" })

	res, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{"owner":"dev1","include_terminal":false}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_list failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	qs, ok := out["query_summary"].(map[string]any)
	if !ok || qs["owner_filter"] != "dev1" {
		t.Fatalf("expected query summary owner filter: %+v", out)
	}
}
