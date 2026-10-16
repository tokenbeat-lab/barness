package e2e

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevBoxCatalog(t *testing.T) {
	for _, name := range []string{"supported", "jev-1.13.0", "jev-preview", "jev-1.14.1", "decisions", "text-only", "criteria-only", "duplicate"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC08-catalog-"+name)
			var config ai.Config
			u := newJevBoxWorld(t, ev, func(c *ai.Config) { config = *c })
			m := &config.Catalog.ClassifierModels[0]
			switch name {
			case "jev-1.13.0", "jev-preview", "jev-1.14.1":
				m.ID = name
			case "decisions":
				m.Provider = ai.ProviderOpenAI
				m.API = ai.APIOpenAIDecisions
				m.Capabilities.MinChoices = 2
				m.Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
				m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext}
			case "text-only":
				m.Input = []ai.Modality{ai.ModalityText}
				m.Capabilities.MaxInputImages = 0
				m.Capabilities.ImageMIMETypes = nil
				m.Capabilities.ImageLocations = nil
			case "criteria-only":
				m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationCriteria}
			case "duplicate":
				m.Capabilities.Kinds = append(m.Capabilities.Kinds, ai.ClassifierQuestionBox)
			}
			ev.Record("candidate-config", config.Catalog)
			client, err := ai.NewClient(config)
			if name == "supported" {
				ev.Check("explicit fixed image candidate accepts Box", err == nil && client != nil, "got %v", err)
			} else {
				ev.Check("contradictory or unconfirmed Box capabilities rejected", errors.Is(err, ai.ErrInvalidConfig) && client == nil, "got %v", err)
			}
			ev.Record("error", errString(err))
			ev.Check("config validation does not read credentials", u.host.ReadsFor("never-called").Credentials == 0, "read")
			builtin := ai.BuiltinCatalog()
			for _, model := range builtin.ClassifierModels {
				for _, kind := range model.Capabilities.Kinds {
					ev.Check("Box only belongs to independently accepted fixed Jev", kind != ai.ClassifierQuestionBox || (model.ID == "jev-1.14.0" && model.Provider == ai.ProviderTypeSafe && model.API == ai.APITypeSafeSystemOne), "promoted Box to another model")
				}
			}
		})
	}
}

func TestJevBoxClosedJSON(t *testing.T) {
	for i, question := range []string{`{"type":"box","instructions":"find","criteria":null}`, `{"type":"box","instructions":"find","image_id":"a"}`, `{"type":"box","Instructions":"find"}`, `{"type":"box","instructions":"find","instructions":"other"}`} {
		ev := run.Case(t, "UC08-closed-json-"+strconv.Itoa(i))
		var req ai.ClassifierRequest
		err := json.Unmarshal([]byte(`{"state":"photo","questions":{"q":`+question+`}}`), &req)
		ev.Check("public Box rejects fields outside its sealed contract", err != nil, "accepted %s", question)
		ev.Record("input-json", question)
		ev.Record("error", errString(err))
	}
}
