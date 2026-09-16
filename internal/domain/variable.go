package domain

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// ReservedPrefix marks the variables SecureEnv uses for its own configuration.
// They live in local .env files and are never synchronised.
const ReservedPrefix = "SECURE_ENV_"

var variableKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// VariableKey is a validated environment variable name.
type VariableKey struct{ value string }

// NewVariableKey validates raw and returns a VariableKey.
func NewVariableKey(raw string) (VariableKey, error) {
	if !variableKeyPattern.MatchString(raw) {
		return VariableKey{}, fmt.Errorf("%w: %q must match %s", ErrInvalidVariableKey, raw, variableKeyPattern)
	}
	if IsReserved(raw) {
		return VariableKey{}, fmt.Errorf("%w: %q uses the %s prefix", ErrReservedVariableKey, raw, ReservedPrefix)
	}
	return VariableKey{value: raw}, nil
}

func (k VariableKey) String() string { return k.value }

// IsReserved reports whether key belongs to SecureEnv itself.
func IsReserved(key string) bool {
	return strings.HasPrefix(strings.ToUpper(key), ReservedPrefix)
}

// Variables is an immutable set of validated environment variables.
// The zero value is an empty set.
type Variables struct{ entries map[string]string }

// NewVariables validates every key of raw and copies it.
func NewVariables(raw map[string]string) (Variables, error) {
	for key := range raw {
		if _, err := NewVariableKey(key); err != nil {
			return Variables{}, err
		}
	}
	return Variables{entries: maps.Clone(raw)}, nil
}

// Get returns the value of key and whether it is set.
func (v Variables) Get(key string) (string, bool) {
	value, ok := v.entries[key]
	return value, ok
}

// Len returns the number of variables.
func (v Variables) Len() int { return len(v.entries) }

// Keys returns the variable names in lexical order.
func (v Variables) Keys() []string {
	return slices.Sorted(maps.Keys(v.entries))
}

// Map returns a copy of the variables.
func (v Variables) Map() map[string]string {
	out := make(map[string]string, len(v.entries))
	maps.Copy(out, v.entries)
	return out
}

// With returns a copy of v where key is set to value.
func (v Variables) With(key VariableKey, value string) Variables {
	out := v.Map()
	out[key.String()] = value
	return Variables{entries: out}
}

// Without returns a copy of v without key.
func (v Variables) Without(key string) Variables {
	out := v.Map()
	delete(out, key)
	return Variables{entries: out}
}
