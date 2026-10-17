package ai

// anthropicCapabilities describes the automatic fields of the Messages
// endpoint, independently of a model's hard constraints. A protocol-compatible
// provider must opt into these vendor features rather than inherit them from
// a model's compat flags (ADR-0032). This is an internal profile, not a host
// capability registry; SDK version and authentication remain protocol concerns.
type anthropicCapabilities struct {
	promptCache bool
	// Native thinking includes display, adaptive/managed effort and the
	// beta headers they require; basic enabled/budget thinking is core.
	nativeThinking      bool
	eagerInputStreaming bool
	browserAccess       bool
}

// All currently served Messages routes retain the native Anthropic defaults.
// A new provider's profile belongs here alongside its own wire evidence.
func anthropicCapabilitiesOf(ProviderID) anthropicCapabilities {
	return anthropicCapabilities{promptCache: true, nativeThinking: true,
		eagerInputStreaming: true, browserAccess: true}
}

func (c anthropicCapabilities) managesEffort(m Model) bool {
	return c.nativeThinking && m.Compat.SupportsMidConvoEffort
}

func (c anthropicCapabilities) usesAdaptiveThinking(m Model) bool {
	return c.nativeThinking && m.Compat.ForceAdaptiveThinking
}
