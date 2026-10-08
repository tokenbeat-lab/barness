package e2e

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Native descriptions can contain JSON numbers far outside float64's range.
// Neither parsing nor comparing the echoed rubric may expand their exponent.
func TestClassifierNativeLegend(t *testing.T) {
	f, _ := loadFixture(t, classifierProtocol, "mixed.json")
	sc := f.Scenarios[0]
	for _, name := range []string{"exact-large-integer", "rounded-large-integer", "equivalent-exponent", "huge-exponent", "huge-exponent-mismatch", "object-order", "array-order", "negative-zero"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E02-native-legend-"+name)
			w := scenarioWorld(t, sc)
			req := classifierInput(t, sc)
			criterion, legend := `{"id":9007199254740993}`, `{"id":9007199254740993}`
			invalid := false
			switch name {
			case "rounded-large-integer":
				legend = `{"id":9007199254740992}`
				invalid = true
			case "equivalent-exponent":
				criterion = `{"id":1.00e3}`
				legend = `{"id":1000}`
			case "huge-exponent":
				criterion = `{"id":1e999999999999999999999}`
				legend = `{"id":10e999999999999999999998}`
			case "huge-exponent-mismatch":
				criterion = `{"id":1e999999999999999999999}`
				legend = `{"id":1e999999999999999999998}`
				invalid = true
			case "object-order":
				criterion = `{"a":1,"b":[false,null,"x"]}`
				legend = `{"b":[false,null,"x"],"a":1.0}`
			case "array-order":
				criterion = `["x","y"]`
				legend = `["y","x"]`
				invalid = true
			case "negative-zero":
				criterion = `{"id":-0e100000000000000000000}`
				legend = `{"id":0}`
			}
			q := req.Questions["anger"].(ai.ScoreQuestion)
			q.Criteria[0] = json.RawMessage(criterion)
			body := strings.Replace(sc.Replies[0].Body, `"0":"Calm"`, `"0":`+legend, 1)
			enqueue(ev, w, fixtureReply{Status: 200, Body: body}.script(t, classifierProtocol, ""))
			res, err := w.client.Classify(ctxFor(t), textScope("req-legend"), sc.target(), req, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			if invalid {
				ev.Check("precise rubric mismatch rejected", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0 && res.Usage.Input == 296, "got %v", err)
			} else {
				ev.Check("equivalent native rubric accepted", err == nil && len(res.Answers) == 3, "got %v", err)
			}
		})
	}
}
