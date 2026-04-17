package e2e

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
)

func (r *TestRig) doRequest(t *testing.T, req *http.Request) (int, []byte) {
	res, err := r.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return res.StatusCode, body
}

func (r *TestRig) createUser(t *testing.T, username, password, email string) func(t *testing.T) {
	data := url.Values{
		"username": {username},
		"password": {password},
		"email":    {email},
	}

	registerEndpoint := r.Wiki.Config.URL.JoinPath("register").String()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, registerEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	actorId := r.Wiki.Config.URL.JoinPath("u", username).String()
	getActor, err := http.NewRequestWithContext(t.Context(), http.MethodGet, actorId, nil)
	getActor.Header.Add("Accept", "application/ld+json")
	if err != nil {
		t.Fatal(err)
	}

	return func(t *testing.T) {
		code, _ := r.doRequest(t, req)

		code, body := r.doRequest(t, getActor)
		if code > 200 {
			t.Errorf("request failed with code %d", code)
		}
		obj, err := activitystreams.ReadObject(body)
		if err != nil {
			t.Fatal(err)
		}

		actor, err := obj.AsActor()
		if err != nil {
			t.Fatal(err)
		}

		if actor.Username != username {
			t.Errorf("expected username %s, got %s", username, actor.Username)
		}
	}
}
