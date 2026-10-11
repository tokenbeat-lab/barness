package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// S21/S22 prove the same three-layer isolation through actual calls, in
// addition to the directory/hash assertions: no returned map can change wire.
func TestThinkingSamplingSnapshotCalls(t *testing.T) {
	for _, p := range thinkingSamplingProtocols {
		f, raw := loadFixture(t, p.proto, "thinking-sampling.json")
		for _, sc := range f.Scenarios[:2] {
			for _, mode := range []callMode{modeStream, modeComplete} {
				t.Run(p.proto.dir+"/"+sc.entry()+"/"+string(mode), func(t *testing.T) {
					ev := run.Case(t, p.prefix+"-E03-thinking-sampling-snapshot-call-"+sc.entry()+"-"+string(mode))
					ev.Fixture("fixture.json", raw)
					var input *ai.Catalog
					patch := withModelPatch(sc.Model, nil, sc.ModelPatch)
					w := newWorldWith(t, func(c *ai.Config) { patch(c); input = c.Catalog }, tenantA)
					before := mustMarshal(t, w.client.Catalog())
					mutate := func(m *ai.Model) {
						m.SamplingParamsByThinkingLevel[ai.ThinkingHigh]["temperature"][2] = '9'
						m.SamplingParamsByThinkingLevel[ai.ThinkingHigh]["top_k"] = json.RawMessage(`999`)
						delete(m.SamplingParamsByThinkingLevel, ai.ThinkingHigh)
					}
					for i := range input.Models {
						if input.Models[i].ID == sc.Model {
							mutate(&input.Models[i])
						}
					}
					returned := w.client.Catalog()
					found, _ := returned.Lookup(p.proto.provider, p.proto.api, sc.Model)
					mutate(&found)
					list := returned.ModelsOf(p.proto.provider, p.proto.api)
					for i := range list {
						if list[i].ID == sc.Model {
							mutate(&list[i])
						}
					}
					for i := range returned.Models {
						if returned.Models[i].ID == sc.Model {
							mutate(&returned.Models[i])
						}
					}
					enqueue(ev, w, sc.replies(t)...)
					o := invokeScenario(t, w, sc, mode, "req-snapshot-call")
					o.record(ev)
					checkScenario(ev, w, sc, mode, o)
					ev.Check("directory remains owned", jsonEqual(before, mustMarshal(t, w.client.Catalog())), "directory changed")
				})
			}
		}
	}
}
