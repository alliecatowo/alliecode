package tools

import (
	"sync"

	"github.com/alliecatowo/alliecode/internal/tasks"
)

type teamMember = tasks.TeamMember
type teamRecord = tasks.Team
type inboxMessage = tasks.TeamMessage

type worktreeSession struct {
	Name        string `json:"name"`
	OriginalDir string `json:"original_dir"`
	Path        string `json:"path"`
	CreatedAt   string `json:"created_at"`
}

type cronJobRecord struct {
	ID        string `json:"id"`
	Cron      string `json:"cron"`
	Prompt    string `json:"prompt"`
	Recurring bool   `json:"recurring"`
	CreatedAt string `json:"created_at"`
	NextRunAt string `json:"next_run_at,omitempty"`
}

type remoteTriggerRecord struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	RunCount  int    `json:"run_count"`
	LastRunAt string `json:"last_run_at,omitempty"`
}

type briefMessageRecord struct {
	Message         string `json:"message"`
	Status          string `json:"status"`
	AttachmentCount int    `json:"attachment_count"`
	SentAt          string `json:"sent_at"`
}

type skillUsageRecord struct {
	Name        string `json:"name"`
	Invocations int    `json:"invocations"`
	LastArgs    string `json:"last_args,omitempty"`
	LastUsedAt  string `json:"last_used_at,omitempty"`
}

var orchestrationState = struct {
	mu           sync.Mutex
	teamStore    *tasks.TeamStore
	worktree     *worktreeSession
	cronJobs     map[string]cronJobRecord
	cronSeq      int
	triggers     map[string]remoteTriggerRecord
	triggerSeq   int
	briefHistory []briefMessageRecord
	skillUsage   map[string]skillUsageRecord
}{
	teamStore:  tasks.NewTeamStore(),
	cronJobs:   map[string]cronJobRecord{},
	triggers:   map[string]remoteTriggerRecord{},
	skillUsage: map[string]skillUsageRecord{},
}

func resetOrchestrationStateForTests() {
	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	orchestrationState.teamStore.ResetForTests()
	orchestrationState.worktree = nil
	orchestrationState.cronJobs = map[string]cronJobRecord{}
	orchestrationState.cronSeq = 0
	orchestrationState.triggers = map[string]remoteTriggerRecord{}
	orchestrationState.triggerSeq = 0
	orchestrationState.briefHistory = nil
	orchestrationState.skillUsage = map[string]skillUsageRecord{}
}
