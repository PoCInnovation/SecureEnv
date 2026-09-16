package apiv1_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/apiv1"
	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

func TestErrorMappingRoundTrip(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		domain.ErrInvalidProjectName,
		domain.ErrInvalidVariableKey,
		domain.ErrReservedVariableKey,
		domain.ErrProjectNotFound,
		domain.ErrProjectExists,
		domain.ErrVariableNotFound,
		domain.ErrVersionConflict,
		domain.ErrUnauthorized,
		domain.ErrForbidden,
	} {
		wrapped := fmt.Errorf("context: %w", err)
		code, status := apiv1.FromError(wrapped)
		if code == apiv1.CodeInternal || status == http.StatusInternalServerError {
			t.Errorf("%v is not mapped", err)
		}
		if back := apiv1.ToError(code); !errors.Is(back, err) {
			t.Errorf("ToError(%s) = %v, want %v", code, back, err)
		}
	}
}

func TestUnknownError(t *testing.T) {
	t.Parallel()

	code, status := apiv1.FromError(errors.New("boom"))
	if code != apiv1.CodeInternal || status != http.StatusInternalServerError {
		t.Fatalf("FromError() = %s %d", code, status)
	}
	if apiv1.ToError(apiv1.CodeInternal) != nil {
		t.Fatal("internal has no domain error")
	}
}
