package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"git.sr.ht/~sirodoht/lakehousewiki/document"
	"git.sr.ht/~sirodoht/lakehousewiki/eliot"
	"git.sr.ht/~sirodoht/lakehousewiki/session"
	"git.sr.ht/~sirodoht/lakehousewiki/user"

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

	// instantiate stores
	documentStore := document.NewSQLStore(db)
	userStore := user.NewSQLStore(db)
	sessionStore := session.NewSQLStore(db)

	// instantiate APIs
	documentAPI := document.NewAPI(documentStore)
	userAPI := user.NewAPI(userStore)

	// instantiate Pages
	userPage := user.NewPage(userStore, sessionStore)
	documentPage := document.NewPage(documentStore)

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
				username = sessionStore.GetUsername(r.Context(), c.Value)
				if err == nil {
					isAuthenticated = true
				}
			}
			ctx := context.WithValue(r.Context(), eliot.KeyUsername, username)
			ctx = context.WithValue(ctx, eliot.KeyIsAuthenticated, isAuthenticated)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t, err := template.ParseFiles("templates/layout.html", "templates/index.html")
		if err != nil {
			panic(err)
		}
		err = t.Execute(w, map[string]interface{}{
			"IsAuthenticated": r.Context().Value(eliot.KeyIsAuthenticated),
			"Username":        r.Context().Value(eliot.KeyUsername),
		})
		if err != nil {
			panic(err)
		}
	})

	// Page Documents
	r.Get("/docs", documentPage.RenderAll)
	r.Get("/new/doc", documentPage.RenderNew)
	r.Post("/new/doc", documentPage.SaveNew)
	r.Get("/docs/{id}", documentPage.RenderOne)
	r.Get("/docs/{id}/edit", documentPage.RenderEdit)
	r.Post("/docs/{id}/edit", documentPage.SaveEdit)

	// API Documents
	r.Post("/api/docs", documentAPI.InsertHandler)
	r.Get("/api/docs", documentAPI.GetAllHandler)
	r.Patch("/api/docs/{id}", documentAPI.UpdateHandler)
	r.Get("/api/docs/{id}", documentAPI.GetOneHandler)

	// API Users
	r.Post("/api/users", userAPI.InsertHandler)
	r.Get("/api/users/{id}", userAPI.GetOneHandler)
	r.Patch("/api/users/{id}", userAPI.UpdateHandler)

	// Page Users
	r.Get("/signup", userPage.RenderNew)
	r.Post("/signup", userPage.SaveNew)
	r.Get("/login", userPage.RenderLogin)
	r.Post("/login", userPage.CreateSession)
	r.Post("/logout", userPage.DeleteSession)

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
