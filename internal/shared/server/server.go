package server

import (
	"net/http"

	"github.com/maevlava/resume-backend/internal/features/ai"
	"github.com/maevlava/resume-backend/internal/features/auth"
	"github.com/maevlava/resume-backend/internal/features/resume"
	"github.com/maevlava/resume-backend/internal/features/upload"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/deepseek"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"github.com/maevlava/resume-backend/internal/shared/storage"
)

type ResumeProtoServer struct {
	cfg           *config.Config
	db            *db.Queries
	storage       storage.Store // change to s3 for production
	Router        *http.ServeMux
	Address       string
	AuthHandler   *auth.Handler
	UploadHandler *upload.Handler
	AIHandler     *ai.Handler
	ResumeHandler *resume.Handler
}

func NewResumeProtoServer(
	cfg *config.Config,
	db *db.Queries,
	store storage.Store,
	aiService *deepseek.Client,
) *ResumeProtoServer {

	s := &ResumeProtoServer{
		cfg:           cfg,
		db:            db,
		storage:       store,
		Address:       cfg.ServerAddress,
		Router:        http.NewServeMux(),
		AuthHandler:   auth.NewHandler(cfg, db, store),
		UploadHandler: upload.NewHandler(store, db, cfg),
		AIHandler:     ai.NewHandler(store, aiService, db),
		ResumeHandler: resume.NewHandler(db, store),
	}

	s.RegisterRoutes()
	return s
}

func (s *ResumeProtoServer) RegisterRoutes() {
	cors := middleware.EnableCORS
	requireAuth := middleware.RequireAuth(s.cfg)

	s.AuthHandler.RegisterRoutes(s.Router, s.cfg.BaseAPIPath, cors)
	s.UploadHandler.RegisterRoutes(s.Router, s.cfg.BaseAPIPath, cors, requireAuth)
	s.AIHandler.RegisterRoutes(s.Router, s.cfg.BaseAPIPath, cors, requireAuth)
	s.ResumeHandler.RegisterRoutes(s.Router, s.cfg.BaseAPIPath, cors, requireAuth)

	// static files
	fileServer := http.FileServer(http.Dir(s.cfg.StoragePath))
	s.Router.Handle("/uploads/", http.StripPrefix("/uploads/", fileServer))

}
