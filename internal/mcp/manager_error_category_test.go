package mcp

import (
	"testing"
	"time"
)

func TestSetServerErrorStoresCategory(t *testing.T) {
	m := &Manager{lastError: map[string]string{}, lastErrorCategory: map[string]ErrorCategory{}, lastErrorAt: map[string]time.Time{}}
	m.setServerErrorLocked("alpha", ErrTransportUnsupported)
	if m.lastErrorCategory["alpha"] != ErrorCategoryTransportUnsupported {
		t.Fatalf("unexpected category: %v", m.lastErrorCategory["alpha"])
	}
}
