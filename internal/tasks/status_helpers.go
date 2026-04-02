package tasks

import "sort"

func isTerminalStatus(status Status) bool {
	switch status {
	case StatusCompleted, StatusFailed, StatusCanceled:
		return true
	default:
		return false
	}
}

type StatusSummary struct {
	Total              int            `json:"total"`
	ByStatus           map[Status]int `json:"by_status"`
	ByOwner            map[string]int `json:"by_owner,omitempty"`
	RecentTask         string         `json:"recent_task_id,omitempty"`
	RecentAt           int64          `json:"recent_updated_unix,omitempty"`
	RecentTerminalTask string         `json:"recent_terminal_task_id,omitempty"`
	RecentTerminalAt   int64          `json:"recent_terminal_updated_unix,omitempty"`
	TerminalPct        float64        `json:"terminal_ratio"`
	Terminal           int            `json:"terminal"`
	NonTerminal        int            `json:"non_terminal"`
}

func newStatusSummary(tasks []Task) StatusSummary {
	out := StatusSummary{
		ByStatus: map[Status]int{
			StatusRunning:   0,
			StatusCompleted: 0,
			StatusFailed:    0,
			StatusCanceled:  0,
		},
		ByOwner: make(map[string]int),
	}
	for _, task := range tasks {
		out.Total++
		out.ByStatus[task.Status]++
		if task.Owner != "" {
			out.ByOwner[task.Owner]++
		}
		if isTerminalStatus(task.Status) {
			out.Terminal++
			if task.UpdatedAt.UnixNano() > out.RecentTerminalAt {
				out.RecentTerminalAt = task.UpdatedAt.UnixNano()
				out.RecentTerminalTask = task.ID
			}
		} else {
			out.NonTerminal++
		}
		if task.UpdatedAt.UnixNano() > out.RecentAt {
			out.RecentAt = task.UpdatedAt.UnixNano()
			out.RecentTask = task.ID
		}
	}
	if out.Total > 0 {
		out.TerminalPct = float64(out.Terminal) / float64(out.Total)
	}
	out.ByOwner = normalizeOwnerSummary(out.ByOwner)
	return out
}

func normalizeOwnerSummary(owners map[string]int) map[string]int {
	if len(owners) == 0 {
		return nil
	}
	keys := make([]string, 0, len(owners))
	for owner := range owners {
		keys = append(keys, owner)
	}
	sort.Strings(keys)
	out := make(map[string]int, len(keys))
	for _, owner := range keys {
		out[owner] = owners[owner]
	}
	return out
}
