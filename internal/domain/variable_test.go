package domain_test

import (
	"errors"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

func TestNewVariableKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		wantErr error
	}{
		{input: "DATABASE_URL"},
		{input: "_private"},
		{input: "api_key2"},
		{input: "", wantErr: domain.ErrInvalidVariableKey},
		{input: "2FA", wantErr: domain.ErrInvalidVariableKey},
		{input: "MY-VAR", wantErr: domain.ErrInvalidVariableKey},
		{input: "A B", wantErr: domain.ErrInvalidVariableKey},
		{input: "SECURE_ENV_TOKEN", wantErr: domain.ErrReservedVariableKey},
		{input: "secure_env_token", wantErr: domain.ErrReservedVariableKey},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewVariableKey(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewVariableKey(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr == nil && got.String() != tt.input {
				t.Fatalf("String() = %q, want %q", got.String(), tt.input)
			}
		})
	}
}

func TestIsReserved(t *testing.T) {
	t.Parallel()

	if !domain.IsReserved("SECURE_ENV_PROJECT") {
		t.Error("SECURE_ENV_PROJECT should be reserved")
	}
	if domain.IsReserved("MY_SECURE_ENV_VALUE") {
		t.Error("only the prefix is reserved")
	}
}

func TestNewVariables(t *testing.T) {
	t.Parallel()

	vars, err := domain.NewVariables(map[string]string{"A": "1", "B": "2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := vars.Keys(); len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Fatalf("Keys() = %v, want sorted [A B]", got)
	}

	if _, err := domain.NewVariables(map[string]string{"bad-key": "1"}); !errors.Is(err, domain.ErrInvalidVariableKey) {
		t.Fatalf("error = %v, want ErrInvalidVariableKey", err)
	}
}

func TestVariablesAreImmutable(t *testing.T) {
	t.Parallel()

	source := map[string]string{"A": "1"}
	vars, err := domain.NewVariables(source)
	if err != nil {
		t.Fatal(err)
	}
	source["A"] = "changed"

	exported := vars.Map()
	exported["A"] = "changed too"

	if v, _ := vars.Get("A"); v != "1" {
		t.Fatalf("Variables leaked a reference: A = %q", v)
	}

	withB := vars.With(mustKey(t, "B"), "2")
	if _, ok := vars.Get("B"); ok {
		t.Fatal("With must not modify the receiver")
	}
	if v, _ := withB.Get("B"); v != "2" {
		t.Fatalf("With did not set B, got %q", v)
	}

	withoutA := vars.Without("A")
	if _, ok := withoutA.Get("A"); ok {
		t.Fatal("Without did not remove A")
	}
	if withoutA.Len() != 0 || vars.Len() != 1 {
		t.Fatal("Without must not modify the receiver")
	}
}

func mustKey(t *testing.T, raw string) domain.VariableKey {
	t.Helper()
	key, err := domain.NewVariableKey(raw)
	if err != nil {
		t.Fatal(err)
	}
	return key
}
