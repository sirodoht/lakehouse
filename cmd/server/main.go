package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	lakehouse "git.sr.ht/~sirodoht/lakehouse/internal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// debug mode
	debugMode := os.Getenv("DEBUG")

	// database connection
	databaseURL := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		panic(err)
	}

	// instantiate
	store := lakehouse.NewSQLStore(db)
	handlerAPI := lakehouse.NewHandlerAPI(store)
	handlerPage := lakehouse.NewHandlerPage(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// midd to check if user is authenticated
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var username string
			isAuthenticated := false
			c, err := r.Cookie("session")
			if err != nil {
				fmt.Println(err)
			} else {
				username = store.GetUsernameSession(r.Context(), c.Value)
				if err == nil {
					isAuthenticated = true
				}
			}
			ctx := context.WithValue(r.Context(), lakehouse.KeyUsername, username)
			ctx = context.WithValue(ctx, lakehouse.KeyIsAuthenticated, isAuthenticated)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Page Index
	r.Get("/", handlerPage.RenderIndex)

	// Page Documents
	r.Get("/docs", handlerPage.RenderAllDocument)
	r.Get("/new/doc", handlerPage.RenderNewDocument)
	r.Post("/new/doc", handlerPage.SaveNewDocument)
	r.Get("/docs/{id}", handlerPage.RenderOneDocument)
	r.Get("/docs/{id}/edit", handlerPage.RenderEditDocument)
	r.Post("/docs/{id}/edit", handlerPage.SaveEditDocument)

	// API Documents
	r.Post("/api/docs", handlerAPI.InsertDocumentHandler)
	r.Get("/api/docs", handlerAPI.GetAllDocumentHandler)
	r.Patch("/api/docs/{id}", handlerAPI.UpdateDocumentHandler)
	r.Get("/api/docs/{id}", handlerAPI.GetOneDocumentHandler)

	// API Users
	r.Post("/api/users", handlerAPI.InsertUserHandler)
	r.Get("/api/users/{id}", handlerAPI.GetOneUserHandler)
	r.Patch("/api/users/{id}", handlerAPI.UpdateUserHandler)

	// Page Users
	r.Get("/signup", handlerPage.RenderNewUser)
	r.Post("/signup", handlerPage.SaveNewUser)
	r.Get("/login", handlerPage.RenderLogin)
	r.Post("/login", handlerPage.CreateSession)
	r.Post("/logout", handlerPage.DeleteSession)

	// dashboard
	r.Get("/dashboard", handlerPage.RenderDashboard)
	r.Get("/editor", handlerPage.RenderEditor)

	// static files
	if debugMode == "1" {
		fileServer := http.FileServer(http.Dir("./static/"))
		r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	}

	// serve
	fmt.Println("Listening on http://127.0.0.1:8000/")
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
