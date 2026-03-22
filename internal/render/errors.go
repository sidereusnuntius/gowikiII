package render

import (
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

func (p *Page) HandleError(err error) {
	header := p.writer.Header()
	switch newerr := err.(type) {
	case wikierr.ValidationError:
		for k, e := range newerr.Fields {
			header.Set("Err-"+k, e.Error())
		}
		p.writer.WriteHeader(http.StatusBadRequest)
	default:
		header.Set("FX-Error", err.Error())
		p.writer.WriteHeader(http.StatusInternalServerError)
		wikilog.Logger.Error().Err(err).Msg("internal server error")
	}
}
