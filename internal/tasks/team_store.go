package tasks

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusArchived TeamStatus = "archived"
	TeamStatusCanceled TeamStatus = "canceled"
)

type TeamMember struct {
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type TeamMessage struct {
	ID        string    `json:"id"`
	Team      string    `json:"team"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Summary   string    `json:"summary,omitempty"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Team struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	Status       TeamStatus    `json:"status"`
	CancelReason string        `json:"cancel_reason,omitempty"`
	Members      []TeamMember  `json:"members"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Messages     []TeamMessage `json:"messages,omitempty"`
	History      []TeamEvent   `json:"history,omitempty"`
}

type TeamEvent struct {
	Index     uint64     `json:"index,omitempty"`
	Type      string     `json:"type"`
	Team      string     `json:"team,omitempty"`
	Status    TeamStatus `json:"status,omitempty"`
	Message   string     `json:"message,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type TeamUpdateParams struct {
	Description *string
	Members     []TeamMember
}

type TeamStore struct {
	mu         sync.RWMutex
	teams      map[string]*Team
	activeTeam string
	messageSeq int
	eventSeq   uint64
	onEvent    func(Team, TeamEvent)
}

type TeamStats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Archived int `json:"archived"`
	Canceled int `json:"canceled"`
}

func NewTeamStore() *TeamStore {
	return &TeamStore{teams: map[string]*Team{}}
}

func (s *TeamStore) SetEventCallback(cb func(Team, TeamEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onEvent = cb
}

func (s *TeamStore) Create(name, description string, members []TeamMember) (Team, error) {
	now := time.Now().UTC()

	s.mu.Lock()
	if _, ok := s.teams[name]; ok {
		s.mu.Unlock()
		return Team{}, fmt.Errorf("team already exists")
	}

	team := &Team{
		Name:        name,
		Description: description,
		Status:      TeamStatusActive,
		Members:     copyTeamMembers(members),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created := newTeamEvent(s.nextEventIndexLocked(), "created", name, TeamStatusActive, "", now)
	team.History = append(team.History, created)
	s.teams[name] = team
	s.activeTeam = name
	out := cloneTeam(team)
	cb := s.onEvent
	s.mu.Unlock()
	if cb != nil {
		cb(out, created)
	}
	return out, nil
}

func (s *TeamStore) Update(name string, params TeamUpdateParams) (Team, bool) {
	s.mu.Lock()
	team, ok := s.teams[name]
	if !ok {
		s.mu.Unlock()
		return Team{}, false
	}
	if params.Description != nil {
		team.Description = *params.Description
	}
	if params.Members != nil {
		team.Members = copyTeamMembers(params.Members)
	}
	team.UpdatedAt = time.Now().UTC()
	ev := newTeamEvent(s.nextEventIndexLocked(), "updated", name, team.Status, "", team.UpdatedAt)
	team.History = append(team.History, ev)
	out := cloneTeam(team)
	cb := s.onEvent
	s.mu.Unlock()
	if cb != nil {
		cb(out, ev)
	}
	return out, true
}

func (s *TeamStore) Get(name string) (Team, bool) {
	s.mu.RLock()
	team, ok := s.teams[name]
	s.mu.RUnlock()
	if !ok {
		return Team{}, false
	}
	return cloneTeam(team), true
}

func (s *TeamStore) Active() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeTeam
}

func (s *TeamStore) SetActive(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.teams[name]; !ok {
		return false
	}
	s.activeTeam = name
	return true
}

func (s *TeamStore) List() []Team {
	s.mu.RLock()
	out := make([]Team, 0, len(s.teams))
	for _, team := range s.teams {
		out = append(out, cloneTeam(team))
	}
	s.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].Name < out[j].Name
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out
}

func (s *TeamStore) Stats() TeamStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := TeamStats{Total: len(s.teams)}
	for _, team := range s.teams {
		switch team.Status {
		case TeamStatusActive:
			out.Active++
		case TeamStatusArchived:
			out.Archived++
		case TeamStatusCanceled:
			out.Canceled++
		}
	}
	return out
}

func (s *TeamStore) StatusSummary() TeamStatusSummary {
	return NewTeamStatusSummary(s.List())
}

