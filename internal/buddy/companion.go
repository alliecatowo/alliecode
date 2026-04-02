package buddy

import (
	"hash/fnv"
	"math"
	"sync"
)

const salt = "friend-2026-401"

type Roll struct {
	Bones           CompanionBones
	InspirationSeed int
}

var rarityFloor = map[Rarity]int{
	RarityCommon:    5,
	RarityUncommon:  15,
	RarityRare:      25,
	RarityEpic:      35,
	RarityLegendary: 50,
}

var (
	rollCacheMu sync.RWMutex
	rollCache   = map[string]Roll{}
)

func hashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func mulberry32(seed uint32) func() float64 {
	a := seed
	return func() float64 {
		a += 0x6d2b79f5
		t := a
		t = uint32(int32(t^(t>>15)) * int32(1|t))
		t ^= t + uint32(int32(t^(t>>7))*int32(61|t))
		v := t ^ (t >> 14)
		return float64(v) / 4294967296.0
	}
}

func pick[T any](rng func() float64, vals []T) T {
	idx := int(math.Floor(rng() * float64(len(vals))))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(vals) {
		idx = len(vals) - 1
	}
	return vals[idx]
}

func pickWeighted[T comparable](rng func() float64, vals []T, weights map[T]int) T {
	total := 0
	for _, v := range vals {
		total += weights[v]
	}
	if total <= 0 {
		return pick(rng, vals)
	}

	roll := rng() * float64(total)
	for _, v := range vals {
		roll -= float64(weights[v])
		if roll < 0 {
			return v
		}
	}
	return vals[len(vals)-1]
}

func rollRarity(rng func() float64) Rarity {
	total := 0
	for _, r := range Rarities {
		total += RarityWeights[r]
	}
	roll := rng() * float64(total)
	for _, rarity := range Rarities {
		roll -= float64(RarityWeights[rarity])
		if roll < 0 {
			return rarity
		}
	}
	return RarityCommon
}

func rollStats(rng func() float64, rarity Rarity) map[StatName]int {
	floor := rarityFloor[rarity]
	peak := pickWeighted(rng, StatNames, PeakStatWeights)
	dump := pickWeighted(rng, StatNames, DumpStatWeights)
	for dump == peak {
		dump = pickWeighted(rng, StatNames, DumpStatWeights)
	}

	stats := make(map[StatName]int, len(StatNames))
	for _, name := range StatNames {
		if name == peak {
			stats[name] = minInt(100, floor+50+int(math.Floor(rng()*30)))
			continue
		}
		if name == dump {
			stats[name] = maxInt(1, floor-10+int(math.Floor(rng()*15)))
			continue
		}
		stats[name] = floor + int(math.Floor(rng()*40))
	}
	return stats
}

func rollFromRNG(rng func() float64) Roll {
	rarity := rollRarity(rng)
	bones := CompanionBones{
		Rarity:  rarity,
		Species: pickWeighted(rng, SpeciesList, SpeciesWeights),
		Eye:     pickWeighted(rng, Eyes, EyeWeights),
		Hat:     pickWeighted(rng, Hats, HatWeights),
		Shiny:   rng() < 0.01,
		Stats:   rollStats(rng, rarity),
	}
	if rarity == RarityCommon {
		bones.Hat = HatNone
	}
	return Roll{Bones: bones, InspirationSeed: int(math.Floor(rng() * 1e9))}
}

func RollForUser(userID string) Roll {
	key := userID + salt

	rollCacheMu.RLock()
	if cached, ok := rollCache[key]; ok {
		rollCacheMu.RUnlock()
		return cached
	}
	rollCacheMu.RUnlock()

	seed := hashString(key)
	rolled := rollFromRNG(mulberry32(seed))

	rollCacheMu.Lock()
	rollCache[key] = rolled
	rollCacheMu.Unlock()

	return rolled
}

func RollWithSeed(seed string) Roll {
	return rollFromRNG(mulberry32(hashString(seed)))
}

func BuildCompanion(userID string, stored StoredCompanion) Companion {
	bones := RollForUser(userID).Bones
	return Companion{CompanionBones: bones, CompanionSoul: stored.CompanionSoul, HatchedAt: stored.HatchedAt}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
