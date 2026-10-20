package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// This success proves only the offline consumer and data proposal. Its mode,
// inactive payload and command boundary cannot authorize production activation.
func regionalControlledPromotion(t *testing.T, dirs []string) {
	ev := run.Case(t, "RP18-controlled-promotion-proposal")
	out := filepath.Join(t.TempDir(), "controlled-proposal.json")
	invoke := func(path string, inputs []string, activate bool) ([]byte, error) {
		args := []string{"run", "./ai/live/cmd/regionalpromote", "-controlled", "-proposal", path}
		if activate {
			args = []string{"run", "./ai/live/cmd/regionalpromote", "-receipt", path}
		}
		child := exec.CommandContext(ctxFor(t), "go", append(args, inputs...)...)
		child.Dir, child.Env = "../..", regionalChildEnvironment()
		return child.CombinedOutput()
	}
	before, err := os.ReadFile("../catalog_regional.json")
	if err != nil {
		t.Fatal(err)
	}
	var current struct {
		Active bool `json:"active"`
	}
	if err := json.Unmarshal(before, &current); err != nil {
		t.Fatal(err)
	}
	_, liveErr := supportmatrix.RegionalPromotionProposal(dirs)
	ev.Check("controlled originals fail the live candidate consumer", liveErr != nil && strings.Contains(liveErr.Error(), "regional own service, grant, stage or mode invalid"), "%v", liveErr)
	raw, err := invoke(out, dirs, false)
	if !ev.Check("complete own controlled proposal", err == nil, "%v %s", err, raw) {
		return
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var proposal struct {
		Kind, Mode string
		Catalog    struct {
			Active  bool
			Version string
			Models  []ai.Model
		}
		Candidates []struct {
			Combo string
			Files map[string]string
		}
	}
	if err := json.Unmarshal(data, &proposal); err != nil {
		t.Fatal(err)
	}
	ev.Check("controlled payload cannot load as production activation", proposal.Kind == "controlled-regional-catalog-proposal" && proposal.Mode == "controlled" && !proposal.Catalog.Active && len(proposal.Catalog.Models) == 9 && len(proposal.Candidates) == 6, "%s", data)
	cat := ai.Catalog{Version: proposal.Catalog.Version, Models: proposal.Catalog.Models}
	w := newWorld(t, tenantA)
	_, err = ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
	ev.Check("all proposed facts construct public client", err == nil, "%v", err)
	ev.Fixture("controlled-proposal.json", data)
	for _, test := range []struct {
		name   string
		inputs []string
		live   bool
	}{
		{"duplicate-route", append(append([]string{}, dirs[:5]...), dirs[0]), false},
		{"missing-route", dirs[:5], false},
		{"controlled-is-not-live", dirs, true},
	} {
		rejected := filepath.Join(t.TempDir(), test.name+".json")
		if test.live {
			base := filepath.Join("../..", ".evidence/provider-sites/final-acceptance-20261020")
			if err := os.MkdirAll(base, 0700); err != nil {
				t.Fatal(err)
			}
			dir, err := os.MkdirTemp(base, "controlled-refusal-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.RemoveAll(dir) })
			rejected, err = filepath.Abs(filepath.Join(dir, "lineage.json"))
			if err != nil {
				t.Fatal(err)
			}
		}
		raw, err := invoke(rejected, test.inputs, test.live)
		if test.live {
			// An already promoted catalog is rejected before candidate consumption.
			// The public live consumer above independently proves mode isolation.
			want := "regional own service, grant, stage or mode invalid"
			if current.Active {
				want = "regional inactive catalog target changed"
			}
			ev.Check("live CLI refuses at the frozen catalog boundary", strings.Contains(string(raw), want), "%s", regionalSafeOutput(raw))
		}
		ev.Check("reject "+test.name, err != nil, "%s", raw)
		_, err = os.Lstat(rejected)
		ev.Check("no rejected lineage "+test.name, os.IsNotExist(err), "file created")
	}
	raw, err = invoke(out, dirs, false)
	ev.Check("proposal receipt is exclusive", err != nil, "%s", raw)
	unchanged, err := os.ReadFile(out)
	ev.Check("existing proposal bytes retained", err == nil && string(unchanged) == string(data), "%v", err)
	after, err := os.ReadFile("../catalog_regional.json")
	ev.Check("offline proposal preserves production bytes", err == nil && string(after) == string(before), "%v", err)
}
