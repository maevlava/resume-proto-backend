package upload

import (
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
	"net/http"
	"path/filepath"
)

type Handler struct {
	store storage.Store
}

func NewHandler(store storage.Store) *Handler {
	return &Handler{
		store: store,
	}
}
func (h *Handler) RegisterRoutes(r *http.ServeMux, mws ...middleware.Middleware) {
	r.Handle("/api/v1/upload", middleware.Chain(http.HandlerFunc(h.Upload), mws...))
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	// user from middleware context
	username, ok := r.Context().Value("username").(string)
	if !ok {
		log.Error().Msg("User not found in context")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// request by form-data
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Error().Msg("Failed to read file from request")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request")
	}
	defer file.Close()

	uploadRequest := UploadRequest{
		File:           file,
		JobTitle:       r.FormValue("job_title"),
		JobDescription: r.FormValue("job_description"),
		CompanyName:    r.FormValue("company_name"),
	}

	// username/job title/file
	path := filepath.Join(username, uploadRequest.JobTitle, header.Filename)

	// save
	if err := h.store.Save(path, file); err != nil {
		log.Error().Err(err).Msg("Error saving file")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
}
