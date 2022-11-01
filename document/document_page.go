package document

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"git.sr.ht/~sirodoht/lakehousewiki/eliot"

	"github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
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
	// parse url doc id
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get document from database
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

	// compile markdown to html
	unsafeHTML := blackfriday.Run([]byte(doc.Body))
	bodyHTML := bluemonday.UGCPolicy().SanitizeBytes(unsafeHTML)

	// respond
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"templates/layout.html",
		"templates/document.html",
	)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(eliot.KeyIsAuthenticated),
		"Username":        r.Context().Value(eliot.KeyUsername),
		"Document":        doc,
		"BodyHTML":        template.HTML(bodyHTML),
	})
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
	t, err := template.ParseFiles(
		"templates/layout.html",
		"templates/document_list.html",
	)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(eliot.KeyIsAuthenticated),
		"Username":        r.Context().Value(eliot.KeyUsername),
		"DocumentList":    docs,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"templates/layout.html",
		"templates/document_new.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile doc new template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(eliot.KeyIsAuthenticated),
		"Username":        r.Context().Value(eliot.KeyUsername),
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveNew(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	type ReqBody struct {
		Title string
		Body  string
	}
	rb := &ReqBody{
		Title: title,
		Body:  body,
	}
	fmt.Printf("%+v", rb)

	if rb.Title == "" || rb.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now()
	d := &Document{
		Title:     rb.Title,
		Body:      rb.Body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := page.store.Insert(r.Context(), d)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/docs", http.StatusFound)
}

func (page *Page) RenderEdit(w http.ResponseWriter, r *http.Request) {
	// parse url doc id
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get doc based on url id
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

	// render
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"templates/layout.html",
		"templates/document_edit.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile doc edit template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(eliot.KeyIsAuthenticated),
		"Username":        r.Context().Value(eliot.KeyUsername),
		"Document":        doc,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveEdit(w http.ResponseWriter, r *http.Request) {
	// parse doc id from url
	idAsString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idAsString, 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// gather post form data
	var data struct {
		Title string
		Body  string
	}
	data.Title = r.FormValue("title")
	data.Body = r.FormValue("body")

	// validate data
	if data.Title == "" || data.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// write updated doc on database
	err = page.store.Update(r.Context(), id, "title", data.Title)
	if err != nil {
		panic(err)
	}
	err = page.store.Update(r.Context(), id, "body", data.Body)
	if err != nil {
		panic(err)
	}

	// respond
	http.Redirect(w, r, "/docs/"+idAsString, http.StatusFound)
}
