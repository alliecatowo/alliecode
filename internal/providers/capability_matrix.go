package providers

// CapabilityRequirements describes minimum capabilities required for a task.
type CapabilityRequirements struct {
	NeedsText        bool
	NeedsImage       bool
	NeedsAudio       bool
	NeedsToolUse     bool
	NeedsTools       bool
	NeedsThinking    bool
	NeedsVision      bool
	NeedsAttachments bool
	MinContextWindow int
}

// CapabilityMatrix answers whether a provider/model can satisfy requirements.
type CapabilityMatrix interface {
	Supports(provider, model string, requirements CapabilityRequirements) bool
}

// RegistryCapabilityMatrix validates capabilities using model metadata.
type RegistryCapabilityMatrix struct {
	registry ModelMetadataRegistry
}

// NewRegistryCapabilityMatrix builds a matrix backed by metadata registry.
func NewRegistryCapabilityMatrix(registry ModelMetadataRegistry) *RegistryCapabilityMatrix {
	return &RegistryCapabilityMatrix{registry: registry}
}

// Supports returns true when provider/model satisfies all requirements.
func (m *RegistryCapabilityMatrix) Supports(provider, model string, requirements CapabilityRequirements) bool {
	meta, ok := m.registry.Get(provider, model)
	if !ok {
		return false
	}
	if requirements.NeedsTools && !meta.SupportsTools {
		return false
	}
	if requirements.NeedsText && !meta.SupportsText {
		return false
	}
	if requirements.NeedsImage && !meta.SupportsImage {
		return false
	}
	if requirements.NeedsAudio && !meta.SupportsAudio {
		return false
	}
	if requirements.NeedsToolUse && !(meta.SupportsToolUse || meta.SupportsTools) {
		return false
	}
	if requirements.NeedsThinking && !meta.SupportsThinking {
		return false
	}
	if requirements.NeedsVision && !meta.SupportsVision {
		return false
	}
	if requirements.NeedsAttachments && !meta.SupportsAttachments {
		return false
	}
	if requirements.MinContextWindow > 0 && meta.ContextWindow > 0 && meta.ContextWindow < requirements.MinContextWindow {
		return false
	}
	return true
}
