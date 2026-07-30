package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"github.com/sidereusnuntius/gowiki/internal/i18n/messages"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
	"golang.org/x/text/language"
)

var (
	localizer *i18n.Localizer
	bundle    *i18n.Bundle
)

// Make the language preference dependent on the browser's preferences.
func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFileFS(messages.MessagesFS, "en.toml")
	bundle.LoadMessageFileFS(messages.MessagesFS, "pt.toml")

	localizer = i18n.NewLocalizer(bundle,
		language.Portuguese.String(),
		language.English.String(),
	)
}

type opt func(*i18n.LocalizeConfig)

// func WithFallback()

// TODO: handle ctx data.
func T(message string) string {
	config := i18n.LocalizeConfig{
		MessageID: message,
	}

	localized, _, err := localizer.LocalizeWithTag(&config)
	if err != nil {
		wikilog.Logger.Error().Err(err).Send()
	}

	return localized
}
