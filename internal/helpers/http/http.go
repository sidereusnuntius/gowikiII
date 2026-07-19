package httphelpers

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/goccy/go-json"
)

var isFederatedRequest = regexp.MustCompile(`application\/((ld)|(activity))\+json`)

func LocalReferer(r *http.Request, host string) (string, bool) {
	header := r.Header.Get("Referer")
	if len(header) == 0 {
		return "", false
	}

	url, err := url.Parse(header)
	if err != nil {
		return "", false
	}

	if url.Host != "" && url.Host != host {
		return "", false
	}

	return url.Path, true
}

func FedWebMux(fed, web http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isFederatedRequest.MatchString(r.Header.Get("Accept")) || isFederatedRequest.MatchString(r.Header.Get("Content-Type")) {
			fed.ServeHTTP(w, r)
			return
		}

		web.ServeHTTP(w, r)
	})
}

func WriteActivity(w http.ResponseWriter, object any) error {
	header := w.Header()
	header.Set("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")

	encoder := json.NewEncoder(w)
	return encoder.Encode(object)
}
