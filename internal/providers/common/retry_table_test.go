package common

import "testing"

func TestDefaultRetryTableContainsProfiles(t *testing.T) {
	table := DefaultRetryTable()
	rate := table.Lookup(ErrorClassRateLimit)
	if !rate.Retryable || rate.MaxRetries < 1 {
		t.Fatalf("rate limit profile should retry")
	}
	auth := table.Lookup(ErrorClassAuth)
	if auth.Retryable {
		t.Fatalf("auth profile should not retry")
	}
}
