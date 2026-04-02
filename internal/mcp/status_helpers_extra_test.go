package mcp

import "testing"

func TestCloneServerStatusPreservesCategory(t *testing.T) {
	cloned := cloneServerStatus(ServerStatus{ServerName: "a", LastErrorCategory: ErrorCategoryConnection})
	if cloned.LastErrorCategory != ErrorCategoryConnection {
		t.Fatalf("unexpected cloned status: %+v", cloned)
	}
}
