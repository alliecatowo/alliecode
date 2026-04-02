package buddy

import "strings"

var idleSequence = []int{0, 0, 0, 0, 1, 0, 0, 0, -1, 0, 0, 2, 0, 0, 0}

var bodies = map[Species][][]string{
	SpeciesDuck: {
		{"            ", "    __      ", "  <({E} )___  ", "   (  ._>   ", "    `--`    "},
		{"            ", "    __      ", "  <({E} )___  ", "   (  ._>   ", "    `--`~   "},
		{"            ", "    __      ", "  <({E} )___  ", "   (  .__>  ", "    `--`    "},
	},
	SpeciesGoose: {
		{"            ", "     ({E}>    ", "     ||     ", "   _(__)_   ", "    ^^^^    "},
		{"            ", "    ({E}>     ", "     ||     ", "   _(__)_   ", "    ^^^^    "},
		{"            ", "     ({E}>>   ", "     ||     ", "   _(__)_   ", "    ^^^^    "},
	},
	SpeciesBlob: {
		{"            ", "   .----.   ", "  ( {E}  {E} )  ", "  (      )  ", "   `----`   "},
		{"            ", "  .------.  ", " (  {E}  {E}  ) ", " (        ) ", "  `------`  "},
		{"            ", "    .--.    ", "   ({E}  {E})   ", "   (    )   ", "    `--`    "},
	},
	SpeciesCat: {
		{"            ", "   /\\_/\\    ", "  ( {E}   {E})  ", "  (  w  )   ", "  (\")_(\")   "},
		{"            ", "   /\\_/\\    ", "  ( {E}   {E})  ", "  (  w  )   ", "  (\")_(\")~  "},
		{"            ", "   /\\-/\\    ", "  ( {E}   {E})  ", "  (  w  )   ", "  (\")_(\")   "},
	},
	SpeciesDragon: {
		{"            ", "  /^\\  /^\\  ", " <  {E}  {E}  > ", " (   ~~   ) ", "  `-vvvv-`  "},
		{"            ", "  /^\\  /^\\  ", " <  {E}  {E}  > ", " (        ) ", "  `-vvvv-`  "},
		{"   ~    ~   ", "  /^\\  /^\\  ", " <  {E}  {E}  > ", " (   ~~   ) ", "  `-vvvv-`  "},
	},
	SpeciesOctopus: {
		{"            ", "   .----.   ", "  ( {E}  {E} )  ", "  (______)  ", "  /\\/\\/\\/\\  "},
		{"            ", "   .----.   ", "  ( {E}  {E} )  ", "  (______)  ", "  \\/\\/\\/\\/  "},
		{"     o      ", "   .----.   ", "  ( {E}  {E} )  ", "  (______)  ", "  /\\/\\/\\/\\  "},
	},
	SpeciesOwl: {
		{"            ", "   /\\  /\\   ", "  (({E})({E}))  ", "  (  ><  )  ", "   `----`   "},
		{"            ", "   /\\  /\\   ", "  (({E})({E}))  ", "  (  ><  )  ", "   .----.   "},
		{"            ", "   /\\  /\\   ", "  (({E})(-))  ", "  (  ><  )  ", "   `----`   "},
	},
	SpeciesPenguin: {
		{"            ", "  .---.     ", "  ({E}>{E})     ", " /(   )\\    ", "  `---`     "},
		{"            ", "  .---.     ", "  ({E}>{E})     ", " |(   )|    ", "  `---`     "},
		{"  .---.     ", "  ({E}>{E})     ", " /(   )\\    ", "  `---`     ", "   ~ ~      "},
	},
	SpeciesTurtle: {
		{"            ", "   _,--._   ", "  ( {E}  {E} )  ", " /[______]\\ ", "  ``    ``  "},
		{"            ", "   _,--._   ", "  ( {E}  {E} )  ", " /[______]\\ ", "   ``  ``   "},
		{"            ", "   _,--._   ", "  ( {E}  {E} )  ", " /[======]\\ ", "  ``    ``  "},
	},
	SpeciesSnail: {
		{"            ", " {E}    .--.  ", "  \\  ( @ )  ", "   \\_`--`   ", "  ~~~~~~~   "},
		{"            ", "  {E}   .--.  ", "  |  ( @ )  ", "   \\_`--`   ", "  ~~~~~~~   "},
		{"            ", " {E}    .--.  ", "  \\  ( @  ) ", "   \\_`--`   ", "   ~~~~~~   "},
	},
	SpeciesGhost: {
		{"            ", "   .----.   ", "  / {E}  {E} \\  ", "  |      |  ", "  ~`~``~`~  "},
		{"            ", "   .----.   ", "  / {E}  {E} \\  ", "  |      |  ", "  `~`~~`~`  "},
		{"    ~  ~    ", "   .----.   ", "  / {E}  {E} \\  ", "  |      |  ", "  ~~`~~`~~  "},
	},
	SpeciesAxolotl: {
		{"            ", "}~(______)~{", "}~({E} .. {E})~{", "  ( .--. )  ", "  (_/  \\_)  "},
		{"            ", "~}(______){~", "~}({E} .. {E}){~", "  ( .--. )  ", "  (_/  \\_)  "},
		{"            ", "}~(______)~{", "}~({E} .. {E})~{", "  (  --  )  ", "  ~_/  \\_~  "},
	},
	SpeciesCapybara: {
		{"            ", "  n______n  ", " ( {E}    {E} ) ", " (   oo   ) ", "  `------`  "},
		{"            ", "  n______n  ", " ( {E}    {E} ) ", " (   Oo   ) ", "  `------`  "},
		{"    ~  ~    ", "  u______n  ", " ( {E}    {E} ) ", " (   oo   ) ", "  `------`  "},
	},
	SpeciesCactus: {
		{"            ", " n  ____  n ", " | |{E}  {E}| | ", " |_|    |_| ", "   |    |   "},
		{"            ", "    ____    ", " n |{E}  {E}| n ", " |_|    |_| ", "   |    |   "},
		{" n        n ", " |  ____  | ", " | |{E}  {E}| | ", " |_|    |_| ", "   |    |   "},
	},
	SpeciesRobot: {
		{"            ", "   .[||].   ", "  [ {E}  {E} ]  ", "  [ ==== ]  ", "  `------`  "},
		{"            ", "   .[||].   ", "  [ {E}  {E} ]  ", "  [ -==- ]  ", "  `------`  "},
		{"     *      ", "   .[||].   ", "  [ {E}  {E} ]  ", "  [ ==== ]  ", "  `------`  "},
	},
	SpeciesRabbit: {
		{"            ", "   (\\__/)   ", "  ( {E}  {E} )  ", " =(  ..  )= ", "  (\")__(\")  "},
		{"            ", "   (|__/)   ", "  ( {E}  {E} )  ", " =(  ..  )= ", "  (\")__(\")  "},
		{"            ", "   (\\__/)   ", "  ( {E}  {E} )  ", " =( .  . )= ", "  (\")__(\")  "},
	},
	SpeciesMushroom: {
		{"            ", " .-o-OO-o-. ", "(__________)", "   |{E}  {E}|   ", "   |____|   "},
		{"            ", " .-O-oo-O-. ", "(__________)", "   |{E}  {E}|   ", "   |____|   "},
		{"   . o  .   ", " .-o-OO-o-. ", "(__________)", "   |{E}  {E}|   ", "   |____|   "},
	},
	SpeciesChonk: {
		{"            ", "  /\\    /\\  ", " ( {E}    {E} ) ", " (   ..   ) ", "  `------`  "},
		{"            ", "  /\\    /|  ", " ( {E}    {E} ) ", " (   ..   ) ", "  `------`  "},
		{"            ", "  /\\    /\\  ", " ( {E}    {E} ) ", " (   ..   ) ", "  `------`~ "},
	},
}

