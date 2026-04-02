package buddy

type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityUncommon  Rarity = "uncommon"
	RarityRare      Rarity = "rare"
	RarityEpic      Rarity = "epic"
	RarityLegendary Rarity = "legendary"
)

var Rarities = []Rarity{
	RarityCommon,
	RarityUncommon,
	RarityRare,
	RarityEpic,
	RarityLegendary,
}

type Species string

const (
	SpeciesDuck     Species = "duck"
	SpeciesGoose    Species = "goose"
	SpeciesBlob     Species = "blob"
	SpeciesCat      Species = "cat"
	SpeciesDragon   Species = "dragon"
	SpeciesOctopus  Species = "octopus"
	SpeciesOwl      Species = "owl"
	SpeciesPenguin  Species = "penguin"
	SpeciesTurtle   Species = "turtle"
	SpeciesSnail    Species = "snail"
	SpeciesGhost    Species = "ghost"
	SpeciesAxolotl  Species = "axolotl"
	SpeciesCapybara Species = "capybara"
	SpeciesCactus   Species = "cactus"
	SpeciesRobot    Species = "robot"
	SpeciesRabbit   Species = "rabbit"
	SpeciesMushroom Species = "mushroom"
	SpeciesChonk    Species = "chonk"
)

var SpeciesList = []Species{
	SpeciesDuck,
	SpeciesGoose,
	SpeciesBlob,
	SpeciesCat,
	SpeciesDragon,
	SpeciesOctopus,
	SpeciesOwl,
	SpeciesPenguin,
	SpeciesTurtle,
	SpeciesSnail,
	SpeciesGhost,
	SpeciesAxolotl,
	SpeciesCapybara,
	SpeciesCactus,
	SpeciesRobot,
	SpeciesRabbit,
	SpeciesMushroom,
	SpeciesChonk,
}

type Eye string

const (
	EyeDot     Eye = "·"
	EyeStar    Eye = "✦"
	EyeCross   Eye = "×"
	EyeWide    Eye = "◉"
	EyeAt      Eye = "@"
	EyeTinyDot Eye = "°"
)

var Eyes = []Eye{
	EyeDot,
	EyeStar,
	EyeCross,
	EyeWide,
	EyeAt,
	EyeTinyDot,
}

type Hat string

const (
	HatNone      Hat = "none"
	HatCrown     Hat = "crown"
	HatTopHat    Hat = "tophat"
	HatPropeller Hat = "propeller"
	HatHalo      Hat = "halo"
	HatWizard    Hat = "wizard"
	HatBeanie    Hat = "beanie"
	HatTinyDuck  Hat = "tinyduck"
)

var Hats = []Hat{
	HatNone,
	HatCrown,
	HatTopHat,
	HatPropeller,
	HatHalo,
	HatWizard,
	HatBeanie,
	HatTinyDuck,
}

type StatName string

const (
	StatDebugging StatName = "DEBUGGING"
	StatPatience  StatName = "PATIENCE"
	StatChaos     StatName = "CHAOS"
	StatWisdom    StatName = "WISDOM"
	StatSnark     StatName = "SNARK"
)

var StatNames = []StatName{
	StatDebugging,
	StatPatience,
	StatChaos,
	StatWisdom,
	StatSnark,
}

type CompanionBones struct {
	Rarity  Rarity
	Species Species
	Eye     Eye
	Hat     Hat
	Shiny   bool
	Stats   map[StatName]int
}

type CompanionSoul struct {
	Name        string
	Personality string
}

type Companion struct {
	CompanionBones
	CompanionSoul
	HatchedAt int64
}

type StoredCompanion struct {
	CompanionSoul
	HatchedAt int64
}

var RarityWeights = map[Rarity]int{
	RarityCommon:    60,
	RarityUncommon:  25,
	RarityRare:      10,
	RarityEpic:      4,
	RarityLegendary: 1,
}

var RarityStars = map[Rarity]string{
	RarityCommon:    "★",
	RarityUncommon:  "★★",
	RarityRare:      "★★★",
	RarityEpic:      "★★★★",
	RarityLegendary: "★★★★★",
}

var RarityColors = map[Rarity]string{
	RarityCommon:    "inactive",
	RarityUncommon:  "success",
	RarityRare:      "permission",
	RarityEpic:      "autoAccept",
	RarityLegendary: "warning",
}

var SpeciesWeights = map[Species]int{
	SpeciesDuck:     14,
	SpeciesGoose:    9,
	SpeciesBlob:     11,
	SpeciesCat:      12,
	SpeciesDragon:   4,
	SpeciesOctopus:  6,
	SpeciesOwl:      5,
	SpeciesPenguin:  8,
	SpeciesTurtle:   7,
	SpeciesSnail:    6,
	SpeciesGhost:    5,
	SpeciesAxolotl:  5,
	SpeciesCapybara: 6,
	SpeciesCactus:   4,
	SpeciesRobot:    3,
	SpeciesRabbit:   9,
	SpeciesMushroom: 4,
	SpeciesChonk:    6,
}

var EyeWeights = map[Eye]int{
	EyeDot:     35,
	EyeStar:    8,
	EyeCross:   12,
	EyeWide:    15,
	EyeAt:      18,
	EyeTinyDot: 12,
}

var HatWeights = map[Hat]int{
	HatNone:      35,
	HatCrown:     5,
	HatTopHat:    11,
	HatPropeller: 9,
	HatHalo:      6,
	HatWizard:    8,
	HatBeanie:    17,
	HatTinyDuck:  9,
}

var PeakStatWeights = map[StatName]int{
	StatDebugging: 26,
	StatPatience:  19,
	StatChaos:     15,
	StatWisdom:    21,
	StatSnark:     19,
}

var DumpStatWeights = map[StatName]int{
	StatDebugging: 14,
	StatPatience:  18,
	StatChaos:     30,
	StatWisdom:    16,
	StatSnark:     22,
}