func (s *TeamStore) StatusSummaryFor(name string) (TeamStatusSummary, bool) {
	team, ok := s.Get(name)
	if !ok {
		return TeamStatusSummary{}, false
	}
	return NewTeamStatusSummary([]Team{team}), true
}

func (s *TeamStore) Query(query TeamQuery) []Team {
	teams := s.List()
	out, _ := ApplyTeamQuery(teams, query)
	return out
}

func (s *TeamStore) QuerySummary(query TeamQuery) TeamQuerySummary {
	_, summary := ApplyTeamQuery(s.List(), query)
	return summary
}

func (s *TeamStore) Archive(name, reason string) (Team, bool) {
	if reason == "" {
		reason = "archived"
	}

	s.mu.Lock()
	team, ok := s.teams[name]
	if !ok {
		s.mu.Unlock()
		return Team{}, false
	}
	if team.Status == TeamStatusArchived {
		out := cloneTeam(team)
		s.mu.Unlock()
		return out, true
	}

	team.Status = TeamStatusArchived
	team.CancelReason = reason
	team.UpdatedAt = time.Now().UTC()
	ev := newTeamEvent(s.nextEventIndexLocked(), "archived", name, TeamStatusArchived, reason, team.UpdatedAt)
	team.History = append(team.History, ev)
	if s.activeTeam == name {
		s.activeTeam = ""
	}

	out := cloneTeam(team)
	cb := s.onEvent
	s.mu.Unlock()
	if cb != nil {
		cb(out, ev)
	}
	return out, true
}

func (s *TeamStore) Cancel(name, reason string) (Team, bool) {
	if reason == "" {
		reason = "canceled"
	}

	s.mu.Lock()
	team, ok := s.teams[name]
	if !ok {
		s.mu.Unlock()
		return Team{}, false
	}
	if team.Status == TeamStatusCanceled {
		out := cloneTeam(team)
		s.mu.Unlock()
		return out, true
	}

	team.Status = TeamStatusCanceled
	team.CancelReason = reason
	team.UpdatedAt = time.Now().UTC()
	ev := newTeamEvent(s.nextEventIndexLocked(), "canceled", name, TeamStatusCanceled, reason, team.UpdatedAt)
	team.History = append(team.History, ev)
	if s.activeTeam == name {
		s.activeTeam = ""
	}

	out := cloneTeam(team)
	cb := s.onEvent
	s.mu.Unlock()
	if cb != nil {
		cb(out, ev)
	}
	return out, true
}

func (s *TeamStore) AddMessage(teamName, from, to, summary, message string) (TeamMessage, error) {
	now := time.Now().UTC()

	s.mu.Lock()
	team, ok := s.teams[teamName]
	if !ok {
		s.mu.Unlock()
		return TeamMessage{}, fmt.Errorf("team not found")
	}

	s.messageSeq++
	rec := TeamMessage{
		ID:        nextMessageID(s.messageSeq),
		Team:      teamName,
		From:      from,
		To:        to,
		Summary:   summary,
		Message:   message,
		CreatedAt: now,
	}
	team.Messages = append(team.Messages, rec)
	team.UpdatedAt = now
	ev := newTeamEvent(s.nextEventIndexLocked(), "message", teamName, team.Status, summary, now)
	team.History = append(team.History, ev)
	out := cloneTeam(team)
	cb := s.onEvent
	s.mu.Unlock()
	if cb != nil {
		cb(out, ev)
	}
	return rec, nil
}

func (s *TeamStore) MessageCount(name string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	team, ok := s.teams[name]
	if !ok {
		return 0, false
	}
	return len(team.Messages), true
}

func (s *TeamStore) LastEvent(name string) (TeamEvent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	team, ok := s.teams[name]
	if !ok || len(team.History) == 0 {
		return TeamEvent{}, false
	}
	return team.History[len(team.History)-1], true
}

func (s *TeamStore) ResetForTests() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teams = map[string]*Team{}
	s.activeTeam = ""
	s.messageSeq = 0
	s.eventSeq = 0
}

func (s *TeamStore) nextEventIndexLocked() uint64 {
	s.eventSeq++
	return s.eventSeq
}

func cloneTeam(team *Team) Team {
	out := *team
	out.Members = copyTeamMembers(team.Members)
	out.Messages = copyTeamMessages(team.Messages)
	out.History = copyTeamHistory(team.History)
	return out
}
