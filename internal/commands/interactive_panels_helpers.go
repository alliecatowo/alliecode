package commands

import "strings"

func statusWord(ok bool, whenTrue, whenFalse string) string {
	if ok {
		return strings.TrimSpace(whenTrue)
	}
	return strings.TrimSpace(whenFalse)
}

func toggleCommand(active bool, whenActive, whenInactive string) string {
	if active {
		return whenActive
	}
	return whenInactive
}

func toggleVerb(active bool, whenActive, whenInactive string) string {
	if active {
		return whenActive
	}
	return whenInactive
}
