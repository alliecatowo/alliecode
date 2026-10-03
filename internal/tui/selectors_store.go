package tui

func (a *App) syncStoreFromLegacy() {
	if a.store.runtime.state != a.state {
		a.store.runtime.state = a.state
	}
	if a.store.runtime.inputMode != a.inputMode {
		a.store.runtime.inputMode = a.inputMode
	}
	if a.store.search.mode != a.searchMode {
		a.store.search.mode = a.searchMode
	}
	if a.store.search.query != a.searchQuery {
		a.store.search.query = a.searchQuery
	}
	if a.store.search.timelineQuery != a.searchTimelineQuery {
		a.store.search.timelineQuery = a.searchTimelineQuery
	}
	if a.store.search.quickOpenQuery != a.searchQuickOpenQuery {
		a.store.search.quickOpenQuery = a.searchQuickOpenQuery
	}
	if a.store.search.historyQuery != a.searchHistoryQuery {
		a.store.search.historyQuery = a.searchHistoryQuery
	}
	if a.store.search.matchPos != a.matchPos {
		a.store.search.matchPos = a.matchPos
	}
}

func (a *App) stateValue() appState {
	a.syncStoreFromLegacy()
	return a.store.runtime.state
}

func (a *App) setStateValue(next appState) {
	a.store.runtime.state = next
	a.state = next
}

func (a *App) inputModeValue() inputMode {
	a.syncStoreFromLegacy()
	return a.store.runtime.inputMode
}

func (a *App) setInputModeValue(next inputMode) {
	a.store.runtime.inputMode = next
	a.inputMode = next
}

func (a *App) searchModeValue() searchMode {
	a.syncStoreFromLegacy()
	return a.store.search.mode
}

func (a *App) setSearchModeValue(next searchMode) {
	a.store.search.mode = next
	a.searchMode = next
}

func (a *App) searchQueryValue() string {
	a.syncStoreFromLegacy()
	return a.store.search.query
}

func (a *App) setSearchQueryValue(next string) {
	a.store.search.query = next
	a.searchQuery = next
}

func (a *App) searchMatchPosValue() int {
	a.syncStoreFromLegacy()
	return a.store.search.matchPos
}

func (a *App) setSearchMatchPosValue(next int) {
	a.store.search.matchPos = next
	a.matchPos = next
}

func (a *App) setSearchTimelineQueryValue(next string) {
	a.store.search.timelineQuery = next
	a.searchTimelineQuery = next
}

func (a *App) setSearchQuickOpenQueryValue(next string) {
	a.store.search.quickOpenQuery = next
	a.searchQuickOpenQuery = next
}

func (a *App) setSearchHistoryQueryValue(next string) {
	a.store.search.historyQuery = next
	a.searchHistoryQuery = next
}

func (a *App) timelineSearchActive() bool {
	return a.stateValue() == stateSearch && a.searchModeValue() == searchModeTimeline
}

func (a *App) quickOpenSearchActive() bool {
	return a.stateValue() == stateSearch && a.searchModeValue() == searchModeQuickOpen
}

func (a *App) historySearchActive() bool {
	return a.stateValue() == stateSearch && a.searchModeValue() == searchModeHistory
}