var hatLines = map[Hat]string{
	HatNone:      "",
	HatCrown:     "   \\^^^/    ",
	HatTopHat:    "   [___]    ",
	HatPropeller: "    -+-     ",
	HatHalo:      "   (   )    ",
	HatWizard:    "    /^\\     ",
	HatBeanie:    "   (___)    ",
	HatTinyDuck:  "    ,>      ",
}

func RenderSprite(bones CompanionBones, frame int) []string {
	frames := bodies[bones.Species]
	body := frames[frame%len(frames)]
	lines := make([]string, 0, len(body))
	for _, line := range body {
		lines = append(lines, strings.ReplaceAll(line, "{E}", string(bones.Eye)))
	}
	if bones.Hat != HatNone && strings.TrimSpace(lines[0]) == "" {
		lines[0] = hatLines[bones.Hat]
	}
	if strings.TrimSpace(lines[0]) == "" {
		allBlank := true
		for _, f := range frames {
			if strings.TrimSpace(f[0]) != "" {
				allBlank = false
				break
			}
		}
		if allBlank {
			lines = lines[1:]
		}
	}
	return lines
}

func SpriteFrameCount(species Species) int {
	return len(bodies[species])
}

func SpriteFrame(species Species, tick int, speaking bool, petting bool) (frame int, blink bool) {
	frameCount := SpriteFrameCount(species)
	if speaking || petting {
		return tick % frameCount, false
	}
	step := idleSequence[tick%len(idleSequence)]
	if step == -1 {
		return 0, true
	}
	return step % frameCount, false
}

func RenderFace(bones CompanionBones) string {
	eye := string(bones.Eye)
	switch bones.Species {
	case SpeciesDuck, SpeciesGoose:
		return "(" + eye + ">"
	case SpeciesBlob:
		return "(" + eye + eye + ")"
	case SpeciesCat:
		return "=" + eye + "w" + eye + "="
	case SpeciesDragon:
		return "<" + eye + "~" + eye + ">"
	case SpeciesOctopus:
		return "~(" + eye + eye + ")~"
	case SpeciesOwl:
		return "(" + eye + ")(" + eye + ")"
	case SpeciesPenguin:
		return "(" + eye + ">)"
	case SpeciesTurtle:
		return "[" + eye + "_" + eye + "]"
	case SpeciesSnail:
		return eye + "(@)"
	case SpeciesGhost:
		return "/" + eye + eye + "\\"
	case SpeciesAxolotl:
		return "}" + eye + "." + eye + "{"
	case SpeciesCapybara:
		return "(" + eye + "oo" + eye + ")"
	case SpeciesCactus, SpeciesMushroom:
		return "|" + eye + "  " + eye + "|"
	case SpeciesRobot:
		return "[" + eye + eye + "]"
	case SpeciesRabbit:
		return "(" + eye + ".." + eye + ")"
	case SpeciesChonk:
		return "(" + eye + "." + eye + ")"
	default:
		return "(" + eye + ")"
	}
}
