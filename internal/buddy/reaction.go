package buddy

import "strings"

// AnalyzeReaction returns a deterministic companion quip when the user directly
// addresses the companion by name or sends a strong emotional cue.
func AnalyzeReaction(companion Companion, userText string) (string, bool) {
	text := strings.TrimSpace(userText)
	if text == "" {
		return "", false
	}
	low := strings.ToLower(text)
	name := strings.ToLower(strings.TrimSpace(companion.Name))
	direct := name != "" && containsWord(low, name)

	if strings.Contains(low, "/buddy pet") {
		return "purr purr", true
	}

	if direct {
		switch {
		case containsAny(low, "panic", "urgent", "asap"):
			return "breathing first, then fixes", true
		case containsAny(low, "error", "exception", "traceback", "failed", "failure", "broken", "flaky"):
			return "errors are clues", true
		case containsAny(low, "slow", "latency", "lag", "timeout", "performance"):
			return "speed pass activated", true
		case containsAny(low, "fixed", "done", "resolved", "works now", "it works"):
			return "clean recovery", true
		case containsAny(low, "ship", "deploy", "release", "merged", "lgtm"):
			return "celebration wiggle", true
		case containsAny(low, "thank", "thanks", "thx"):
			return "tiny salute", true
		case containsAny(low, "help", "stuck", "how do i", "can you"):
			return "i believe in you", true
		case containsAny(low, "sorry", "my bad", "i messed up"):
			return "all good, onward", true
		case containsAnyWord(low, "hi", "hello", "hey", "yo"):
			return "hi hi", true
		case containsAny(low, "good night", "sleep") || containsAnyWord(low, "gn"):
			return "rest well", true
		default:
			return defaultReaction(companion), true
		}
	}

	if containsAny(low, "help", "stuck", "how do i", "can someone help") {
		return "you got this", true
	}

	if containsAny(low, "error", "failed", "exception", "traceback", "broken") {
		return "debug mode paws on", true
	}

	if containsAny(low, "fixed", "resolved", "green build", "all tests pass") {
		return "nice fix", true
	}

	if containsAny(low, "fast", "faster", "optimized", "speedup") {
		return "zoom zoom", true
	}

	if containsAny(low, "lets go", "ship it", "lgtm") {
		return "lets go", true
	}

	return "", false
}

func defaultReaction(companion Companion) string {
	snark := companion.Stats[StatSnark]
	wisdom := companion.Stats[StatWisdom]
	chaos := companion.Stats[StatChaos]

	switch {
	case snark >= 70:
		return "bold move"
	case wisdom >= 70:
		return "steady and clear"
	case chaos >= 70:
		return "chaos, but cute"
	default:
		return "im listening"
	}
}

func containsAny(text string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(text, n) {
			return true
		}
	}
	return false
}

func containsAnyWord(text string, words ...string) bool {
	for _, w := range words {
		if containsWord(text, w) {
			return true
		}
	}
	return false
}

func containsWord(text string, word string) bool {
	if word == "" {
		return false
	}
	tokens := strings.FieldsFunc(text, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	for _, t := range tokens {
		if t == word {
			return true
		}
	}
	return false
}
