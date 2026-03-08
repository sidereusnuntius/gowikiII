package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/db/sqlite"
	"github.com/sidereusnuntius/gowiki/internal/server"
	"github.com/sidereusnuntius/gowiki/internal/transactions"
	"github.com/sidereusnuntius/gowiki/internal/wiki"
)

func main() {
	pool, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		fmt.Printf("failed to open connection: %w", err)
	}

	if err = pool.Ping(); err != nil {
		_ = pool.Close()
		fmt.Printf("failed to ping database: %w", err)
	}

	tm := txdb.TxManager{DB: pool}

	store, err := sqlite.Init(pool)
	if err != nil {
		fmt.Println(err)
	}

	auth := wiki.NewAuth(store, &tm)
	handler := server.AuthHandler{
		AuthService: auth,
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	fileServer := http.FileServer(http.Dir("./static"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
