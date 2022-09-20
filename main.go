package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"git.sr.ht/~sirodoht/lakehousewiki/document"
	"git.sr.ht/~sirodoht/lakehousewiki/user"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", databaseUrl)
	if err != nil {
		panic(err)
	}

	// Instantiate stores
	documentStore := document.NewSQLStore(db)
	userStore := user.NewSQLStore(db)

	// Instantiate APIs
	documentApi := document.NewAPI(documentStore)
	userApi := user.NewAPI(userStore)

	// Instantiate Pages
	userPage := user.NewPage()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t, err := template.ParseFiles("templates/index.html")
		if err != nil {
			panic(err)
		}
		t.Execute(w, nil)
	})

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
	r.Get("/login", userPage.Render)
	r.Post("/login", userPage.Form)

	// Server
	fmt.Println("Listening on http://127.0.0.1:8000/")
	http.ListenAndServe(":8000", r)
}
