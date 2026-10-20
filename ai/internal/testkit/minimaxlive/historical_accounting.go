package minimaxlive

// accountingIdentity admits current targets and only the four immutable CN
// receipts already consumed before ADR-0042. The outer reader verifies each
// full receipt hash/chain first. These pins confer no call or proof authority:
// append, report, origin and predecessor selection still require current Select.
func accountingIdentity(r Receipt) bool {
	id, e := Select(r.Identity.Route, r.Identity.Model, r.Identity.Purpose)
	if e == nil && id == r.Identity {
		return true
	}
	if r.Stage != "candidate" || r.Sequence < 1 || r.Sequence > 4 {
		return false
	}
	pins := [...]string{
		"e9be5ccb57f85d6489c739574c99aa5dde9af9b7d30630fbf27610d56cf66d22",
		"f44a51b40705e17b9e7e96284476795b937dd4c099a7779b22d9c45714cdf7a3",
		"f9d37c734e0eea1251c19f39171247aa0cf5311771d2be4f61b3ac338e860e60",
		"4277d99a1d84a227e24570a24afa79a859fb5bb8ed98ba1642c33b81bc93075e",
	}
	return r.Hash == pins[r.Sequence-1]
}
