package types

// AgentCompactionView carries compacted transcript view metadata.
type AgentCompactionView struct {
	Forced      bool `json:"forced,omitempty"`
	Applied     bool `json:"applied,omitempty"`
	BeforeCount int  `json:"before_count,omitempty"`
	AfterCount  int  `json:"after_count,omitempty"`
}
