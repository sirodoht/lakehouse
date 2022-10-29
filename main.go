package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"git.sr.ht/~sirodoht/lakehousewiki/document"
	"git.sr.ht/~sirodoht/lakehousewiki/session"
	"git.sr.ht/~sirodoht/lakehousewiki/user"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type key int

const (
	keyUsername key = iota
)

func main() {
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", databaseUrl)
	if err != nil {
		panic(err)
	}

	// instantiate stores
	documentStore := document.NewSQLStore(db)
	userStore := user.NewSQLStore(db)
	sessionStore := session.NewSQLStore(db)

	// instantiate APIs
	documentApi := document.NewAPI(documentStore)
	userApi := user.NewAPI(userStore)

	// instantiate Pages
	userPage := user.NewPage(userStore, sessionStore)
	documentPage := document.NewPage(documentStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var username string
			c, err := r.Cookie("session")
			if err != nil {
				fmt.Println(err)
			} else {
				username = sessionStore.GetUsername(r.Context(), c.Value)
			}
			ctx := context.WithValue(r.Context(), keyUsername, username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// midd to check if user is authenticated
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		isAuthenticated := false
		c, err := r.Cookie("session")
		if err != nil {
			fmt.Println(err)
		} else {
			_, err := sessionStore.GetOne(r.Context(), c.Value)
			if err == nil {
				isAuthenticated = true
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t, err := template.ParseFiles("templates/layout.html", "templates/index.html")
		if err != nil {
			panic(err)
		}
		t.Execute(w, map[string]interface{}{
			"isAuthenticated": isAuthenticated,
			"username":        r.Context().Value(keyUsername),
		})
	})

	// Page Documents
	r.Get("/docs", documentPage.RenderAll)
	r.Get("/new/doc", documentPage.RenderNew)
	r.Post("/new/doc", documentPage.SaveNew)
	r.Get("/docs/{id}", documentPage.RenderOne)

	// API Documents
	r.Post("/api/docs", documentApi.InsertHandler)
	r.Get("/api/docs", documentApi.GetAllHandler)
	r.Patch("/api/docs/{id}", documentApi.UpdateHandler)
	r.Get("/api/docs/{id}", documentApi.GetOneHandler)

	// API Users
	r.Post("/api/users", userApi.InsertHandler)
	r.Get("/api/users/{id}", userApi.GetOneHandler)
	r.Patch("/api/users/{id}", userApi.UpdateHandler)

	// Page Users
	r.Get("/signup", userPage.RenderNew)
	r.Post("/signup", userPage.SaveNew)
	r.Get("/login", userPage.RenderLogin)
	r.Post("/login", userPage.CreateSession)
	r.Post("/logout", userPage.DeleteSession)

	// static files
	fileServer := http.FileServer(http.Dir("./static/"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// serve
	fmt.Println("Listening on http://127.0.0.1:8000/")
	http.ListenAndServe(":8000", r)
}
