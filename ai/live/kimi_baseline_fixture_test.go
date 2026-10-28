//go:build live && kimi_baseline_fixture

package live

import "testing"

// This nine-call fixture entry is absent from the normal live binary. Its
// admission additionally requires controlled mode before any key or claim.
func init()                                   { regionalBaselineFixture = true }
func TestKimiBaselineControlled(t *testing.T) { driveRegional(t) }
