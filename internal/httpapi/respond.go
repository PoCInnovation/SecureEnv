package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/PoCInnovation/SecureEnv/internal/apiv1"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

var errInvalidRequest = errors.New("invalid request")

// writeError renders err as an apiv1.ErrorResponse. Details of server side
// failures are logged, never returned.
func writeError(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	var (
		code    apiv1.Code
		status  int
		message = err.Error()
		tooBig  *http.MaxBytesError
	)
	switch {
	case errors.As(err, &tooBig):
		code, status = apiv1.CodeInvalidRequest, http.StatusRequestEntityTooLarge
	case errors.Is(err, errInvalidRequest):
		code, status = apiv1.CodeInvalidRequest, http.StatusBadRequest
	default:
		code, status = apiv1.FromError(err)
	}

	if status >= http.StatusInternalServerError {
		logger.ErrorContext(r.Context(), "request failed", slog.Any("error", err))
		message = "internal server error"
	}
	writeJSON(w, status, apiv1.ErrorResponse{Error: apiv1.ErrorBody{Code: code, Message: message}})
}

// decodeJSON strictly decodes a single JSON object from the request body.
func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return err
		}
		return fmt.Errorf("%w: %w", errInvalidRequest, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: body must contain a single JSON object", errInvalidRequest)
	}
	return nil
}
