package ai

import "strings"

// UnsupportedDefinitionError identifies metadata that asks for a route or fixed
// Provider subset the current program cannot implement. It makes no claim about
// which future program version or vendor account could support the definition.
type UnsupportedDefinitionError struct{ Problem string }

func (e *UnsupportedDefinitionError) Error() string {
	return "ai: unsupported model definition: " + e.Problem
}

// ValidateDefinitions checks locally loaded definitions against complete model
// identities and implemented operation routes, then applies the same metadata
// constraints as NewClient and Hash. It permits an empty catalog. It does not
// contact vendors or grant any Binding permission. Hash retains its existing
// encoding and configuration contract for explicitly constructed Host catalogs.
func (c Catalog) ValidateDefinitions() error {
	identity := func(op Operation, provider ProviderID, api API, id, name string) error {
		if strings.TrimSpace(string(provider)) == "" || strings.TrimSpace(string(api)) == "" || strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
			return &ConfigError{Field: "Catalog", Problem: "model requires provider, API, ID and name"}
		}
		if !supportsOperationAPI(op, provider, api) {
			return &UnsupportedDefinitionError{Problem: "operation and API route is not implemented"}
		}
		return nil
	}
	for _, m := range c.Models {
		if err := identity(OperationChat, m.Provider, m.API, m.ID, m.Name); err != nil {
			return err
		}
		if m.ContextWindow <= 0 || m.MaxTokens <= 0 {
			return &ConfigError{Field: "Catalog", Problem: "chat capacity limits must be positive"}
		}
		if len(m.Input) == 0 {
			return &ConfigError{Field: "Catalog", Problem: "chat input must declare at least one modality"}
		}
		if unsupportedModality(m.Input) {
			return &UnsupportedDefinitionError{Problem: "model input modality is not implemented"}
		}
		if problem := kimiModelProblem(m); problem != "" {
			return &UnsupportedDefinitionError{Problem: problem}
		}
		if problem := xaiModelProblem(m); problem != "" {
			return &UnsupportedDefinitionError{Problem: problem}
		}
	}
	for _, m := range c.ImageModels {
		if err := identity(OperationImage, m.Provider, m.API, m.ID, m.Name); err != nil {
			return err
		}
		if unsupportedModality(m.Input) || unsupportedModality(m.Output) {
			return &UnsupportedDefinitionError{Problem: "image modality is not implemented"}
		}
		if problem := xaiImageModelProblem(m); problem != "" {
			return &UnsupportedDefinitionError{Problem: problem}
		}
	}
	for _, m := range c.ClassifierModels {
		if err := identity(OperationClassifier, m.Provider, m.API, m.ID, m.Name); err != nil {
			return err
		}
		if unsupportedModality(m.Input) {
			return &UnsupportedDefinitionError{Problem: "classifier modality is not implemented"}
		}
		for _, kind := range m.Capabilities.Kinds {
			if !classifierKindSupported(kind) {
				return &UnsupportedDefinitionError{Problem: "classifier question kind is not implemented"}
			}
		}
	}
	return c.validate()
}

func unsupportedModality(values []Modality) bool {
	for _, value := range values {
		if value != ModalityText && value != ModalityImage {
			return true
		}
	}
	return false
}

// supportsOperationAPI is the runtime's existing route predicate, also used by
// local definition acceptance. Compatible chat Providers keep their existing
// generic adapter semantics; unary protocols remain explicitly Provider-bound.
func supportsOperationAPI(op Operation, provider ProviderID, api API) bool {
	switch op {
	case OperationChat:
		_, supported := registry()[api]
		return supported
	case OperationClassifier:
		return (provider == ProviderTypeSafe && api == APITypeSafeSystemOne) || (provider == ProviderOpenAI && api == APIOpenAIDecisions)
	case OperationImage:
		return (provider == ProviderOpenAI && api == APIOpenAIImages) || (provider == ProviderGoogle && api == APIGoogleInteractions) || (provider == ProviderXAI && api == APIXAIImages)
	}
	return false
}
