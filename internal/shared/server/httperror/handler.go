package httperror

import (
	"encoding/json/v2"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Handler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		var appErr *AppError
		if errors.As(err, &appErr) {
			if appErr.Err != nil {
				log.Error().Err(appErr.Err).Str("code", appErr.Code).Msg("Error handling request")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(appErr.StatusCode)
			body, _ := json.Marshal(map[string]interface{}{
				"error": appErr,
			})
			_, _ = w.Write(body)
			return
		}

		log.Error().Err(err).Msg("Unexpected error")
		body, _ := json.Marshal(map[string]any{
			"error": map[string]any{
				"code":    "INTERNAL_ERROR",
				"message": "internal server error",
			},
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(body)
	}
}
