package e2e

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Public catalog boundary: the accepted sourced snapshot is independent of
// BuiltinCatalog; every protocol/model owns its own mutable input slice.
func TestGLMFormalCatalog(t *testing.T) {
	ev := run.Case(t, "GLM12-formal-catalog")
	raw, err := os.ReadFile("../live/testdata/glm-candidate/candidate.json")
	var accepted ai.Catalog
	if err != nil || json.Unmarshal(raw, &accepted) != nil {
		t.Fatal("accepted sourced catalog unavailable")
	}
	cat := ai.BuiltinCatalog()
	ev.Check("new formal snapshot version", cat.Version != "2026-10-17.1" && cat.Version != accepted.Version, "version=%s", cat.Version)
	hash, err := cat.Hash()
	ev.Check("content hash", err == nil && strings.HasPrefix(hash, "sha256:"), "hash=%s err=%v", hash, err)
	count := 0
	for _, m := range cat.Models {
		if m.Provider == ai.ProviderZhipu {
			count++
		}
	}
	ev.Check("six independent identities", count == 6, "count=%d", count)
	for _, want := range accepted.Models {
		got, ok := cat.Lookup(want.Provider, want.API, want.ID)
		if !ev.Check("own identity "+string(want.API)+"/"+want.ID, ok, "missing") {
			continue
		}
		want.Name = got.Name
		ev.Check("accepted sourced facts "+string(want.API)+"/"+want.ID, reflect.DeepEqual(got, want), "model=%+v", got)
	}
	if count == 6 {
		for i := range cat.Models {
			if cat.Models[i].Provider == ai.ProviderZhipu {
				cat.Models[i].Input[0] = ai.ModalityImage
				break
			}
		}
		fresh := ai.BuiltinCatalog()
		freshHash, _ := fresh.Hash()
		ev.Check("fresh catalog owns nested data", freshHash == hash, "mutation escaped")
		for _, m := range fresh.Models {
			if m.Provider == ai.ProviderZhipu {
				ev.Check("text-only "+string(m.API)+"/"+m.ID, len(m.Input) == 1 && m.Input[0] == ai.ModalityText, "input=%v", m.Input)
			}
		}
	}
	ev.Record("catalog", cat)
}
