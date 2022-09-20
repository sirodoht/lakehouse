package document

import (
	"database/sql"
	"html/template"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Page struct {
	store  Store
	logger *zap.Logger
}

func NewPage(store Store) *Page {
	return &Page{
		store: store,
	}
}

func (page *Page) RenderOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	doc, err := page.store.GetOne(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get document")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("templates/layout.html", "templates/document.html")
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, doc)
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderAll(w http.ResponseWriter, r *http.Request) {
	docs, err := page.store.GetAll(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get all documents")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("templates/layout.html", "templates/document_list.html")
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, docs)
	if err != nil {
		panic(err)
	}
}
