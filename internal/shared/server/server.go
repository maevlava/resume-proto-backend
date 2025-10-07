package server

import (
	"github.com/maevlava/resume-backend/internal/shared/config"
	"net/http"
)

type ResumeProtoServer struct {
	cfg     *config.Config
	Router  *http.ServeMux
	Address string
}

func NewResumeProtoServer(cfg *config.Config) *ResumeProtoServer {
	s := &ResumeProtoServer{
		Address: cfg.ServerAddress,
		Router:  http.NewServeMux(),
	}

	s.RegisterRoutes()
	return s
}

func (s *ResumeProtoServer) RegisterRoutes() {
	s.Router.HandleFunc("/api/v1/resume", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
}
