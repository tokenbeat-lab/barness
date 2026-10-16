package e2e

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

type jevZeroEntropy struct{ reads int }

func (r *jevZeroEntropy) Read(p []byte) (int, error) { r.reads++; clear(p); return len(p), nil }

// Randomness is an external seam. The public Classify call still runs the
// real adapter, while deterministic entropy makes collisions observable.
func TestJevContextCollisionBound(t *testing.T) {
	for _, name := range []string{"value", "escaped-key", "mutual"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC04-collision-"+name)
			entropy := &jevZeroEntropy{}
			saved := rand.Reader
			rand.Reader = entropy
			t.Cleanup(func() { rand.Reader = saved })
			key := "<image:" + rand.Text() + ">"
			entropy.reads = 0
			u := newJevContextWorld(t, ev, nil)
			req := jevContextInput(t)
			switch name {
			case "value":
				req.Images = req.Images[:1]
				req.State = json.RawMessage(`"before ` + key + ` after"`)
			case "escaped-key":
				req.Images = req.Images[:1]
				req.State = json.RawMessage(`{"` + strings.ReplaceAll(key, "A", `\u0041`) + `":"kept"}`)
			}
			res, err := u.client.Classify(ctxFor(t), textScope("req-jev-collision-"+name), jevContextTarget(), req, jevContextOptions())
			ev.Check("bounded collision rejects before Attempt", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseRequest}) && len(res.Metadata.Attempts) == 0, "got %v", err)
			ev.Check("finite entropy reads", entropy.reads <= 12, "unbounded selections %d", entropy.reads)
			recordJevContext(ev, u, req, res, err)
		})
	}
}
