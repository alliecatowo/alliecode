package state

import (
	"sort"
	"strings"
	"time"
)

type ProjectsRegistry struct {
	Projects              map[string]ProjectMetadata `json:"projects,omitempty"`
	LastSessionID         string                     `json:"last_session_id,omitempty"`
	LastOpenedProject     string                     `json:"last_opened_project,omitempty"`
	TotalProjects         int                        `json:"total_projects,omitempty"`
	TotalOpens            int                        `json:"total_opens,omitempty"`
	LastUpdatedAt         int64                      `json:"last_updated_at,omitempty"`
	MostRecentProject     string                     `json:"most_recent_project,omitempty"`
	MostRecentSessionPath string                     `json:"most_recent_session_path,omitempty"`
}

type ProjectMetadata struct {
	LastSessionID        string `json:"last_session_id,omitempty"`
	LastSessionPath      string `json:"last_session_path,omitempty"`
	LastSessionTitle     string `json:"last_session_title,omitempty"`
	LastSessionSummary   string `json:"last_session_summary,omitempty"`
	LastOpenedAt         int64  `json:"last_opened_at,omitempty"`
	FirstOpenedAt        int64  `json:"first_opened_at,omitempty"`
	OpenCount            int    `json:"open_count,omitempty"`
	DistinctSessionCount int    `json:"distinct_session_count,omitempty"`
	LastRuntimeSurface   string `json:"last_runtime_surface,omitempty"`
	LastCommandSurface   string `json:"last_command_surface,omitempty"`
	LastProvider         string `json:"last_provider,omitempty"`
	LastModel            string `json:"last_model,omitempty"`
}

type ProjectOpenRecord struct {
	ProjectPath    string
	SessionID      string
	SessionPath    string
	SessionTitle   string
	SessionSummary string
	RuntimeSurface string
	CommandSurface string
	Provider       string
	Model          string
	OpenedAt       time.Time
}

type ProjectQuery struct {
	ContainsPath    string
	ContainsTitle   string
	ContainsSummary string
	Limit           int
	SortByOpenCount bool
}

type ProjectSnapshot struct {
	Path     string          `json:"path"`
	Metadata ProjectMetadata `json:"metadata"`
}

func ReadProjectsRegistry(path string) (ProjectsRegistry, error) {
	state := ProjectsRegistry{Projects: make(map[string]ProjectMetadata)}
	if err := readJSONState(path, "projects registry", &state); err != nil {
		return ProjectsRegistry{}, err
	}
	return normalizeProjectsRegistry(state), nil
}

func WriteProjectsRegistry(path string, state ProjectsRegistry) error {
	state = normalizeProjectsRegistry(state)
	return writeJSONState(path, "projects registry", state)
}

func normalizeProjectsRegistry(state ProjectsRegistry) ProjectsRegistry {
	state.LastSessionID = strings.TrimSpace(state.LastSessionID)
	state.LastOpenedProject = strings.TrimSpace(state.LastOpenedProject)
	state.MostRecentProject = strings.TrimSpace(state.MostRecentProject)
	state.MostRecentSessionPath = strings.TrimSpace(state.MostRecentSessionPath)
	if state.TotalProjects < 0 {
		state.TotalProjects = 0
	}
	if state.TotalOpens < 0 {
		state.TotalOpens = 0
	}
	if state.LastUpdatedAt < 0 {
		state.LastUpdatedAt = 0
	}
	cleaned := make(map[string]ProjectMetadata)
	var totalOpens int
	var latestOpenedAt int64
	var latestProjectPath string
	var latestSessionPath string
	for projectPath, metadata := range state.Projects {
		projectPath = strings.TrimSpace(projectPath)
		if projectPath == "" {
			continue
		}
		metadata.LastSessionID = strings.TrimSpace(metadata.LastSessionID)
		metadata.LastSessionPath = strings.TrimSpace(metadata.LastSessionPath)
		metadata.LastSessionTitle = strings.TrimSpace(metadata.LastSessionTitle)
		metadata.LastSessionSummary = strings.TrimSpace(metadata.LastSessionSummary)
		metadata.LastRuntimeSurface = strings.TrimSpace(metadata.LastRuntimeSurface)
		metadata.LastCommandSurface = strings.TrimSpace(metadata.LastCommandSurface)
		metadata.LastProvider = strings.TrimSpace(metadata.LastProvider)
		metadata.LastModel = strings.TrimSpace(metadata.LastModel)
		if metadata.LastOpenedAt < 0 {
			metadata.LastOpenedAt = 0
		}
		if metadata.FirstOpenedAt < 0 {
			metadata.FirstOpenedAt = 0
		}
		if metadata.FirstOpenedAt > 0 && metadata.LastOpenedAt > 0 && metadata.FirstOpenedAt > metadata.LastOpenedAt {
			metadata.FirstOpenedAt = metadata.LastOpenedAt
		}
		if metadata.OpenCount < 0 {
			metadata.OpenCount = 0
		}
		if metadata.DistinctSessionCount < 0 {
			metadata.DistinctSessionCount = 0
		}
		if metadata.DistinctSessionCount > metadata.OpenCount && metadata.OpenCount > 0 {
			metadata.DistinctSessionCount = metadata.OpenCount
		}
		totalOpens += metadata.OpenCount
		if metadata.LastOpenedAt >= latestOpenedAt {
			latestOpenedAt = metadata.LastOpenedAt
			latestProjectPath = projectPath
			latestSessionPath = metadata.LastSessionPath
		}
		cleaned[projectPath] = metadata
	}
	state.Projects = cleaned
	state.TotalProjects = len(cleaned)
	if state.TotalOpens == 0 {
		state.TotalOpens = totalOpens
	}
	if state.LastUpdatedAt == 0 {
		state.LastUpdatedAt = latestOpenedAt
	}
	if state.MostRecentProject == "" {
		state.MostRecentProject = latestProjectPath
	}
	if state.MostRecentSessionPath == "" {
		state.MostRecentSessionPath = latestSessionPath
	}
	return state
}

