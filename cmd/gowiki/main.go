package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/cmd/gowiki/setup"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	dbConfig := config.DbConfig{
		URL: "./test.db",
	}

	// log.With().Stack().Logger()
	log.Info().Msg("started gowiki")
	log.Info().Str("database URL", dbConfig.URL).Msg("attempting to connect to Sqlite database")
	db, err := db.Open(dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	defer func() {
		log.Info().Msg("closing database handle")
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database handle")
		}
	}()

	wiki := setup.SetupWiki(db)
	log.Info().
		Int("port", config.Config.Port).
		Str("address", config.Config.Host).
		Str("name", config.Config.Name).
		Msgf("starting gowiki")
	if err := wiki.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("failed to start http server")
	}
	log.Info().Msg("stopped server")
}
