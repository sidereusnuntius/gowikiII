package integrationtests

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/sidereusnuntius/gowiki/internal/actors"
// 	actorstore "github.com/sidereusnuntius/gowiki/internal/actors/sql"
// 	"github.com/sidereusnuntius/gowiki/internal/articles"
// 	articlesql "github.com/sidereusnuntius/gowiki/internal/articles/sql"
// 	"github.com/sidereusnuntius/gowiki/internal/config"
// 	"github.com/sidereusnuntius/gowiki/internal/federation"
// 	"github.com/sidereusnuntius/gowiki/internal/federation/client"
// 	"github.com/sidereusnuntius/gowiki/internal/keystore"
// 	sqlkeystore "github.com/sidereusnuntius/gowiki/internal/keystore/sql"
// 	"github.com/sidereusnuntius/gowiki/internal/tests"
// 	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
// )

// func testInstance(t *testing.T, config config.WikiConfig) federation.Federation {
// 	db, err := tests.TestDB(t.Context())
// 	if err != nil {
// 		t.Fatalf("failed to create in-memory database: %v", err)
// 	}

// 	client := client.New()
// 	// search, err := search.Start()
// 	// if err != nil {
// 	// 	t.Errorf("failed to start search: %v", err)
// 	// }

// 	keysStore := keystore.New(config, sqlkeystore.New(db), client)

// 	txManager := txdb.TxManager{
// 		DB: db,
// 	}
// 	actorsStore := actorstore.New(db)
// 	actors := actors.New(config, &actorsStore, &keysStore, &txManager)

// 	articles := articles.New(config, articlesql.New(db), &txManager, nil, client)

// 	return federation.New(config, client, &actors, articles)
// }

// func TestInbox(t *testing.T) {
// 	mux1 := http.NewServeMux()
// 	server1 := httptest.NewServer(mux1)

// 	mux2 := http.NewServeMux()
// 	server2 := httptest.NewServer(mux2)

// 	config1 := tests.TestConfig(server1.URL)
// 	config2 := tests.TestConfig(server2.URL)

// 	federation1 := testInstance(t, config1)
// 	federation2 := testInstance(t, config2)

// 	server1.Close()
// 	server2.Close()
// }
