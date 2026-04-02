package mcp

import (
	"errors"
	"testing"
)

func TestClassifyError(t *testing.T) {
	if got := ClassifyError(errors.New("401 unauthorized")); got.Category != ErrorCategoryAuthenticationFailed {
		t.Fatalf("unexpected classify result: %+v", got)
	}
}
