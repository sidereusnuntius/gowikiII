package wikilog

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO: builder function
var Logger *Wikilogger = &Wikilogger{
	Logger: log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(zerolog.DebugLevel),
}

type Wikilogger struct {
	zerolog.Logger
}

func (wl *Wikilogger) Fatalf(format string, v ...any) {
	wl.Logger.Fatal().Msgf(format, v...)
}

func init() {

}
