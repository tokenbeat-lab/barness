package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func newJevInstructionsWorld(t *testing.T, ev *evidence.Case, change func(*ai.Config)) *unaryWorld {
	t.Helper()
	return newJevStateWorld(t, ev, func(c *ai.Config) {
		c.Catalog.Version = "host-jev-1.14-instructions-candidate"
		c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext, ai.ClassifierImageLocationState, ai.ClassifierImageLocationInstructions}
		if change != nil {
			change(c)
		}
	})
}

func jevInstructionsQuestion(kind string, instructions json.RawMessage) ai.ClassifierQuestion {
	switch kind {
	case "choice":
		return ai.ChoiceQuestion{Instructions: instructions, Criteria: map[string]json.RawMessage{"yes/~": json.RawMessage(`null`), "no": json.RawMessage(`"other"`)}}
	case "score":
		return ai.ScoreQuestion{Instructions: instructions, Criteria: []json.RawMessage{json.RawMessage(`"low"`), json.RawMessage(`"high"`)}}
	default:
		return ai.BoolQuestion{Instructions: instructions}
	}
}

func jevInstructionsAnswer(kind string) string {
	switch kind {
	case "choice":
		return `{"type":"choice","choice":"yes/~","probabilities":{"yes/~":1,"no":0},"confidence":1}`
	case "score":
		return `{"type":"score","score":1,"confidence":1,"probabilities":{"0":0,"1":1},"legend":{"0":"low","1":"high"}}`
	default:
		return `{"type":"noul","noul":0.75}`
	}
}
