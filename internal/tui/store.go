package tui

type runtimeStore struct {
	state     appState
	inputMode inputMode
}

type searchStore struct {
	mode           searchMode
	query          string
	timelineQuery  string
	quickOpenQuery string
	historyQuery   string
	matchPos       int
}

type tuiStore struct {
	runtime runtimeStore
	search  searchStore
}

func newTUIStore() tuiStore {
	return tuiStore{
		runtime: runtimeStore{state: stateIdle, inputMode: inputModeChat},
		search:  searchStore{mode: searchModeTimeline},
	}
}
