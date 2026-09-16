package domain

// Changes lists the keys that differ between a local and a remote set of
// variables, seen from the local side. Every slice is sorted.
type Changes struct {
	Added    []string // present locally only
	Modified []string // present on both sides with different values
	Removed  []string // present remotely only
}

// InSync reports whether both sides hold exactly the same variables.
func (c Changes) InSync() bool {
	return len(c.Added) == 0 && len(c.Modified) == 0 && len(c.Removed) == 0
}

// Diff compares local against remote.
func Diff(local, remote Variables) Changes {
	var changes Changes
	for _, key := range local.Keys() {
		remoteValue, ok := remote.Get(key)
		localValue, _ := local.Get(key)
		switch {
		case !ok:
			changes.Added = append(changes.Added, key)
		case remoteValue != localValue:
			changes.Modified = append(changes.Modified, key)
		}
	}
	for _, key := range remote.Keys() {
		if _, ok := local.Get(key); !ok {
			changes.Removed = append(changes.Removed, key)
		}
	}
	return changes
}
