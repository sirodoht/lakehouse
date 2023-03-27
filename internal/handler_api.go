package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	chi "github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type API struct {
	store  *SQLStore
	logger *zap.Logger
}

func NewHandlerAPI(store *SQLStore) *API {
	return &API{
		store: store,
	}
}

func (api *API) InsertUserHandler(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		Username string
		Email    string
	}
	decoder := json.NewDecoder(r.Body)
	var rb ReqBody
	err := decoder.Decode(&rb)
	if err != nil {
		panic(err)
	}

	if rb.Username == "" || rb.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now()
	u := &User{
		Username:  rb.Username,
		Email:     rb.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = api.store.InsertUser(r.Context(), u)
	if err != nil {
		panic(err)
	}
}

func (api *API) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		api.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type ReqBody struct {
		Username *string `json:"username"`
		Email    *string `json:"email"`
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
	if rb.Username != nil {
		err = api.store.UpdateUser(r.Context(), id, "username", *rb.Username)
		if err != nil {
			panic(err)
		}
	}
	if rb.Email != nil {
		err = api.store.UpdateUser(r.Context(), id, "email", *rb.Email)
		if err != nil {
			panic(err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (api *API) GetOneUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		api.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	user, err := api.store.GetOneUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		api.logger.With(
			zap.Error(err),
		).Error("failed to get user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		panic(err)
	}
	_, err = w.Write(res)
	if err != nil {
		panic(err)
	}
}

func (api *API) InsertDocumentHandler(w http.ResponseWriter, r *http.Request) {
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

	_, err = api.store.InsertDocument(r.Context(), d)
	if err != nil {
		panic(err)
	}
}

func (api *API) UpdateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		api.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type ReqBody struct {
		Title *string `json:"title"`
		Body  *string `json:"body"`
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
		err = api.store.UpdateDocument(r.Context(), id, "title", *rb.Title)
		if err != nil {
			panic(err)
		}
	}
	if rb.Body != nil {
		err = api.store.UpdateDocument(r.Context(), id, "body", *rb.Body)
		if err != nil {
			panic(err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (api *API) GetAllDocumentHandler(w http.ResponseWriter, r *http.Request) {
	docs, err := api.store.GetAllDocument(r.Context())
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
		panic(err)
	}
	_, err = w.Write(res)
	if err != nil {
		panic(err)
	}
}

func (api *API) GetOneDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		api.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	doc, err := api.store.GetOneDocument(r.Context(), id)
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
		panic(err)
	}
	_, err = w.Write(res)
	if err != nil {
		panic(err)
	}
}
