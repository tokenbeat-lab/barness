package minimaxlive_test

import (
	"sync/atomic"
	"testing"
)

// Complete historical authorization deliberately rehashes tens of thousands
// of files. Independent public processes can overlap, but only two fixture
// owners may consume that work at once. This keeps the formal package20m bound
// without caching history or weakening any assertion (provider-sites ticket25).
var renewalOwnerSlots = make(chan struct{}, 2)
var activeRenewalOwners atomic.Int32

func admitRenewalOwner(t *testing.T) {
	t.Helper()
	renewalOwnerSlots <- struct{}{}
	active := activeRenewalOwners.Add(1)
	t.Logf("synthetic renewal owner admitted active=%d limit=2", active)
	// Call before TempDir: its later cleanup deletes all owned files before the
	// slot is returned. No holder may acquire another slot for nested subtests.
	t.Cleanup(func() {
		active := activeRenewalOwners.Add(-1)
		<-renewalOwnerSlots
		t.Logf("synthetic renewal owner released active=%d limit=2", active)
	})
}
