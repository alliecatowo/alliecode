package mcp

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyErrorMatrix(t *testing.T) {
	cases := []struct {
		err      error
		category ErrorCategory
	}{
		{err: context.DeadlineExceeded, category: ErrorCategoryTimeout},
		{err: errors.New("401 unauthorized"), category: ErrorCategoryAuthenticationFailed},
		{err: errors.New("needs auth token"), category: ErrorCategoryNeedsAuthentication},
		{err: errors.New("too many requests"), category: ErrorCategoryRateLimited},
		{err: errors.New("service unavailable"), category: ErrorCategoryUnavailable},
	}
	for _, tc := range cases {
		got := ClassifyError(tc.err)
		if got.Category != tc.category {
			t.Fatalf("classify(%q) => %v, want %v", tc.err, got.Category, tc.category)
		}
	}
}
