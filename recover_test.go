package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecover(t *testing.T) {
	router := NewRouter()

	router.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
		WriteError(w, http.StatusInternalServerError, err.Error())
	})
	router.Use(Recover())

	router.HandleFunc("/recover", func(w http.ResponseWriter, r *http.Request) {
		panic("wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/recover", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !(rec.Result().StatusCode == http.StatusInternalServerError && rec.Body.String() == "wrong") {
		t.Fatal(rec.Result().StatusCode, rec.Body.String())
	}

}
