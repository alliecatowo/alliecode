package buddy

import (
	"reflect"
	"testing"
)

func TestRollForUser_DeterministicAndCached(t *testing.T) {
	userID := "user-123"
	first := RollForUser(userID)
	second := RollForUser(userID)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected identical cached roll")
	}

	third := RollForUser("user-456")
	if reflect.DeepEqual(first, third) {
		t.Fatalf("expected different user ids to roll differently")
	}
}

func TestRollWithSeed_Deterministic(t *testing.T) {
	seed := "seed-abc"
	a := RollWithSeed(seed)
	b := RollWithSeed(seed)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expected deterministic roll with seed")
	}
}

func TestRollRarity_WeightedSanity(t *testing.T) {
	rng := mulberry32(hashString("rarity-sanity"))
	counts := map[Rarity]int{}
	n := 10000
	for i := 0; i < n; i++ {
		counts[rollRarity(rng)]++
	}

	if !(counts[RarityCommon] > counts[RarityUncommon] &&
		counts[RarityUncommon] > counts[RarityRare] &&
		counts[RarityRare] > counts[RarityEpic] &&
		counts[RarityEpic] > counts[RarityLegendary]) {
		t.Fatalf("rarity distribution out of expected order: %#v", counts)
	}

	if counts[RarityLegendary] == 0 {
		t.Fatalf("expected at least one legendary in %d rolls", n)
	}
}
