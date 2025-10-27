package server

import (
	"net/http"

	"github.com/maevlava/resume-backend/internal/features/ai"
	"github.com/maevlava/resume-backend/internal/features/auth"
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
		AuthHandler:   auth.NewHandler(cfg, db),
		UploadHandler: upload.NewHandler(store),
		AIHandler:     ai.NewHandler(aiService, store),
	}

	s.RegisterRoutes()
	return s
}

func (s *ResumeProtoServer) RegisterRoutes() {
	cors := middleware.EnableCORS
	requireAuth := middleware.RequireAuth(s.cfg)

	s.AuthHandler.RegisterRoutes(s.Router, cors)
	s.UploadHandler.RegisterRoutes(s.Router, cors, requireAuth)
	s.AIHandler.RegisterRoutes(s.Router, cors, requireAuth)

	// static files
	fileServer := http.FileServer(http.Dir(s.cfg.StoragePath))
	s.Router.Handle("/uploads/", http.StripPrefix("/uploads/", fileServer))

	//TOOD buka static image server

}
