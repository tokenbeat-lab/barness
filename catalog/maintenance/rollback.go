package maintenance

// rollbackProvider restores prior reviewed content as a new candidate. It
// retains later definitions withdrawn, so rollback never destroys recovery
// data or rewrites the selected historical revision's version and bytes.
func rollbackProvider(store string, state *State, request ChangeRequest) error {
	if request.Identity.Provider == "" || request.Revision == "" {
		return failure("invalid_argument", "rollback requires an explicit Provider and historical revision")
	}
	history, err := History(store)
	if err != nil {
		return err
	}
	var historical *ProviderRecord
	for _, revision := range history {
		if revision.ID != request.Revision {
			continue
		}
		for _, provider := range revision.State.Providers {
			if provider.ID == request.Identity.Provider {
				historical = &provider
				break
			}
		}
		break
	}
	if historical == nil {
		return failure("not_found", "selected Provider revision is not in committed history")
	}
	for pi := range state.Providers {
		current := &state.Providers[pi]
		if current.ID != request.Identity.Provider {
			continue
		}
		seen := map[Identity]bool{}
		for _, entry := range historical.Entries {
			seen[entryIdentity(entry)] = true
		}
		entries := append([]Entry(nil), historical.Entries...)
		for _, entry := range current.Entries {
			if !seen[entryIdentity(entry)] {
				entry.Lifecycle = Withdrawn
				entries = append(entries, entry)
			}
		}
		current.Entries = entries
		return nil
	}
	return failure("not_found", "Provider is not currently maintained")
}
