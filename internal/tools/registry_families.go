package tools

func registerShellFamilyTools(r *Registry) {
	r.Register(&BashTool{})
	r.Register(&PowerShellTool{})
	r.Register(&REPLTool{})
}

func registerFileFamilyTools(r *Registry) {
	r.Register(&LSTool{})
	r.Register(&FileReadTool{})
	r.Register(&FileWriteTool{})
	r.Register(&FileEditTool{})
	r.Register(&GrepTool{})
	r.Register(&GlobTool{})
}

func registerWebFamilyTools(r *Registry) {
	r.Register(&WebFetchTool{})
	r.Register(&WebSearchTool{})
}

func registerTaskFamilyTools(r *Registry) {
	r.Register(&TaskListTool{})
	r.Register(&TaskGetTool{})
	r.Register(&TaskCreateTool{})
	r.Register(&TaskUpdateTool{})
	r.Register(&TaskStopTool{})
	r.Register(&TaskOutputTool{})
}

func registerTeamFamilyTools(r *Registry) {
	r.Register(&SendMessageTool{})
	r.Register(&TeamCreateTool{})
	r.Register(&TeamListTool{})
	r.Register(&TeamStatusTool{})
	r.Register(&TeamUpdateTool{})
	r.Register(&TeamDeleteTool{})
}

func registerMCPFamilyTools(r *Registry, mgr MCPManager) {
	for _, t := range mgr.Tools() {
		r.Register(t)
	}
	r.Register(&MCPResourceListTool{manager: mgr})
	r.Register(&MCPResourceReadTool{manager: mgr})
	r.Register(&MCPToolInvokeTool{manager: mgr})
	r.Register(&MCPAuthLocalTool{manager: mgr})
	r.Register(&MCPAuthStatusTool{manager: mgr})
}

func registerOrchestrationTools(r *Registry) {
	r.Register(&EnterPlanModeTool{})
	r.Register(&ExitPlanModeTool{})
	r.Register(&SyntheticOutputTool{})
	r.Register(&AgentTool{})
	r.Register(&TodoWriteTool{})
	r.Register(&NotebookEditTool{})
	r.Register(&AskUserQuestionTool{})
	r.Register(&SleepTool{})
	r.Register(&ToolSearchTool{})
	r.Register(&SkillTool{})
	r.Register(&ConfigTool{})
	r.Register(&LSPTool{})
	r.Register(&EnterWorktreeTool{})
	r.Register(&ExitWorktreeTool{})
	r.Register(&CronCreateTool{})
	r.Register(&CronDeleteTool{})
	r.Register(&CronListTool{})
	r.Register(&RemoteTriggerTool{})
	r.Register(&BriefTool{})
	r.Register(&MemoryTool{})
}
