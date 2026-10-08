package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
)

func TestMixedPolicies(t *testing.T) {
	for name, p := range map[string]*ai.ResourcePolicy{
		"local": localassembly.MixedPolicy(), "cloud": hostintegration.CloudMixedPolicy(),
	} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "E10-mixed-policy-"+name)
			ev.Check("mixed example enables finite image and classifier policies", p.Image != nil && p.Classifier != nil, "image=%+v classifier=%+v", p.Image, p.Classifier)
		})
	}
}
