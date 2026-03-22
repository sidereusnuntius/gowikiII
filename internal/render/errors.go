package render

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

func (p *Page) HandleError(err error) {
	switch newerr := err.(type) {
	case wikierr.ValidationError:
		header := p.writer.Header()
		for k, e := range newerr.Fields {
			header.Set("Err-"+k, e.Error())
		}
		p.writer.WriteHeader(http.StatusBadRequest)
	default:
		log.Error().Err(err).Msg("internal server error")
	}
}
