package tui

import (
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/buddy"
)

const (
	buddyFullSpriteMinWidth = 100
	bubbleVisibleDuration   = 10 * time.Second
	bubbleFadeWindow        = 3 * time.Second
	petBurstDuration        = 2500 * time.Millisecond
	narrowBubbleCap         = 24
)

type buddyState struct {
	companion     buddy.Companion
	bubble        string
	bubbleUntil   time.Time
	bubbleFadeAt  time.Time
	petBurstUntil time.Time
}

func newBuddyState() buddyState {
	roll := buddy.RollWithSeed("tui")
	return buddyState{companion: buddy.Companion{CompanionBones: roll.Bones, CompanionSoul: buddy.CompanionSoul{Name: "buddy"}}}
}

func renderBuddySprite(state buddyState, width int, now time.Time) string {
	petting := now.Before(state.petBurstUntil)
	speaking := strings.TrimSpace(state.bubble) != "" && now.Before(state.bubbleUntil)
	frame, blink := buddy.SpriteFrame(state.companion.Species, int(now.UnixNano()/int64(500*time.Millisecond)), speaking, petting)

	if width < buddyFullSpriteMinWidth {
		face := buddy.RenderFace(state.companion.CompanionBones)
		label := strings.TrimSpace(state.companion.Name)
		if label == "" {
			label = "buddy"
		}
		if speaking {
			reaction := truncateDisplayWidth(state.bubble, narrowBubbleCap, "...")
			return face + " \"" + reaction + "\""
		}
		if petting {
			return "<3 " + face + " " + label
		}
		return face + " " + label
	}

	lines := buddy.RenderSprite(state.companion.CompanionBones, frame)
	if blink {
		for i, line := range lines {
			lines[i] = strings.ReplaceAll(line, string(state.companion.Eye), "-")
		}
	}
	if petting {
		lines = append([]string{"  <3   <3"}, lines...)
	}
	name := strings.TrimSpace(state.companion.Name)
	if name == "" {
		name = "buddy"
	}
	lines = append(lines, "  "+name)
	return strings.Join(lines, "\n")
}

func renderBuddyBubble(state buddyState, width int, now time.Time) (string, bool) {
	if width < 24 || strings.TrimSpace(state.bubble) == "" || !now.Before(state.bubbleUntil) {
		return "", false
	}
	label := "buddy: " + strings.TrimSpace(state.bubble)
	return truncateDisplayWidth(label, width, "..."), now.After(state.bubbleFadeAt)
}

func updateBuddyState(state buddyState, entry timelineEntry, now time.Time) buddyState {
	if strings.TrimSpace(state.bubble) != "" && !now.Before(state.bubbleUntil) {
		state.bubble = ""
	}
	reaction, ok := buddyReactionFromTimeline(state.companion, entry)
	if ok {
		state.bubble = reaction
		state.bubbleUntil = now.Add(bubbleVisibleDuration)
		state.bubbleFadeAt = state.bubbleUntil.Add(-bubbleFadeWindow)
	}

	if buddyPetBurstEvent(entry) {
		state.petBurstUntil = now.Add(petBurstDuration)
	}

	if buddyMuteEvent(entry) {
		state = clearBuddyReaction(state)
	}
	return state
}

func clearBuddyReaction(state buddyState) buddyState {
	state.bubble = ""
	state.bubbleUntil = time.Time{}
	state.bubbleFadeAt = time.Time{}
	return state
}

func buddyReactionFromTimeline(companion buddy.Companion, entry timelineEntry) (string, bool) {
	if msg, ok := buddyReactionFromAssistant(entry); ok {
		return msg, true
	}
	if entry.kind == timelineUser {
		return buddy.AnalyzeReaction(companion, entry.text)
	}
	if entry.kind == timelineAssistant {
		return buddyAssistantReaction(entry.text), true
	}
	return "", false
}

func buddyAssistantReaction(text string) string {
	norm := strings.ToLower(text)
	if strings.Contains(norm, "error") || strings.Contains(norm, "cannot") || strings.Contains(norm, "can't") {
		return "oops"
	}
	return "nice"
}

func buddyReactionFromAssistant(entry timelineEntry) (string, bool) {
	if entry.kind != timelineAssistant {
		return "", false
	}
	norm := strings.ToUpper(strings.TrimSpace(entry.text))
	switch {
	case strings.HasPrefix(norm, "BUDDY_PET"):
		if strings.Contains(norm, "STATUS=PET_MUTED") {
			return "", false
		}
		return "purr purr", true
	case strings.HasPrefix(norm, "BUDDY_HATCH"):
		if strings.Contains(norm, "STATUS=HATCHED") {
			return "hi hi", true
		}
	case strings.HasPrefix(norm, "BUDDY_UNMUTE"):
		return "im listening", true
	}
	return "", false
}

func buddyPetBurstEvent(entry timelineEntry) bool {
	if entry.kind != timelineAssistant {
		return false
	}
	norm := strings.ToUpper(strings.TrimSpace(entry.text))
	return strings.HasPrefix(norm, "BUDDY_PET") && (strings.Contains(norm, "STATUS=PET") || strings.Contains(norm, "STATUS=PET_MUTED"))
}

func buddyMuteEvent(entry timelineEntry) bool {
	if entry.kind != timelineAssistant {
		return false
	}
	norm := strings.ToUpper(strings.TrimSpace(entry.text))
	return strings.HasPrefix(norm, "BUDDY_MUTE")
}
