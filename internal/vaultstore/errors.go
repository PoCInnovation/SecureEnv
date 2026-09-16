package vaultstore

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	vault "github.com/hashicorp/vault/api"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// casMismatch is the message Vault returns when a check-and-set fails.
const casMismatch = "check-and-set parameter did not match the current version"

// mapError translates Vault client errors into domain errors, keeping the
// original error in the chain for logging.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, vault.ErrSecretNotFound) {
		return fmt.Errorf("%w: %w", domain.ErrProjectNotFound, err)
	}

	var responseErr *vault.ResponseError
	if !errors.As(err, &responseErr) {
		return fmt.Errorf("vault: %w", err)
	}

	switch responseErr.StatusCode {
	case http.StatusBadRequest:
		if hasMessage(responseErr, casMismatch) {
			return fmt.Errorf("%w: %w", domain.ErrVersionConflict, err)
		}
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %w", domain.ErrUnauthorized, err)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %w", domain.ErrForbidden, err)
	case http.StatusNotFound:
		return fmt.Errorf("%w: %w", domain.ErrProjectNotFound, err)
	}
	return fmt.Errorf("vault: %w", err)
}

func hasMessage(err *vault.ResponseError, message string) bool {
	for _, e := range err.Errors {
		if strings.Contains(e, message) {
			return true
		}
	}
	return false
}
