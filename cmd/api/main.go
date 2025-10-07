package main

import (
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/logger"
	"github.com/maevlava/resume-backend/internal/shared/server"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	logger.Init()
	cfg := config.LoadConfig()
	srv := server.NewResumeProtoServer(cfg)

	log.Info().Msgf("Starting server on %s", srv.Address)
	err := http.ListenAndServe(srv.Address, srv.Router)
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}
}
