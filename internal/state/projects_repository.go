package state

import "sort"

type ProjectsRepository struct {
	path string
}

func NewProjectsRepository(path string) *ProjectsRepository {
	return &ProjectsRepository{path: path}
}

func (r *ProjectsRepository) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}

func (r *ProjectsRepository) Get() (ProjectsRegistry, error) {
	return ReadProjectsRegistry(r.path)
}

func (r *ProjectsRepository) Save(registry ProjectsRegistry) error {
	return WriteProjectsRegistry(r.path, registry)
}

func (r *ProjectsRepository) RecordOpen(record ProjectOpenRecord) (ProjectsRegistry, error) {
	return RecordProjectOpen(r.path, record)
}

func (r *ProjectsRepository) Query(query ProjectQuery) ([]ProjectSnapshot, error) {
	return QueryProjects(r.path, query)
}

type ProjectIndexQuery struct {
	PathContains      string
	SessionContains   string
	RuntimeSurface    string
	CommandSurface    string
	Provider          string
	Model             string
	MinOpenCount      int
	MinDistinctVisits int
	Offset            int
	Limit             int
}

type ProjectIndexRecord struct {
	Path                 string `json:"path"`
	LastSessionID        string `json:"last_session_id,omitempty"`
	LastOpenedAtUnix     int64  `json:"last_opened_at_unix,omitempty"`
	OpenCount            int    `json:"open_count,omitempty"`
	DistinctSessionCount int    `json:"distinct_session_count,omitempty"`
	RuntimeSurface       string `json:"runtime_surface,omitempty"`
	CommandSurface       string `json:"command_surface,omitempty"`
	Provider             string `json:"provider,omitempty"`
	Model                string `json:"model,omitempty"`
}

type ProjectIndexPage struct {
	Items   []ProjectIndexRecord `json:"items,omitempty"`
	Total   int                  `json:"total"`
	Offset  int                  `json:"offset"`
	Limit   int                  `json:"limit"`
	HasMore bool                 `json:"has_more"`
}

func (r *ProjectsRepository) QueryIndex(query ProjectIndexQuery) ([]ProjectIndexRecord, error) {
	registry, err := r.Get()
	if err != nil {
		return nil, err
	}
	return BuildProjectIndex(registry, query), nil
}

func (r *ProjectsRepository) QueryIndexPage(query ProjectIndexQuery) (ProjectIndexPage, error) {
	items, err := r.QueryIndex(ProjectIndexQuery{
		PathContains:      query.PathContains,
		SessionContains:   query.SessionContains,
		RuntimeSurface:    query.RuntimeSurface,
		CommandSurface:    query.CommandSurface,
		Provider:          query.Provider,
		Model:             query.Model,
		MinOpenCount:      query.MinOpenCount,
		MinDistinctVisits: query.MinDistinctVisits,
		Limit:             0,
	})
	if err != nil {
		return ProjectIndexPage{}, err
	}
	total := len(items)
	start, end := paginateBounds(total, query.Offset, query.Limit)
	return ProjectIndexPage{Items: items[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}, nil
}

func BuildProjectIndex(registry ProjectsRegistry, query ProjectIndexQuery) []ProjectIndexRecord {
	if len(registry.Projects) == 0 {
		return nil
	}
	pathContains := normalizeContains(query.PathContains)
	sessionContains := normalizeContains(query.SessionContains)
	runtimeSurface := normalizeEquals(query.RuntimeSurface)
	commandSurface := normalizeEquals(query.CommandSurface)
	provider := normalizeEquals(query.Provider)
	model := normalizeEquals(query.Model)

	out := make([]ProjectIndexRecord, 0, len(registry.Projects))
	for path, meta := range registry.Projects {
		if !matchesContains(path, pathContains) {
			continue
		}
		if !matchesContains(meta.LastSessionID, sessionContains) {
			continue
		}
		if !matchesEquals(meta.LastRuntimeSurface, runtimeSurface) {
			continue
		}
		if !matchesEquals(meta.LastCommandSurface, commandSurface) {
			continue
		}
		if !matchesEquals(meta.LastProvider, provider) {
			continue
		}
		if !matchesEquals(meta.LastModel, model) {
			continue
		}
		if query.MinOpenCount > 0 && meta.OpenCount < query.MinOpenCount {
			continue
		}
		if query.MinDistinctVisits > 0 && meta.DistinctSessionCount < query.MinDistinctVisits {
			continue
		}
		out = append(out, ProjectIndexRecord{
			Path:                 path,
			LastSessionID:        meta.LastSessionID,
			LastOpenedAtUnix:     meta.LastOpenedAt,
			OpenCount:            meta.OpenCount,
			DistinctSessionCount: meta.DistinctSessionCount,
			RuntimeSurface:       meta.LastRuntimeSurface,
			CommandSurface:       meta.LastCommandSurface,
			Provider:             meta.LastProvider,
			Model:                meta.LastModel,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].OpenCount != out[j].OpenCount {
			return out[i].OpenCount > out[j].OpenCount
		}
		if out[i].LastOpenedAtUnix != out[j].LastOpenedAtUnix {
			return out[i].LastOpenedAtUnix > out[j].LastOpenedAtUnix
		}
		return out[i].Path < out[j].Path
	})

	start, end := paginateBounds(len(out), query.Offset, query.Limit)
	return out[start:end]
}
