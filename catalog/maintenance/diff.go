package maintenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

func stateChanges(before, after State) []FieldChange {
	old := map[string]Entry{}
	for _, provider := range before.Providers {
		for _, entry := range provider.Entries {
			old[identityKey(entryIdentity(entry))] = entry
		}
	}
	changes := []FieldChange{}
	for _, provider := range after.Providers {
		for _, entry := range provider.Entries {
			identity := entryIdentity(entry)
			previous, exists := old[identityKey(identity)]
			if !exists {
				changes = append(changes, FieldChange{Identity: identity, Field: "/", After: entry.Definition})
				continue
			}
			if previous.Lifecycle != entry.Lifecycle {
				a, _ := json.Marshal(previous.Lifecycle)
				b, _ := json.Marshal(entry.Lifecycle)
				changes = append(changes, FieldChange{Identity: identity, Field: "/lifecycle", Before: a, After: b})
			}
			changes = append(changes, definitionChanges(identity, "", previous.Definition, entry.Definition)...)
			oldSource, _ := json.Marshal(previous.Source)
			nextSource, _ := json.Marshal(entry.Source)
			changes = append(changes, definitionChanges(identity, "/maintenance/source", oldSource, nextSource)...)
			oldFields, _ := json.Marshal(previous.Fields)
			nextFields, _ := json.Marshal(entry.Fields)
			changes = append(changes, definitionChanges(identity, "/maintenance/fields", oldFields, nextFields)...)
		}
	}
	return changes
}

func definitionChanges(identity Identity, prefix string, before, after json.RawMessage) []FieldChange {
	var old, next map[string]json.RawMessage
	if json.Unmarshal(before, &old) == nil && old != nil && json.Unmarshal(after, &next) == nil && next != nil {
		keys := make([]string, 0, len(old)+len(next))
		seen := map[string]bool{}
		for key := range old {
			keys = append(keys, key)
			seen[key] = true
		}
		for key := range next {
			if !seen[key] {
				keys = append(keys, key)
			}
		}
		slices.Sort(keys)
		changes := []FieldChange{}
		for _, key := range keys {
			changes = append(changes, definitionChanges(identity, prefix+"/"+key, old[key], next[key])...)
		}
		return changes
	}
	// Arrays preserve element order, but object key order inside an element is
	// not a definition change. Keep the existing whole-array diff for real edits.
	var oldItems, nextItems []json.RawMessage
	if json.Unmarshal(before, &oldItems) == nil && oldItems != nil && json.Unmarshal(after, &nextItems) == nil && nextItems != nil && len(oldItems) == len(nextItems) {
		equal := true
		for i := range oldItems {
			if len(definitionChanges(identity, prefix, oldItems[i], nextItems[i])) != 0 {
				equal = false
				break
			}
		}
		if equal {
			return nil
		}
	}
	var a, b bytes.Buffer
	_ = json.Compact(&a, before)
	_ = json.Compact(&b, after)
	if bytes.Equal(a.Bytes(), b.Bytes()) {
		return nil
	}
	return []FieldChange{{Identity: identity, Field: prefix, Before: before, After: after}}
}

func identityKey(identity Identity) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", identity.Operation, identity.Provider, identity.API, identity.ID)
}
