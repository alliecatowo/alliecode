package buddy

import (
	"strings"
	"testing"
)

func TestCompanionIntroText_RendersNameAndSpecies(t *testing.T) {
	text := CompanionIntroText("Mochi", SpeciesDuck)
	if !strings.Contains(text, "small duck named Mochi") {
		t.Fatalf("intro missing species/name: %q", text)
	}
	if !strings.Contains(text, "ONE line or less") {
		t.Fatalf("intro missing instruction: %q", text)
	}
}

func TestGetCompanionIntroAttachment_SkipsWhenAlreadyAnnounced(t *testing.T) {
	companion := &Companion{CompanionSoul: CompanionSoul{Name: "Mochi"}, CompanionBones: CompanionBones{Species: SpeciesDuck}}
	messages := []IntroMessage{{Type: "attachment", Attachment: IntroAttachment{Type: "companion_intro", Name: "Mochi"}}}
	attachments := GetCompanionIntroAttachment(messages, companion, false, true)
	if len(attachments) != 0 {
		t.Fatalf("expected no new attachment, got %d", len(attachments))
	}
}

func TestGetCompanionIntroAttachment_AddsWhenMissing(t *testing.T) {
	companion := &Companion{CompanionSoul: CompanionSoul{Name: "Mochi"}, CompanionBones: CompanionBones{Species: SpeciesDuck}}
	attachments := GetCompanionIntroAttachment(nil, companion, false, true)
	if len(attachments) != 1 {
		t.Fatalf("expected one attachment, got %d", len(attachments))
	}
	if attachments[0].Type != "companion_intro" || attachments[0].Name != "Mochi" {
		t.Fatalf("unexpected attachment: %#v", attachments[0])
	}
}

func TestTranscriptHasCompanionIntro_NormalizedName(t *testing.T) {
	messages := []IntroMessage{{Type: "attachment", Attachment: IntroAttachment{Type: "companion_intro", Name: "  moCHI  "}}}
	if !TranscriptHasCompanionIntro(messages, "Mochi") {
		t.Fatalf("expected transcript intro match by normalized name")
	}
}

func TestGetCompanionIntroAttachment_OneTimePerName(t *testing.T) {
	companion := &Companion{CompanionSoul: CompanionSoul{Name: "Mochi"}, CompanionBones: CompanionBones{Species: SpeciesDuck}}
	first := GetCompanionIntroAttachment(nil, companion, false, true)
	if len(first) != 1 {
		t.Fatalf("expected initial attachment")
	}

	messages := []IntroMessage{{Type: "attachment", Attachment: first[0]}}
	second := GetCompanionIntroAttachment(messages, companion, false, true)
	if len(second) != 0 {
		t.Fatalf("expected dedupe for same companion name")
	}

	companion.Name = "Pico"
	third := GetCompanionIntroAttachment(messages, companion, false, true)
	if len(third) != 1 {
		t.Fatalf("expected intro for new companion name")
	}
}
