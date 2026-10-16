package e2e

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func newJevCriteriaWorld(t *testing.T, ev *evidence.Case, change func(*ai.Config)) *unaryWorld {
	return newJevInstructionsWorld(t, ev, func(c *ai.Config) {
		c.Catalog.Version = "host-jev-1.14-criteria-candidate"
		c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = append(c.Catalog.ClassifierModels[0].Capabilities.ImageLocations, ai.ClassifierImageLocationCriteria)
		if change != nil {
			change(c)
		}
	})
}

func jevCriteriaQuestion(kind string, raw json.RawMessage) (ai.ClassifierQuestion, []string, string) {
	instructions := json.RawMessage(`"judge"`)
	switch kind {
	case "choice":
		return ai.ChoiceQuestion{Instructions: instructions, Criteria: map[string]json.RawMessage{"only/~": raw}}, []string{"/only~1~0"}, `{"type":"choice","choice":"only/~","probabilities":{"only/~":1},"confidence":1}`
	case "score":
		return ai.ScoreQuestion{Instructions: instructions, Criteria: []json.RawMessage{raw, raw}}, []string{"/0", "/1"}, `{"type":"score","score":1,"confidence":1,"probabilities":{"0":0,"1":1}}`
	default:
		return ai.BoolQuestion{Instructions: instructions, Criteria: &ai.BoolCriteria{True: raw, False: raw}}, []string{"/true", "/false"}, `{"type":"noul","noul":0.75}`
	}
}

// The general fixture comparator uses float64; visual rubrics also exercise
// integers and exponents outside that range, so compare their exact tokens.
func jevCriteriaJSONEqual(a, b []byte) bool {
	decode := func(raw []byte) (any, error) {
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		var value any
		err := d.Decode(&value)
		return value, err
	}
	x, err := decode(a)
	if err != nil {
		return false
	}
	y, err := decode(b)
	return err == nil && reflect.DeepEqual(x, y)
}
