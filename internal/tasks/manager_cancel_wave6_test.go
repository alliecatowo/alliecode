package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestManagerCancelLifecycleWave6(t *testing.T) {
	mgr := NewManager(func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	})
	task, err := mgr.Start(context.Background(), "p", "m", types.ToolContext{})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	_, ok := mgr.Cancel(task.ID, "stop")
	if !ok {
		t.Fatalf("cancel should succeed")
	}
	waited, err := mgr.Wait(context.Background(), task.ID, 10*time.Millisecond)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("wait failed: %v", err)
	}
	if waited.Status != StatusCanceled {
		t.Fatalf("expected canceled status, got %q", waited.Status)
	}
}
