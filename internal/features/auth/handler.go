package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	cfg *config.Config
	db  *db.Queries
}

func NewHandler(cfg *config.Config, db *db.Queries) *Handler {
	return &Handler{
		cfg: cfg,
		db:  db,
	}
}
func (h *Handler) RegisterRoutes(r *http.ServeMux, mws ...middleware.Middleware) {
	r.Handle("/api/v1/auth/login", middleware.Chain(http.HandlerFunc(h.Login), mws...))
	r.Handle("POST /api/v1/auth/register", middleware.Chain(http.HandlerFunc(h.Register), mws...))
	r.Handle("/api/v1/auth/validate", middleware.Chain(http.HandlerFunc(h.Validate), mws...))
	r.Handle("/api/v1/auth/logout", middleware.Chain(http.HandlerFunc(h.Logout), mws...))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// get parameters
	var loginRequest LoginRequest
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := json.Unmarshal(reqBody, &loginRequest); err != nil {
		log.Error().Err(err).Msg("Error parsing request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
	}

	// validate
	user, err := h.db.GetUserByEmail(r.Context(), loginRequest.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			common.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Error().Err(err).Msg("Error getting user by email")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	err = common.CheckPasswordHash(loginRequest.Password, user.HashedPassword)
	if err != nil {
		common.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// generate token
	tokenDuration := 24 * time.Hour
	tokenString, err := common.MakeJWT(user, h.cfg.JWTSecret, tokenDuration)
	if err != nil {
		log.Error().Err(err).Msg("Error generating token")
	}

	// register to cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(tokenDuration),
		HttpOnly: true,
		Secure:   false, // change to true in production
	})

	common.RespondWithJSON(w, http.StatusOK, map[string]string{})
	return
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   false,
		Path:     "/api/v1/auth",
	})

	common.RespondWithJSON(w, http.StatusOK, map[string]string{})
	log.Info().Msg("User logged out")
	return
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// get parameters
	var registerRequest RegisterRequest
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := json.Unmarshal(reqBody, &registerRequest); err != nil {
		log.Error().Err(err).Msg("Error parsing request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
	}
	hashPassword, err := common.HashPassword(registerRequest.Password)
	if err != nil {
		log.Error().Err(err).Msg("Error hashing password")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	user, err := h.db.CreateUser(r.Context(), db.CreateUserParams{
		ID:             uuid.New(),
		Name:           registerRequest.Name,
		Email:          registerRequest.Email,
		HashedPassword: hashPassword,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			common.RespondWithError(w, http.StatusBadRequest, "email already exists")
			return
		}
		log.Error().Err(err).Msg("Error creating user")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	common.RespondWithJSON(w, http.StatusCreated, user)
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	// get token from cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		log.Error().Err(err).Msg("Error getting cookie")
		common.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// validate token
	_, err = common.ValidateJWT(cookie.Value, h.cfg.JWTSecret)
	if err != nil {
		log.Error().Err(err).Msg("Error validating token")
		common.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	w.WriteHeader(http.StatusOK)
}
