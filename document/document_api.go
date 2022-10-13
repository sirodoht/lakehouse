package document

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type API struct {
	store  Store
	logger *zap.Logger
}

func NewAPI(store Store) *API {
	return &API{
		store: store,
	}
}

func (api *API) InsertHandler(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		Title string
		Body  string
	}
	decoder := json.NewDecoder(r.Body)
	var rb ReqBody
	err := decoder.Decode(&rb)
	if err != nil {
		panic(err)
	}

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

	_, err = api.store.Insert(r.Context(), d)
	if err != nil {
		panic(err)
	}
}

func (api *API) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		api.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type ReqBody struct {
		Title *string "json:title"
		Body  *string "json:body"
	}
	var rb ReqBody
	err = json.Unmarshal(b, &rb)
	if err != nil {
		api.logger.With(
			zap.Error(err),
		).Error("failed to parse input data")
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if rb.Title != nil {
		api.store.Update(r.Context(), id, "title", *rb.Title)
	}
	if rb.Body != nil {
		api.store.Update(r.Context(), id, "body", *rb.Body)
	}
	w.WriteHeader(http.StatusOK)
}

func (api *API) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	docs, err := api.store.GetAll(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		api.logger.With(
			zap.Error(err),
		).Error("failed to get all documents")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
	}
	w.Write(res)
}

func (api *API) GetOneHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		api.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	doc, err := api.store.GetOne(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		api.logger.With(
			zap.Error(err),
		).Error("failed to get document")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
	}
	w.Write(res)
}
