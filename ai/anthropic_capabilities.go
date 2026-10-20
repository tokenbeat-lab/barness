package ai

// anthropicCapabilities describes the automatic fields of the Messages
// endpoint, independently of a model's hard constraints. A protocol-compatible
// provider must opt into these vendor features rather than inherit them from
// a model's compat flags (ADR-0032). This is an internal profile, not a host
// capability registry; SDK version and authentication remain protocol concerns.
type anthropicCapabilities struct {
	promptCache bool
	// Adaptive thinking is independent of native display, beta headers and
	// managed per-turn controls. Compatible Providers opt in separately.
	adaptiveThinking    bool
	nativeThinking      bool
	eagerInputStreaming bool
	browserAccess       bool
}

// Zhipu keeps only the basic Messages shape established by its own public
// Client wire captures (ADR-0035). Native Anthropic defaults stay unchanged.
func anthropicCapabilitiesOf(provider ProviderID) anthropicCapabilities {
	if provider == ProviderMiniMax {
		return anthropicCapabilities{adaptiveThinking: true}
	}
	if provider == ProviderZhipu {
		return anthropicCapabilities{}
	}
	return anthropicCapabilities{promptCache: true, adaptiveThinking: true, nativeThinking: true,
		eagerInputStreaming: true, browserAccess: true}
}

func (c anthropicCapabilities) managesEffort(m Model) bool {
	return c.nativeThinking && m.Compat.SupportsMidConvoEffort
}

func (c anthropicCapabilities) usesAdaptiveThinking(m Model) bool {
	return c.adaptiveThinking && (m.Provider == ProviderMiniMax || m.Compat.ForceAdaptiveThinking)
}

// The China Messages endpoint returns X-Request-Id in both successful and
// failed attempts; native Anthropic keeps Request-Id (ADR-0035).
func anthropicRequestIDHeaderFor(provider ProviderID) string {
	if provider == ProviderZhipu {
		return "x-request-id"
	}
	return anthropicRequestIDHeader
}
