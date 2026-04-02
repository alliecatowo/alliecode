package vim

import "strings"

// ParseKeys converts raw input to a deterministic key stream.
func ParseKeys(input string) []Key {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	if strings.HasPrefix(input, "<") && strings.HasSuffix(input, ">") {
		if k, ok := parseToken(strings.TrimSpace(input)); ok {
			return []Key{k}
		}
	}
	keys := make([]Key, 0, len(input))
	for _, r := range input {
		if r == 0x1b {
			keys = append(keys, Key{Special: SpecialEsc})
			continue
		}
		keys = append(keys, Key{Rune: r, Special: SpecialNone})
	}
	return keys
}

func parseToken(token string) (Key, bool) {
	norm := strings.ToLower(strings.TrimSpace(token))
	switch norm {
	case "<esc>":
		return Key{Special: SpecialEsc}, true
	case "<enter>", "<cr>":
		return Key{Special: SpecialEnter}, true
	case "<tab>":
		return Key{Special: SpecialTab}, true
	case "<bs>", "<backspace>":
		return Key{Special: SpecialBackspace}, true
	case "<del>", "<delete>":
		return Key{Special: SpecialDelete}, true
	case "<left>":
		return Key{Special: SpecialLeft}, true
	case "<right>":
		return Key{Special: SpecialRight}, true
	case "<up>":
		return Key{Special: SpecialUp}, true
	case "<down>":
		return Key{Special: SpecialDown}, true
	case "<home>":
		return Key{Special: SpecialHome}, true
	case "<end>":
		return Key{Special: SpecialEnd}, true
	default:
		return Key{}, false
	}
}
