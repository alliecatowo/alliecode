package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskListQueryFiltersByOwnerAndLimit(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()

	taskAdapterUpdate("a", "running", "", "")
	taskAdapterSetFields("a", func(rec *taskAdapterRecord) { rec.Owner = "dev1" })
	taskAdapterUpdate("b", "running", "", "")
	taskAdapterSetFields("b", func(rec *taskAdapterRecord) { rec.Owner = "dev1" })
	taskAdapterUpdate("c", "running", "", "")
	taskAdapterSetFields("c", func(rec *taskAdapterRecord) { rec.Owner = "dev2" })

	res, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{"owner":"dev1","limit":1}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task list failed: err=%v content=%q", err, res.Content)
	}
	var out struct {
		Total        int `json:"total"`
		QuerySummary struct {
			Matched   int  `json:"matched"`
			Returned  int  `json:"returned"`
			Truncated bool `json:"truncated"`
		} `json:"query_summary"`
	}
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatal(err)
	}
	if out.Total != 1 || out.QuerySummary.Matched != 2 || !out.QuerySummary.Truncated {
		t.Fatalf("unexpected query summary: %+v", out)
	}
}
