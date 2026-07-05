package e2e

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	"testing"

	"github.com/sidereusnuntius/gowiki/cmd/gowiki/setup"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/processor"
	"github.com/sidereusnuntius/gowiki/internal/search"
)

type TestRig struct {
	Wiki   setup.Wiki
	Client *http.Client
	DbURL  string
}

var wg sync.WaitGroup

func Start(t *testing.T, addr string) TestRig {
	buf := make([]byte, 4)
	rand.Read(buf)

	dbURL := "./" + base64.URLEncoding.EncodeToString(buf)
	cfg := config.DefaultConfig(addr)
	t.Log("config", cfg)
	db, err := db.Open(context.Background(), config.DbConfig{
		URL:  dbURL,
		Test: true,
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

	processor, err := processor.SqliteProcessor(t.Context(), db)
	if err != nil {
		t.Fatal(err)
	}

	wiki, err := setup.SetupWiki(cfg, db, search, processor)
	if err != nil {
		t.Fatal(err)
	}
	rig := TestRig{
		Wiki: wiki,
		Client: &http.Client{
			Jar: jar,
		},
		DbURL: dbURL,
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
		t.Errorf("failed to shutdown test server: %v", err)
	}

	if err := r.Wiki.DB.Close(); err != nil {
		t.Errorf("failed to close test database: %v", err)
	}

	if err := os.Remove(r.DbURL); err != nil {
		t.Errorf("failed to remove test database file: %v", err)
	}
}
