package main

import (
	"context"
	"database/sql"
	_ "database/sql"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/logger"
	"github.com/maevlava/resume-backend/internal/shared/server"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func main() {
	logger.Init()
	cfg := config.LoadConfig()
	conn, queries := connectToDatabase(context.Background(), cfg)
	defer conn.Close()

	srv := server.NewResumeProtoServer(cfg, queries)

	log.Info().Msgf("Starting server on %s", srv.Address)
	err := http.ListenAndServe(srv.Address, srv.Router)
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
