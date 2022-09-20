package document

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAPI_GetAllHandler(t *testing.T) {
	mockStore := NewMockStore(gomock.NewController(t))
	api := NewAPI(mockStore)

	expRes := []*Document{{
		ID:    1,
		Title: "title",
	}}

	mockStore.EXPECT().
		GetAll(gomock.Any()).
		Return(expRes, nil)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	api.GetAllHandler(rr, req)

	actRes := []*Document{}
	err = json.Unmarshal(rr.Body.Bytes(), &actRes)
	require.NoError(t, err)

	require.Equal(t, expRes, actRes)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAPI_GetAllHandler_Error(t *testing.T) {
	mockStore := NewMockStore(gomock.NewController(t))
	api := NewAPI(mockStore)

	mockStore.EXPECT().
		GetAll(gomock.Any()).
		Return(nil, sql.ErrNoRows)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	api.GetAllHandler(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}
