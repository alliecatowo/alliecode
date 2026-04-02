package history

import "strings"

type Repository struct {
	store Store
}

func NewRepository(store Store) *Repository {
	return &Repository{store: store}
}

func (r *Repository) Store() Store {
	if r == nil {
		return nil
	}
	return r.store
}

func (r *Repository) QueryEvents(query HistoryQuery) ([]Event, error) {
	if r == nil || r.store == nil {
		return nil, nil
	}
	return r.store.GetHistory(query)
}

func (r *Repository) QueryTimestamped(query HistoryQuery) ([]TimestampedHistoryEntry, error) {
	if r == nil || r.store == nil {
		return nil, nil
	}
	return r.store.GetTimestampedHistory(query)
}

func (r *Repository) QuerySessions(query SessionQuery) ([]SessionSummary, error) {
	if r == nil || r.store == nil {
		return nil, nil
	}
	return r.store.QueryRecentSessions(query)
}

func (r *Repository) BuildIndex(query HistoryQuery, indexQuery IndexQuery) ([]IndexRecord, error) {
	if r == nil || r.store == nil {
		return nil, nil
	}
	if strings.TrimSpace(query.Project) == "" {
		query.Project = strings.TrimSpace(indexQuery.ProjectContains)
	}
	events, err := r.store.GetHistory(query)
	if err != nil {
		return nil, err
	}
	return BuildIndex(events, indexQuery), nil
}

func (r *Repository) QuerySnapshots(query HistoryQuery, snapshotQuery SnapshotQuery) ([]Snapshot, error) {
	if r == nil || r.store == nil {
		return nil, nil
	}
	events, err := r.store.GetHistory(query)
	if err != nil {
		return nil, err
	}
	return BuildSnapshots(events, snapshotQuery), nil
}

func (r *Repository) BuildIndexSummary(query HistoryQuery, indexQuery IndexQuery) (IndexSummary, error) {
	items, err := r.BuildIndex(query, indexQuery)
	if err != nil {
		return IndexSummary{}, err
	}
	return SummarizeIndex(items), nil
}