func RecordProjectOpen(path string, record ProjectOpenRecord) (ProjectsRegistry, error) {
	registry, err := ReadProjectsRegistry(path)
	if err != nil {
		return ProjectsRegistry{}, err
	}

	projectPath := strings.TrimSpace(record.ProjectPath)
	if projectPath == "" {
		return registry, nil
	}

	openedAt := record.OpenedAt.UTC().Unix()
	if openedAt < 0 {
		openedAt = 0
	}

	meta := registry.Projects[projectPath]
	previousSession := meta.LastSessionID
	if sessionID := strings.TrimSpace(record.SessionID); sessionID != "" {
		meta.LastSessionID = sessionID
	}
	if sessionPath := strings.TrimSpace(record.SessionPath); sessionPath != "" {
		meta.LastSessionPath = sessionPath
	}
	if sessionTitle := strings.TrimSpace(record.SessionTitle); sessionTitle != "" {
		meta.LastSessionTitle = sessionTitle
	}
	if sessionSummary := strings.TrimSpace(record.SessionSummary); sessionSummary != "" {
		meta.LastSessionSummary = sessionSummary
	}
	if runtimeSurface := strings.TrimSpace(record.RuntimeSurface); runtimeSurface != "" {
		meta.LastRuntimeSurface = runtimeSurface
	}
	if commandSurface := strings.TrimSpace(record.CommandSurface); commandSurface != "" {
		meta.LastCommandSurface = commandSurface
	}
	if provider := strings.TrimSpace(record.Provider); provider != "" {
		meta.LastProvider = provider
	}
	if model := strings.TrimSpace(record.Model); model != "" {
		meta.LastModel = model
	}
	if meta.FirstOpenedAt == 0 || (openedAt > 0 && openedAt < meta.FirstOpenedAt) {
		meta.FirstOpenedAt = openedAt
	}
	meta.LastOpenedAt = openedAt
	meta.OpenCount++
	if meta.OpenCount < 0 {
		meta.OpenCount = 0
	}
	if strings.TrimSpace(meta.LastSessionID) != "" && previousSession != meta.LastSessionID {
		meta.DistinctSessionCount++
	}
	if meta.DistinctSessionCount == 0 && strings.TrimSpace(meta.LastSessionID) != "" {
		meta.DistinctSessionCount = 1
	}
	if meta.DistinctSessionCount > meta.OpenCount {
		meta.DistinctSessionCount = meta.OpenCount
	}

	registry.LastSessionID = strings.TrimSpace(record.SessionID)
	if registry.LastSessionID == "" {
		registry.LastSessionID = meta.LastSessionID
	}
	registry.LastOpenedProject = projectPath
	registry.MostRecentProject = projectPath
	registry.MostRecentSessionPath = meta.LastSessionPath
	registry.TotalProjects = len(registry.Projects)
	registry.TotalOpens++
	registry.LastUpdatedAt = openedAt
	registry.Projects[projectPath] = meta

	if err := WriteProjectsRegistry(path, registry); err != nil {
		return ProjectsRegistry{}, err
	}
	return registry, nil
}

func QueryProjects(path string, query ProjectQuery) ([]ProjectSnapshot, error) {
	registry, err := ReadProjectsRegistry(path)
	if err != nil {
		return nil, err
	}

	containsPath := strings.ToLower(strings.TrimSpace(query.ContainsPath))
	containsTitle := strings.ToLower(strings.TrimSpace(query.ContainsTitle))
	containsSummary := strings.ToLower(strings.TrimSpace(query.ContainsSummary))

	out := make([]ProjectSnapshot, 0, len(registry.Projects))
	for projectPath, metadata := range registry.Projects {
		if containsPath != "" && !strings.Contains(strings.ToLower(projectPath), containsPath) {
			continue
		}
		if containsTitle != "" && !strings.Contains(strings.ToLower(metadata.LastSessionTitle), containsTitle) {
			continue
		}
		if containsSummary != "" && !strings.Contains(strings.ToLower(metadata.LastSessionSummary), containsSummary) {
			continue
		}
		out = append(out, ProjectSnapshot{Path: projectPath, Metadata: metadata})
	}

	if query.SortByOpenCount {
		sortProjectsByOpenCount(out)
	} else {
		sortProjectsByOpenedAt(out)
	}

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out, nil
}

func sortProjectsByOpenedAt(items []ProjectSnapshot) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Metadata.LastOpenedAt > items[j].Metadata.LastOpenedAt
	})
}

func sortProjectsByOpenCount(items []ProjectSnapshot) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i].Metadata.OpenCount
		right := items[j].Metadata.OpenCount
		if left != right {
			return left > right
		}
		return items[i].Metadata.LastOpenedAt > items[j].Metadata.LastOpenedAt
	})
}
