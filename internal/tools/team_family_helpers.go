package tools

import (
	"sort"
	"strings"
	"time"
)

func buildTeamMembers(memberNames []string, now time.Time) []teamMember {
	members := []teamMember{{Name: "team_lead", Role: "lead", JoinedAt: now}}
	seen := map[string]struct{}{"team_lead": {}}
	for _, member := range memberNames {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		if _, ok := seen[member]; ok {
			continue
		}
		seen[member] = struct{}{}
		members = append(members, teamMember{Name: member, Role: "member", JoinedAt: now})
	}
	sort.Slice(members, func(i, j int) bool {
		if members[i].Role != members[j].Role {
			return members[i].Role < members[j].Role
		}
		return members[i].Name < members[j].Name
	})
	return members
}
