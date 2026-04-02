package buddy

import "strings"

type IntroAttachment struct {
	Type    string
	Name    string
	Species Species
}

type IntroMessage struct {
	Type       string
	Attachment IntroAttachment
}

func CompanionIntroText(name string, species Species) string {
	return "# Companion\n\n" +
		"A small " + string(species) + " named " + name + " sits beside the user's input box and occasionally comments in a speech bubble. You're not " + name + " - it's a separate watcher.\n\n" +
		"When the user addresses " + name + " directly (by name), its bubble will answer. Your job in that moment is to stay out of the way: respond in ONE line or less, or just answer any part of the message meant for you. Don't explain that you're not " + name + " - they know. Don't narrate what " + name + " might say - the bubble handles that."
}

func GetCompanionIntroAttachment(messages []IntroMessage, companion *Companion, muted bool, enabled bool) []IntroAttachment {
	if !enabled || muted || companion == nil {
		return nil
	}
	if TranscriptHasCompanionIntro(messages, companion.Name) {
		return nil
	}
	return []IntroAttachment{{Type: "companion_intro", Name: companion.Name, Species: companion.Species}}
}

func TranscriptHasCompanionIntro(messages []IntroMessage, name string) bool {
	needle := normalizeCompanionName(name)
	if needle == "" {
		return false
	}
	for _, msg := range messages {
		if isCompanionIntroAttachment(msg) && normalizeCompanionName(msg.Attachment.Name) == needle {
			return true
		}
	}
	return false
}

func isCompanionIntroAttachment(msg IntroMessage) bool {
	return msg.Type == "attachment" && msg.Attachment.Type == "companion_intro"
}

func normalizeCompanionName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
