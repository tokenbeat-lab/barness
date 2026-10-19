package minimaxlive_test

import "testing"

// Never depend on whether the real repository already has an approved grant.
func TestFormalRenewalRequiresCanonicalApproval(t *testing.T) {
	testRenewalFixture(t, "missing")
}
