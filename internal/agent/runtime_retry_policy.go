package agent

import (
	"context"
	"errors"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

type continuationRecoveryAction struct {
	ShouldRetry       bool
	NextRecoveryCount int
	Canceled          bool
}

func evaluateContinuationRecovery(stopReason types.StopReason, msg types.Message, recoveries, maxRecoveries int) continuationRecoveryAction {
	if stopReason == types.StopCanceled {
		return continuationRecoveryAction{ShouldRetry: false, NextRecoveryCount: recoveries, Canceled: true}
	}
	if stopReason != types.StopMaxTokens {
		return continuationRecoveryAction{ShouldRetry: false, NextRecoveryCount: recoveries}
	}
	if recoveries >= maxRecoveries {
		return continuationRecoveryAction{ShouldRetry: false, NextRecoveryCount: recoveries}
	}
	if !shouldContinueAfterMaxTokensMessage(msg) {
		return continuationRecoveryAction{ShouldRetry: false, NextRecoveryCount: recoveries}
	}
	return continuationRecoveryAction{ShouldRetry: true, NextRecoveryCount: recoveries + 1}
}

func retryPlan(kind string, attempt, max int, from, to types.AgentTurnPhase, canceled bool) types.AgentRetryPlan {
	return types.AgentRetryPlan{
		Kind:      kind,
		Attempt:   attempt,
		Max:       max,
		FromPhase: from,
		ToPhase:   to,
		Canceled:  canceled,
	}
}

func shouldContinueAfterMaxTokensMessage(msg types.Message) bool {
	if len(msg.GetToolUses()) > 0 {
		return false
	}
	text := strings.TrimSpace(msg.GetText())
	if len(text) < 80 {
		return false
	}
	last := text[len(text)-1]
	return last != '.' && last != '!' && last != '?'
}

type toolRetryPolicy struct{}

func defaultToolRetryPolicy() toolRetryPolicy {
	return toolRetryPolicy{}
}

func (toolRetryPolicy) ShouldRetry(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			return false
		default:
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "temporary") || strings.Contains(msg, "try again")
}
