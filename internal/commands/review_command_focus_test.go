package commands

import (
	"context"
	"strings"
	"testing"
)

func TestReviewFocus(t *testing.T) {
	cmd := NewReviewCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "review", Args: []string{"focus", "security"}})
	if err != nil || !strings.Contains(res.Message, "REVIEW_FOCUS") {
		t.Fatalf("review focus failed: %v %q", err, res.Message)
	}
}
