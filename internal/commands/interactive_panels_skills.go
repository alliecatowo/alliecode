package commands

import (
	"fmt"
	"sort"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildSkillsInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	skills := uniqueSortedStrings(append([]string(nil), state.Skills...))
	sources, origins, enabled := normalizedSkillMaps(state)
	enabledCount := 0
	disabledCount := 0
	for _, name := range skills {
		if enabled[name] {
			enabledCount++
		} else {
			disabledCount++
		}
	}
	panel := InteractivePanel{
		Command:       "skills",
		Title:         "skills panel: /skills",
		Subtitle:      fmt.Sprintf("skills=%d enabled=%d disabled=%d conflicts=%d", len(skills), enabledCount, disabledCount, state.SkillsConflictCount),
		HeaderIntents: skillsListIntents(skills, state.SkillsViewCount, sources, origins, enabled, state.SkillsConflictCount),
		Items: []InteractivePanelItem{
			{Key: "list", Section: "Overview", Label: "List skills", Detail: "Show skill inventory with source, origin, and enabled state", Status: statusWord(len(skills) > 0, "inventory", "empty"), ApplyInput: "/skills list", ApplyMode: PanelApplySubmit, PreviewIntents: skillsListIntents(skills, state.SkillsViewCount, sources, origins, enabled, state.SkillsConflictCount), Preview: previewLines(fmt.Sprintf("Skills: %d", len(skills)), fmt.Sprintf("Conflicts: %d", state.SkillsConflictCount))},
			{Key: "status", Section: "Overview", Label: "Skills status", Detail: "Show diagnostics counters and last sync source", Status: statusWord(state.SkillsConflictCount == 0, "clean", "conflicts"), ApplyInput: "/skills status", ApplyMode: PanelApplySubmit, PreviewIntents: skillsDoctorIntents(len(skills), enabledCount, disabledCount, state.SkillsConflictCount, state.SkillsViewCount, state.SkillsSyncCount, state.SkillsDoctorCount, state.SkillsLastSyncSource, "/skills sync", sources, origins), Preview: previewLines(fmt.Sprintf("Last sync source: %s", defaultDash(state.SkillsLastSyncSource)), fmt.Sprintf("Sync count: %d", state.SkillsSyncCount))},
			{Key: "doctor", Section: "Diagnostics", Label: "Skills doctor", Detail: "Inspect source drift, disabled skills, and collision diagnostics", Status: "diagnostics", ApplyInput: "/skills doctor", ApplyMode: PanelApplySubmit, PreviewIntents: skillsDoctorIntents(len(skills), enabledCount, disabledCount, state.SkillsConflictCount, state.SkillsViewCount, state.SkillsSyncCount, state.SkillsDoctorCount, state.SkillsLastSyncSource, "/skills sync", sources, origins), Preview: previewLines("/skills doctor surfaces collision and sync diagnostics.")},
			{Key: "sync", Section: "Actions", Label: "Sync skill sources", Detail: "Resolve skills from files plus plugins and refresh diagnostics", Status: "sync", ApplyInput: "/skills sync", ApplyMode: PanelApplySubmit, PreviewIntents: skillsRepairIntents("sync", true, "/skills status", len(skills)), Preview: previewLines("Rebuilds skill catalog from files and plugin commands.")},
			{Key: "repair-sync", Section: "Actions", Label: "Repair via sync", Detail: "Run deterministic repair flow in sync mode", Status: "repair", ApplyInput: "/skills repair sync", ApplyMode: PanelApplySubmit, PreviewIntents: skillsRepairIntents("sync", true, "/skills status", len(skills)), Preview: previewLines("Use when sources drift or conflicts persist.")},
			{Key: "repair-dedupe", Section: "Actions", Label: "Repair duplicates", Detail: "Deduplicate local state-only skill entries", Status: "repair", ApplyInput: "/skills repair dedupe", ApplyMode: PanelApplySubmit, PreviewIntents: skillsRepairIntents("dedupe", true, "/skills list", len(skills)), Preview: previewLines("Use for stale duplicate rows in session state.")},
		},
	}

	if len(skills) == 0 {
		panel.Items = append(panel.Items, InteractivePanelItem{Key: "add", Section: "Actions", Label: "Add a skill", Detail: "Stage command to add a skill name", Status: "manual", ApplyInput: "/skills add ", ApplyMode: PanelApplyStage, PreviewIntents: []types.RenderIntent{actionListIntent("Manual add", "Add a custom skill name to session inventory.", action("Add skill", "/skills add <name>", "Adds a manual skill row.", "manual"))}, Preview: previewLines("No skills available yet; add one manually or sync from sources.")})
	}

	names := append([]string(nil), skills...)
	sort.Strings(names)
	for _, name := range names {
		source := defaultDash(sources[name])
		origin := defaultDash(origins[name])
		stateWord := boolState(enabled[name], "enabled", "disabled")
		panel.Items = append(panel.Items,
			InteractivePanelItem{Key: name + ":inspect", Section: "Skills", Label: "Inspect " + name, Detail: fmt.Sprintf("source=%s origin=%s", source, origin), Status: stateWord, ApplyInput: "/skills list", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Skill row", "Resolved skill metadata.", detailRow("Skill", name, stateWord, "Skill name."), detailRow("State", stateWord, stateWord, "Enabled/disabled state."), detailRow("Source", source, "source", "Resolved source type."), detailRow("Origin", origin, "origin", "Winning source path."))}, Preview: previewLines("Skill: "+name, "Source: "+source, "Origin: "+origin)},
			InteractivePanelItem{Key: name + ":remove", Section: "Skills", Label: "Remove " + name, Detail: "Remove from session state inventory", Status: "remove", ApplyInput: "/skills remove " + name, ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionListIntent("Remove skill", "Remove skill from state inventory.", action("Remove", "/skills remove "+name, "Removes this skill from current state list.", "remove"))}, Preview: previewLines("This removes the skill entry from current session state.")},
		)
	}

	if state.SkillsConflictCount > 0 {
		panel.FooterIntents = append(panel.FooterIntents, actionHintsIntent("Conflict actions", hint("Resolve by syncing", "/skills sync"), hint("Review diagnostics", "/skills doctor")))
	}

	return panel
}
