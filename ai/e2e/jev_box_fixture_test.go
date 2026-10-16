package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

const jevBoxAnswer = `{"type":"box","box":{"top":0.25,"bottom":0.82,"left":0.44,"right":0.84},"confidence":null}`

func newJevBoxWorld(t *testing.T, ev *evidence.Case, change func(*ai.Config)) *unaryWorld {
	return newJevCriteriaWorld(t, ev, func(c *ai.Config) {
		c.Catalog.Version = "host-jev-1.14-box-candidate"
		m := &c.Catalog.ClassifierModels[0]
		m.Capabilities.Kinds = append(m.Capabilities.Kinds, ai.ClassifierQuestionBox)
		if change != nil {
			change(c)
		}
	})
}

func jevBoxInput(t *testing.T) ai.ClassifierRequest {
	req := jevContextInput(t)
	req.Images = req.Images[:1]
	req.Questions = map[string]ai.ClassifierQuestion{"location": ai.BoxQuestion{Instructions: json.RawMessage(`"locate the mug"`)}, "exists": ai.BoolQuestion{Instructions: json.RawMessage(`"is the mug present?"`)}}
	return req
}

func enqueueJevBox(t *testing.T, ev *evidence.Case, u *unaryWorld, box, existence string) {
	enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"model":"jev-1.14.0","answers":{"location":` + box + `,"exists":{"type":"noul","noul":` + existence + `}},"usage":{"input_tokens":100,"output_tokens":20}}`}.script(t, classifierProtocol, ""))
}
