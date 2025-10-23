package ai

import (
	"context"
	"encoding/json/v2"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/deepseek"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"time"
)

type Handler struct {
	AIService *deepseek.Client
	Store     storage.Store
}

func NewHandler(AIService *deepseek.Client, store storage.Store) *Handler {
	return &Handler{
		AIService: AIService,
		Store:     store,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux, mws ...middleware.Middleware) {
	router.Handle("/api/v1/ai/analyze", middleware.Chain(http.HandlerFunc(h.AnalyzeResume), mws...))
}

func (h *Handler) AnalyzeResume(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var resumeRequest AnalyzeResumeRequest
	if err := json.Unmarshal(reqBody, &resumeRequest); err != nil {
		log.Error().Err(err).Msg("Error parsing request body")
		common.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	chatResponse, err := Analyze(ctx, h.AIService, h.Store, resumeRequest)
	if err != nil {
		log.Error().Err(err).Msg("Error analyzing resume")
		common.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	common.RespondWithJSON(w, http.StatusOK, chatResponse)
}
