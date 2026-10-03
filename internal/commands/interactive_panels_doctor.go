package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildDoctorInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	status, sections := doctorChecks(state)
	return InteractivePanel{
		Command:       "doctor",
		Title:         "doctor panel: /doctor",
		Subtitle:      fmt.Sprintf("status=%s sections=%d", status, len(sections)),
		HeaderIntents: doctorReportIntents(status, sections),
		Items: []InteractivePanelItem{
			{Key: "human", Section: "Overview", Label: "Doctor summary", Detail: "Render the structured human report", Status: status, ApplyInput: "/doctor", ApplyMode: PanelApplySubmit, PreviewIntents: doctorReportIntents(status, sections), Preview: previewLines(fmt.Sprintf("Overall status: %s", status), fmt.Sprintf("Sections: %d", len(sections)))},
			{Key: "fix", Section: "Diagnostics", Label: "Doctor fix plan", Detail: "Generate quick fixes and corrective loops", Status: statusWord(status == "ok", "healthy", "repair"), ApplyInput: "/doctor fix", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionListIntent("Doctor fix plan", "Generate quick fixes and corrective loops.", action("Run doctor fix", "/doctor fix", "Return a deterministic repair plan with quick-fix commands.", "repair"))}, Preview: previewLines("/doctor fix returns a deterministic repair plan with quick-fix commands.")},
			{Key: "json", Section: "Diagnostics", Label: "Doctor JSON", Detail: "Emit machine-readable diagnostic output", Status: "json", ApplyInput: "/doctor json", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionListIntent("Doctor JSON", "Emit machine-readable diagnostic output.", action("Emit JSON", "/doctor json", "Preserve the same sections and checks in JSON form.", "json"))}, Preview: previewLines("/doctor json preserves the same sections and checks in JSON form.")},
		},
	}
}
