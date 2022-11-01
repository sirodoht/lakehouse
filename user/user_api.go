package user

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
	_, err = api.store.Insert(r.Context(), u)
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
		err = api.store.Update(r.Context(), id, "username", *rb.Username)
		if err != nil {
			panic(err)
		}
	}
	if rb.Email != nil {
		err = api.store.Update(r.Context(), id, "email", *rb.Email)
		if err != nil {
			panic(err)
		}
	}
	w.WriteHeader(http.StatusOK)
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

	user, err := api.store.GetOne(r.Context(), id)
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
