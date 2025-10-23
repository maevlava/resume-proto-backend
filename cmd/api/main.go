package main

import (
	"context"
	"database/sql"
	_ "database/sql"
	"net/http"
	"time"

	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/deepseek"
	"github.com/maevlava/resume-backend/internal/shared/logger"
	"github.com/maevlava/resume-backend/internal/shared/server"
	"github.com/maevlava/resume-backend/internal/shared/storage"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init()
	cfg := config.LoadConfig()
	conn, queries := connectToDatabase(context.Background(), cfg)
	defer conn.Close()

	FSStore, err := storage.NewFSStore(cfg.StoragePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating storage")
		return
	}

	deepseekClient := deepseek.NewClient(cfg.DeepseekAPIKey)

	srv := server.NewResumeProtoServer(cfg, queries, FSStore, deepseekClient)

	log.Info().Msgf("Starting server on %s", srv.Address)
	err = http.ListenAndServe(srv.Address, srv.Router)
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}
}

func connectToDatabase(ctx context.Context, cfg *config.Config) (*sql.DB, *db.Queries) {
	conn, err := sql.Open("postgres", cfg.DBString)
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to database")
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("Error pinging database")
	}
	log.Info().Msg("Successfully connected to the database!")

	queries := db.New(conn)

	return conn, queries
}
