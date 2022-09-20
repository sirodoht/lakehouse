package user

import (
	"fmt"
	"html/template"
	"net/http"

	"go.uber.org/zap"
)

type Page struct {
	logger *zap.Logger
}

func NewPage() *Page {
	return &Page{}
}

func (page *Page) Render(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile login template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}

func (page *Page) Form(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	type ReqBody struct {
		Username string
		Password string
	}
	rb := &ReqBody{
		Username: username,
		Password: password,
	}
	fmt.Printf("%+v", rb)
}
