package ai

import (
	"context"
	"net/http"
	"time"

	"github.com/maevlava/resume-backend/internal/features/resume"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/deepseek"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/server/httperror"
	"github.com/maevlava/resume-backend/internal/shared/storage"
)

var (
	AnalyzeTimeout = 45 * time.Second
)

type Handler struct {
	service *Service
}

func NewHandler(store storage.Store, ai *deepseek.Client, db *db.Queries) *Handler {
	rs := resume.NewService(db, store)

	return &Handler{
		service: NewService(store, ai, db, rs),
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux, baseApiPath string, mws ...middleware.Middleware) {
	fullPath := baseApiPath + "/ai"
	use := func(path string, fn httperror.HandlerFunc) {
		route := httperror.Handler(fn)
		route = middleware.Chain(route, mws...)
		router.Handle(fullPath+path, route)
	}
	use("/analyze/{id}", h.AnalyzeResume)
}

func (h *Handler) AnalyzeResume(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), AnalyzeTimeout)
	defer cancel()

	resumeID := r.PathValue("id")
	if resumeID == "" {
		return httperror.BadRequest("AnalyzeResume: missing resume id", nil)
	}

	updatedResumeID, err := h.service.Analyze(ctx, resumeID)
	if err != nil {
		return httperror.InternalServerError("AnalyzeResume: failed to analyze resume", err)
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]string{"resumeID": updatedResumeID.String()})
	return nil
}
