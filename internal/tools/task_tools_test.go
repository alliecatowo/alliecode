package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskToolsWithLocalAdapter(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("task-1", "running", "", "")

	getRes, err := (&TaskGetTool{}).Execute(context.Background(), []byte(`{"task_id":"task-1"}`), types.ToolContext{})
	if err != nil || getRes.IsError {
		t.Fatalf("task_get failed: err=%v content=%q", err, getRes.Content)
	}
	var getOut map[string]any
	if err := json.Unmarshal([]byte(getRes.Content), &getOut); err != nil {
		t.Fatalf("invalid task_get JSON: %v", err)
	}
	if getOut["task"] == nil {
		t.Fatalf("expected task in task_get response")
	}

	updateRes, err := (&TaskUpdateTool{}).Execute(context.Background(), []byte(`{"task_id":"task-1","status":"completed","result":"done"}`), types.ToolContext{})
	if err != nil || updateRes.IsError {
		t.Fatalf("task_update failed: err=%v content=%q", err, updateRes.Content)
	}
	var updateOut map[string]any
	if err := json.Unmarshal([]byte(updateRes.Content), &updateOut); err != nil {
		t.Fatalf("invalid task_update JSON: %v", err)
	}
	if updateOut["task"] == nil {
		t.Fatalf("expected task object in task_update response")
	}

	outRes, err := (&TaskOutputTool{}).Execute(context.Background(), []byte(`{"task_id":"task-1","block":false}`), types.ToolContext{})
	if err != nil || outRes.IsError {
		t.Fatalf("task_output failed: err=%v content=%q", err, outRes.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(outRes.Content), &out); err != nil {
		t.Fatalf("invalid task_output JSON: %v", err)
	}
	if out["retrieval_status"] != "success" {
		t.Fatalf("unexpected retrieval_status: %v", out["retrieval_status"])
	}
	if out["output_status"] != "completed" {
		t.Fatalf("unexpected output_status: %v", out["output_status"])
	}
	if out["task"] == nil {
		t.Fatalf("expected task object in output response")
	}
	const expectedOutputPrefix = `{"retrieval_status":"success","output_status":"completed","task":{"task_id":"task-1","status":"completed"`
	if !strings.HasPrefix(outRes.Content, expectedOutputPrefix) {
		t.Fatalf("task_output should use deterministic key order, got: %s", outRes.Content)
	}

	taskAdapterUpdate("task-2", "running", "", "")
	stopRes, err := (&TaskStopTool{}).Execute(context.Background(), []byte(`{"task_id":"task-2","reason":"cancel"}`), types.ToolContext{})
	if err != nil || stopRes.IsError {
		t.Fatalf("task_stop failed: err=%v content=%q", err, stopRes.Content)
	}
	var stopOut map[string]any
	if err := json.Unmarshal([]byte(stopRes.Content), &stopOut); err != nil {
		t.Fatalf("invalid task_stop JSON: %v", err)
	}
	if stopOut["task"] == nil {
		t.Fatalf("expected task object in task_stop response")
	}
	const expectedStopPrefix = `{"task":{"task_id":"task-2","status":"canceled"`
	if !strings.HasPrefix(stopRes.Content, expectedStopPrefix) {
		t.Fatalf("task_stop should use deterministic key order, got: %s", stopRes.Content)
	}
}

func TestTaskToolsContractEdgeCases(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()

	getMissing, err := (&TaskGetTool{}).Execute(context.Background(), []byte(`{"task_id":"missing"}`), types.ToolContext{})
	if err != nil || getMissing.IsError {
		t.Fatalf("task_get missing should be non-error: err=%v content=%q", err, getMissing.Content)
	}
	var getMissingOut map[string]any
	if err := json.Unmarshal([]byte(getMissing.Content), &getMissingOut); err != nil {
		t.Fatalf("invalid task_get missing JSON: %v", err)
	}
	if v, ok := getMissingOut["task"]; !ok || v != nil {
		t.Fatalf("expected task=null for missing task, got %v", getMissingOut["task"])
	}

	taskAdapterUpdate("run-1", "running", "", "")
	badUpdate, err := (&TaskUpdateTool{}).Execute(context.Background(), []byte(`{"task_id":"run-1"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("task_update empty payload error: %v", err)
	}
	if !badUpdate.IsError {
		t.Fatalf("expected task_update to reject empty update payload")
	}

	negPoll, err := (&TaskOutputTool{}).Execute(context.Background(), []byte(`{"task_id":"run-1","poll_ms":-1}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("task_output negative poll error: %v", err)
	}
	if !negPoll.IsError {
		t.Fatalf("expected task_output to reject negative poll_ms")
	}

	nfOut, err := (&TaskOutputTool{}).Execute(context.Background(), []byte(`{"task_id":"missing","block":false}`), types.ToolContext{})
	if err != nil || nfOut.IsError {
		t.Fatalf("task_output missing should be non-error JSON: err=%v content=%q", err, nfOut.Content)
	}
	var nf map[string]any
	if err := json.Unmarshal([]byte(nfOut.Content), &nf); err != nil {
		t.Fatalf("invalid task_output missing JSON: %v", err)
	}
	if nf["retrieval_status"] != "not_found" {
		t.Fatalf("unexpected retrieval_status for missing task: %v", nf["retrieval_status"])
	}
	if nf["output_status"] != "empty" {
		t.Fatalf("unexpected output_status for missing task: %v", nf["output_status"])
	}
	if nf["runtime"] == nil {
		t.Fatalf("expected runtime payload for not_found, got: %s", nfOut.Content)
	}

	runningOut, err := (&TaskOutputTool{}).Execute(context.Background(), []byte(`{"task_id":"run-1","block":false}`), types.ToolContext{})
	if err != nil || runningOut.IsError {
		t.Fatalf("task_output running should be non-error JSON: err=%v content=%q", err, runningOut.Content)
	}
	var running map[string]any
	if err := json.Unmarshal([]byte(runningOut.Content), &running); err != nil {
		t.Fatalf("invalid task_output running JSON: %v", err)
	}
	if running["retrieval_status"] != "not_ready" {
		t.Fatalf("unexpected retrieval_status for running task: %v", running["retrieval_status"])
	}
	if running["output_status"] != "running" {
		t.Fatalf("unexpected output_status for running task: %v", running["output_status"])
	}
}

func TestTaskCreateToolCreatesTaskEnvelope(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()

	res, err := (&TaskCreateTool{}).Execute(context.Background(), []byte(`{"subject":"Write tests","description":"Add coverage"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("task_create failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("task_create returned error: %s", res.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid task_create JSON: %v", err)
	}
	taskAny, ok := out["task"]
	if !ok || taskAny == nil {
		t.Fatalf("expected task object, got %v", out["task"])
	}
	taskMap, ok := taskAny.(map[string]any)
	if !ok {
		t.Fatalf("expected task map, got %T", taskAny)
	}
	if taskMap["status"] != "pending" {
		t.Fatalf("expected pending status, got %v", taskMap["status"])
	}
	taskID, _ := taskMap["task_id"].(string)
	if taskID == "" {
		t.Fatalf("expected non-empty task_id")
	}

	if _, ok := taskAdapterGet(context.Background(), taskID); !ok {
		t.Fatalf("expected task to be persisted in adapter")
	}
}

func TestTaskCreateAndUpdateExtendedFieldsAndDelete(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()

	createRes, err := (&TaskCreateTool{}).Execute(context.Background(), []byte(`{"subject":"Ship","description":"Deploy now","activeForm":"Deploying","owner":"dev1","metadata":{"priority":"high"},"status":"in_progress"}`), types.ToolContext{})
	if err != nil || createRes.IsError {
		t.Fatalf("task_create failed: err=%v content=%q", err, createRes.Content)
	}
	var created map[string]any
	if err := json.Unmarshal([]byte(createRes.Content), &created); err != nil {
		t.Fatalf("invalid create json: %v", err)
	}
	taskMap := created["task"].(map[string]any)
	taskID := taskMap["task_id"].(string)

	getRes, err := (&TaskGetTool{}).Execute(context.Background(), []byte(`{"task_id":"`+taskID+`"}`), types.ToolContext{})
	if err != nil || getRes.IsError {
		t.Fatalf("task_get failed: err=%v content=%q", err, getRes.Content)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(getRes.Content), &got); err != nil {
		t.Fatalf("invalid get json: %v", err)
	}
	g := got["task"].(map[string]any)
	if g["status"] != "in_progress" {
		t.Fatalf("expected in_progress status, got %v", g["status"])
	}
	if g["subject"] != "Ship" || g["description"] != "Deploy now" {
		t.Fatalf("unexpected task text fields: %v", g)
	}

	updateRes, err := (&TaskUpdateTool{}).Execute(context.Background(), []byte(`{"task_id":"`+taskID+`","status":"failed","error":"boom","owner":"dev2","metadata":{"priority":"p0"}}`), types.ToolContext{})
	if err != nil || updateRes.IsError {
		t.Fatalf("task_update failed: err=%v content=%q", err, updateRes.Content)
	}

	deleteRes, err := (&TaskUpdateTool{}).Execute(context.Background(), []byte(`{"task_id":"`+taskID+`","status":"deleted"}`), types.ToolContext{})
	if err != nil || deleteRes.IsError {
		t.Fatalf("task_update delete failed: err=%v content=%q", err, deleteRes.Content)
	}
	var deleted map[string]any
	if err := json.Unmarshal([]byte(deleteRes.Content), &deleted); err != nil {
		t.Fatalf("invalid delete payload: %v", err)
	}
	if deleted["task"] != nil || deleted["summary"] == nil {
		t.Fatalf("unexpected delete payload: %s", deleteRes.Content)
	}
}

func TestTaskCreateToolRequiresSubjectAndDescription(t *testing.T) {
	res, err := (&TaskCreateTool{}).Execute(context.Background(), []byte(`{"subject":"","description":"x"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("task_create subject validation error: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "subject is required") {
		t.Fatalf("expected subject validation error, got %q", res.Content)
	}

	res, err = (&TaskCreateTool{}).Execute(context.Background(), []byte(`{"subject":"x","description":""}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("task_create description validation error: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "description is required") {
		t.Fatalf("expected description validation error, got %q", res.Content)
	}
}
