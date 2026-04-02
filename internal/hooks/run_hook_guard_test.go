package hooks

import (
	"context"
	"testing"
)

func TestRunHookRejectsNullByteCommand(t *testing.T) {
	err := runHook(context.Background(), EventPreTool, "echo hi\x00", nil)
	if err == nil {
		t.Fatalf("expected invalid hook command error")
	}
}
