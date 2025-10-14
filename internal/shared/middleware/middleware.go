package middleware

import (
	"context"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Middleware func(http.Handler) http.HandlerFunc

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
func EnableCORS(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	}
}
func RequireAuth(cfg *config.Config) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				log.Error().Err(err).Msg("Error getting cookie")
				common.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			claims, err := common.ValidateJWT(cookie.Value, cfg.JWTSecret)
			if err != nil {
				log.Error().Err(err).Msg("Error validating token")
				common.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), "username", claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
