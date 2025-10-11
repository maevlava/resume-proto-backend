package server

import (
	"github.com/maevlava/resume-backend/internal/features/auth"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/middleware"
	"net/http"
)

type ResumeProtoServer struct {
	cfg         *config.Config
	db          *db.Queries
	Router      *http.ServeMux
	Address     string
	AuthHandler *auth.Handler
}

func NewResumeProtoServer(cfg *config.Config, db *db.Queries) *ResumeProtoServer {
	s := &ResumeProtoServer{
		cfg:         cfg,
		db:          db,
		Address:     cfg.ServerAddress,
		Router:      http.NewServeMux(),
		AuthHandler: auth.NewHandler(cfg, db),
	}

	s.RegisterRoutes()
	return s
}

func (s *ResumeProtoServer) RegisterRoutes() {
	cors := middleware.EnableCORS
	s.AuthHandler.RegisterRoutes(s.Router, cors)
}
