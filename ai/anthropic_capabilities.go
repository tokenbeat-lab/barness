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

// Zhipu keeps only the basic Messages shape established by its own public
// Client wire captures (ADR-0035). Native Anthropic defaults stay unchanged.
func anthropicCapabilitiesOf(provider ProviderID) anthropicCapabilities {
	if provider == ProviderZhipu {
		return anthropicCapabilities{}
	}
	return anthropicCapabilities{promptCache: true, nativeThinking: true,
		eagerInputStreaming: true, browserAccess: true}
}

func (c anthropicCapabilities) managesEffort(m Model) bool {
	return c.nativeThinking && m.Compat.SupportsMidConvoEffort
}

func (c anthropicCapabilities) usesAdaptiveThinking(m Model) bool {
	return c.nativeThinking && m.Compat.ForceAdaptiveThinking
}

// The China Messages endpoint returns X-Request-Id in both successful and
// failed attempts; native Anthropic keeps Request-Id (ADR-0035).
func anthropicRequestIDHeaderFor(provider ProviderID) string {
	if provider == ProviderZhipu {
		return "x-request-id"
	}
	return anthropicRequestIDHeader
}
