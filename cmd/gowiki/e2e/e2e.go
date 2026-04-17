package e2e

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"testing"

	"github.com/sidereusnuntius/gowiki/cmd/gowiki/setup"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/search"
	"github.com/sidereusnuntius/gowiki/internal/tests"
)

type TestRig struct {
	Wiki   setup.Wiki
	Client *http.Client
}

var wg sync.WaitGroup

func Start(t *testing.T, addr string) TestRig {
	cfg := tests.TestConfig(addr)
	t.Log("config", cfg)
	db, err := db.Open(context.Background(), config.DbConfig{
		URL: ":memory:",
	})
	if err != nil {
		t.Fatal(err)
	}

	search, err := search.TestSearch()
	if err != nil {
		t.Fatal(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	rig := TestRig{
		Wiki: setup.SetupWiki(cfg, db, search),
		Client: &http.Client{
			Jar: jar,
		},
	}

	wg.Go(func() {
		t.Log("listening!")
		if err := rig.Wiki.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatal("failed to listen and serve:", err)
		}
	})

	return rig
}

func (r *TestRig) Close(t *testing.T) {
	if err := r.Wiki.Server.Shutdown(t.Context()); err != nil {
		t.Error(err)
	}
}
