package ai

// Model-specific Messages options follow authorization and precede credential
// resolution. Existing Zhipu keeps its established adapter-phase validation;
// the new Z.AI route opts into preflight without changing that old contract.
func validateAnthropicModelOptions(m Model, options Options) *Error {
	if m.API != APIAnthropicMessages {
		return nil
	}
	if anthropicCapabilitiesOf(m.Provider).optionsBeforeCredential {
		opts, _ := options.(AnthropicOptions)
		if failure := validateGLMMessagesOptions(m, opts); failure != nil {
			failure.Phase = PhaseCapability
			return failure
		}
	}
	return validateMiniMaxMessagesOptions(m, options)
}
