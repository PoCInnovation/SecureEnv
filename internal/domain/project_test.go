package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

func TestNewProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "simple", input: "backend"},
		{name: "git remote style", input: "PoCInnovation_SecureEnv"},
		{name: "dots and dashes", input: "my-app.v2"},
		{name: "empty", input: "", wantErr: true},
		{name: "slash would create a sub path", input: "a/b", wantErr: true},
		{name: "leading dot", input: ".hidden", wantErr: true},
		{name: "space", input: "my app", wantErr: true},
		{name: "path traversal", input: "..", wantErr: true},
		{name: "too long", input: strings.Repeat("a", 129), wantErr: true},
		{name: "max length", input: strings.Repeat("a", 128)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewProjectName(tt.input)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrInvalidProjectName) {
					t.Fatalf("NewProjectName(%q) error = %v, want ErrInvalidProjectName", tt.input, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewProjectName(%q) unexpected error: %v", tt.input, err)
			}
			if got.String() != tt.input {
				t.Fatalf("String() = %q, want %q", got.String(), tt.input)
			}
		})
	}
}
