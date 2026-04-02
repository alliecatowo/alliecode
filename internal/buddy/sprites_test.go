package buddy

import (
	"strings"
	"testing"
)

func TestSpriteFrame_IdleBlinkStep(t *testing.T) {
	frame, blink := SpriteFrame(SpeciesDuck, 8, false, false)
	if frame != 0 || !blink {
		t.Fatalf("expected blink on idle step; got frame=%d blink=%v", frame, blink)
	}
}

func TestSpriteFrame_SpeakingCyclesAllFrames(t *testing.T) {
	frame, blink := SpriteFrame(SpeciesDuck, 5, true, false)
	if blink {
		t.Fatalf("blink should be false while speaking")
	}
	if frame != 2 {
		t.Fatalf("expected frame 2, got %d", frame)
	}
}

func TestRenderSprite_HatAppliedWhenSlotBlank(t *testing.T) {
	bones := CompanionBones{Species: SpeciesDuck, Eye: EyeDot, Hat: HatCrown}
	lines := RenderSprite(bones, 0)
	if len(lines) == 0 {
		t.Fatalf("expected lines")
	}
	if !strings.Contains(lines[0], "^^^") {
		t.Fatalf("expected crown on first line, got %q", lines[0])
	}
}

func TestRenderSprite_DropsBlankHatRowWhenUnused(t *testing.T) {
	bones := CompanionBones{Species: SpeciesDuck, Eye: EyeDot, Hat: HatNone}
	lines := RenderSprite(bones, 0)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines after blank row drop, got %d", len(lines))
	}
}
