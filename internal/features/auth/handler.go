package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/maevlava/resume-backend/internal/features/resume"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/server/httperror"
	"github.com/maevlava/resume-backend/internal/shared/storage"
)

type Handler struct {
	service       *Service
	resumeService *resume.Service
}

func NewHandler(cfg *config.Config, db *db.Queries, store storage.Store) *Handler {
	return &Handler{
		service:       NewService(cfg, db),
		resumeService: resume.NewService(db, store),
	}
}
func (h *Handler) RegisterRoutes(r *http.ServeMux, baseApiPath string, mws ...middleware.Middleware) {
	fullPath := baseApiPath + "/auth"

	use := func(path string, fn httperror.HandlerFunc) {
		route := httperror.Handler(fn)
		route = middleware.Chain(route, mws...)
		r.Handle(fullPath+path, route)
	}

	use("/login", h.Login)
	use("/logout", h.Logout)
	use("/register", h.Register)
	use("/validate", h.Validate)
	use("/me", h.GetSignedInUser)
	use("/me/resumes", h.GetUserResumes)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var loginRequest LoginRequest
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return httperror.BadRequest("login: failed reading request body", err)
	}

	if err := json.Unmarshal(reqBody, &loginRequest); err != nil {
		return httperror.BadRequest("login: failed parsing request body", err)
	}

	token, err := h.service.Login(r.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		return httperror.BadRequest("login: failed to login", err)
	}

	// register to cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token.Value,
		Path:     "/",
		Expires:  token.Duration.Time,
		HttpOnly: true,
		Secure:   false, // change to true in production
	})

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	// delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	common.RespondWithJSON(w, http.StatusOK, map[string]string{})

	return nil
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	// get parameters
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return httperror.BadRequest("register: failed reading request body", err)
	}

	var registerRequest RegisterRequest
	if err = json.Unmarshal(reqBody, &registerRequest); err != nil {
		return httperror.BadRequest("register: failed parsing request body", err)
	}

	err = h.service.Register(r.Context(), registerRequest.Name, registerRequest.Email, registerRequest.Password)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return httperror.BadRequest("register: failed to register", err)
		}
		return httperror.InternalServerError("register: failed to register", err)
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) error {
	// get token from cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return httperror.BadRequest("validate: failed to get cookie", err)
	}

	err = h.service.Validate(cookie.Value)
	if err != nil {
		return httperror.Unauthorized("validate: failed to validate token")
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) GetSignedInUser(w http.ResponseWriter, r *http.Request) error {
	// check cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return httperror.BadRequest("getSignedInUser: failed to get cookie", err)
	}

	//return user
	user, err := h.service.GetSignedInUser(r.Context(), cookie.Value)
	if err != nil {
		return httperror.BadRequest("getSignedInUser: failed to get user", err)
	}

	common.RespondWithJSON(w, http.StatusOK, user)
	return nil
}

func (h *Handler) GetUserResumes(w http.ResponseWriter, r *http.Request) error {
	// check cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return httperror.BadRequest("getUserResumes: failed to get cookie", err)
	}

	//user
	user, err := h.service.GetSignedInUser(r.Context(), cookie.Value)
	if err != nil {
		return httperror.BadRequest("getUserResumes: failed to get user", err)
	}

	// resumes
	resumes, err := h.resumeService.GetAllResumesByID(r.Context(), user.ID)
	if err != nil {
		return httperror.InternalServerError("getUserResumes: failed to get resumes", err)
	}

	common.RespondWithJSON(w, http.StatusOK, resumes)
	return nil
}
