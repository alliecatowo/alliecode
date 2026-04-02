package tasks

import (
	"fmt"
	"sort"
	"time"
)

func copyTeamMembers(src []TeamMember) []TeamMember {
	if len(src) == 0 {
		return nil
	}
	out := make([]TeamMember, len(src))
	copy(out, src)
	return out
}

func copyTeamMessages(src []TeamMessage) []TeamMessage {
	if len(src) == 0 {
		return nil
	}
	out := make([]TeamMessage, len(src))
	copy(out, src)
	return out
}

func copyTeamHistory(src []TeamEvent) []TeamEvent {
	if len(src) == 0 {
		return nil
	}
	out := make([]TeamEvent, len(src))
	copy(out, src)
	return out
}

func nextMessageID(seq int) string {
	return fmt.Sprintf("msg-%06d", seq)
}

func newTeamEvent(index uint64, eventType, team string, status TeamStatus, message string, at time.Time) TeamEvent {
	return TeamEvent{Index: index, Type: eventType, Team: team, Status: status, Message: message, CreatedAt: at}
}

type TeamStatusSummary struct {
	Total        int                `json:"total"`
	ByStatus     map[TeamStatus]int `json:"by_status"`
	MemberCounts map[string]int     `json:"member_counts,omitempty"`
	MessageCount int                `json:"message_count"`
	RecentTeam   string             `json:"recent_team,omitempty"`
	RecentAt     int64              `json:"recent_updated_unix,omitempty"`
}

func NewTeamStatusSummary(teams []Team) TeamStatusSummary {
	out := TeamStatusSummary{
		ByStatus: map[TeamStatus]int{
			TeamStatusActive:   0,
			TeamStatusArchived: 0,
			TeamStatusCanceled: 0,
		},
		MemberCounts: map[string]int{},
	}
	for _, team := range teams {
		out.Total++
		out.ByStatus[team.Status]++
		out.MessageCount += len(team.Messages)
		for _, member := range team.Members {
			out.MemberCounts[member.Name]++
		}
		if team.UpdatedAt.UnixNano() > out.RecentAt {
			out.RecentAt = team.UpdatedAt.UnixNano()
			out.RecentTeam = team.Name
		}
	}
	out.MemberCounts = normalizeMemberSummary(out.MemberCounts)
	return out
}

func normalizeMemberSummary(counts map[string]int) map[string]int {
	if len(counts) == 0 {
		return nil
	}
	keys := make([]string, 0, len(counts))
	for name := range counts {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	out := make(map[string]int, len(keys))
	for _, name := range keys {
		out[name] = counts[name]
	}
	return out
}
