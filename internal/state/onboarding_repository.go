package state

import "sort"

type ProjectOnboardingRepository struct {
	path string
}

type OnboardingQuery struct {
	CompletedOnly            bool
	SkippedOnly              bool
	StepContains             string
	ShownSessionContains     string
	CompletedSessionContains string
	EvaluatedAfterUnix       int64
	Limit                    int
}

func NewProjectOnboardingRepository(path string) *ProjectOnboardingRepository {
	return &ProjectOnboardingRepository{path: path}
}

func (r *ProjectOnboardingRepository) Get() (ProjectOnboardingState, error) {
	return ReadProjectOnboardingState(r.path)
}

func (r *ProjectOnboardingRepository) Save(st ProjectOnboardingState) error {
	return WriteProjectOnboardingState(r.path, st)
}

func QueryOnboardingStates(items []ProjectOnboardingState, query OnboardingQuery) []ProjectOnboardingState {
	if len(items) == 0 {
		return nil
	}
	stepContains := normalizeContains(query.StepContains)
	shownContains := normalizeContains(query.ShownSessionContains)
	completedContains := normalizeContains(query.CompletedSessionContains)

	out := make([]ProjectOnboardingState, 0, len(items))
	for _, item := range items {
		if query.CompletedOnly && !item.HasCompleted {
			continue
		}
		if query.SkippedOnly && !item.Skipped {
			continue
		}
		if !matchesContains(item.LastStepKey, stepContains) {
			continue
		}
		if !matchesContains(item.LastShownSessionID, shownContains) {
			continue
		}
		if !matchesContains(item.LastCompletedSessionID, completedContains) {
			continue
		}
		if query.EvaluatedAfterUnix > 0 && item.LastEvaluationAtUnix < query.EvaluatedAfterUnix {
			continue
		}
		out = append(out, item)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LastEvaluationAtUnix != out[j].LastEvaluationAtUnix {
			return out[i].LastEvaluationAtUnix > out[j].LastEvaluationAtUnix
		}
		return out[i].SeenCount > out[j].SeenCount
	})

	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}
