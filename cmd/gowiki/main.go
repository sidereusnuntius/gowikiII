package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/sidereusnuntius/gowiki/cmd/gowiki/setup"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

func main() {
	ctx := context.Background()
	dbConfig := config.DbConfig{
		URL: "./test.db",
	}

	// log.With().Stack().Logger()
	wikilog.Logger.Info().Str("database URL", dbConfig.URL).Msg("attempting to connect to Sqlite database")
	db, err := db.Open(ctx, dbConfig)
	if err != nil {
		wikilog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	defer func() {
		wikilog.Logger.Info().Msg("closing database handle")
		if err := db.Close(); err != nil {
			wikilog.Logger.Error().Err(err).Msg("failed to close database handle")
		}
	}()

	wiki := setup.SetupWiki(db)
	wikilog.Logger.Info().
		Int("port", config.Config.Port).
		Str("address", config.Config.Host).
		Str("name", config.Config.Name).
		Msgf("starting gowiki")
	if err := wiki.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		wikilog.Logger.Fatal().Err(err).Msg("failed to start http server")
	}
	wikilog.Logger.Info().Msg("stopped server")
}
