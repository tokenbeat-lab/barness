package ai

// Catalog returns an independent copy of the Client's validated catalog and
// price snapshot. Discovery reads no binding or credential and grants no
// authorization. A host must not modify a configuration during construction.
func (c *Client) Catalog() Catalog { return c.catalog.clone() }

// Lookup finds a chat model by Provider, API and ID. It returns an independent
// copy. Use a validated Client.Catalog snapshot; NewClient rejects duplicates.
func (c Catalog) Lookup(provider ProviderID, api API, id string) (Model, bool) {
	for _, m := range c.Models {
		if m.Provider == provider && m.API == api && m.ID == id {
			return m.clone(), true
		}
	}
	return Model{}, false
}

// ModelsOf lists chat models for one Provider and API in catalog order.
// Every model returned owns its nested values.
func (c Catalog) ModelsOf(provider ProviderID, api API) []Model {
	var out []Model
	for _, m := range c.Models {
		if m.Provider == provider && m.API == api {
			out = append(out, m.clone())
		}
	}
	return out
}

// LookupImage finds an image model independently of the chat/classifier
// collections and returns an independent copy of its metadata.
func (c Catalog) LookupImage(provider ProviderID, api API, id string) (ImageModel, bool) {
	for _, m := range c.ImageModels {
		if m.Provider == provider && m.API == api && m.ID == id {
			return m.clone(), true
		}
	}
	return ImageModel{}, false
}

// ImageModelsOf lists independent image models for one Provider and API.
func (c Catalog) ImageModelsOf(provider ProviderID, api API) []ImageModel {
	var out []ImageModel
	for _, m := range c.ImageModels {
		if m.Provider == provider && m.API == api {
			out = append(out, m.clone())
		}
	}
	return out
}

// LookupClassifier finds a classifier model independently of the chat/image
// collections and returns an independent copy of its metadata.
func (c Catalog) LookupClassifier(provider ProviderID, api API, id string) (ClassifierModel, bool) {
	for _, m := range c.ClassifierModels {
		if m.Provider == provider && m.API == api && m.ID == id {
			return m.clone(), true
		}
	}
	return ClassifierModel{}, false
}

// ClassifierModelsOf lists independent classifier models for a Provider/API.
func (c Catalog) ClassifierModelsOf(provider ProviderID, api API) []ClassifierModel {
	var out []ClassifierModel
	for _, m := range c.ClassifierModels {
		if m.Provider == provider && m.API == api {
			out = append(out, m.clone())
		}
	}
	return out
}

// index locates each full identity in its typed collection. Only called after
// validation: there is never an ambiguous first match in the call path.
func (c Catalog) index() map[modelKey]int {
	index := make(map[modelKey]int, len(c.Models)+len(c.ImageModels)+len(c.ClassifierModels))
	for i, m := range c.Models {
		index[modelKey{OperationChat, m.Provider, m.API, m.ID}] = i
	}
	for i, m := range c.ImageModels {
		index[modelKey{OperationImage, m.Provider, m.API, m.ID}] = i
	}
	for i, m := range c.ClassifierModels {
		index[modelKey{OperationClassifier, m.Provider, m.API, m.ID}] = i
	}
	return index
}
