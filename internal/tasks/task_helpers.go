package tasks

import (
	"fmt"
	"time"
)

func nextTaskID(now time.Time, seq uint64) string {
	return fmt.Sprintf("task-%d-%d", now.UnixNano(), seq)
}

func copyMetadata(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
