package mcp

import (
	"errors"
	"testing"
)

func TestClassifyErrorNeedsAuthWave6(t *testing.T) {
	classified := ClassifyError(errors.New("401 unauthorized"))
	if classified.Category != ErrorCategoryAuthenticationFailed {
		t.Fatalf("unexpected category: %+v", classified)
	}
}
