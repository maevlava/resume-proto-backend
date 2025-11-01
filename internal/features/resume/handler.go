package resume

import (
	"net/http"

	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/server/httperror"
	"github.com/maevlava/resume-backend/internal/shared/storage"
)

type Handler struct {
	service *Service
}

func NewHandler(db *db.Queries, store storage.Store) *Handler {
	return &Handler{
		service: NewService(db, store),
	}
}
func (h *Handler) RegisterRoutes(r *http.ServeMux, baseApiPath string, mws ...middleware.Middleware) {
	fullPath := baseApiPath + "/resume"

	use := func(path string, fn httperror.HandlerFunc) {
		route := httperror.Handler(fn)
		route = middleware.Chain(route, mws...)
		r.Handle(fullPath+path, route)
	}

	use("/{id}", h.GetResume)
}

func (h *Handler) GetResume(w http.ResponseWriter, r *http.Request) error {
	resumeID := r.PathValue("id")

	resume, err := h.service.GetResumeByID(r.Context(), resumeID)
	if err != nil {
		return httperror.BadRequest("GetResume: failed to get resume", err)
	}

	common.RespondWithJSON(w, http.StatusOK, resume)
	return nil
}
