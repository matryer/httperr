package httperr_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/httperr"
	"github.com/matryer/is"
)

func TestSuccess(t *testing.T) {
	is := is.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	client := http.Client{Timeout: 1 * time.Second}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	is.NoErr(err) // http.NewRequest
	resp, err := httperr.Check(client.Do(req))
	is.NoErr(err) // httperr.Check
	is.Equal(resp.StatusCode, http.StatusOK)
}

func TestBadRequest(t *testing.T) {
	is := is.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "invalid request\n")
		is.NoErr(err)
	}))
	defer srv.Close()
	client := http.Client{Timeout: 1 * time.Second}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	is.NoErr(err) // http.NewRequest
	_, err = httperr.Check(client.Do(req))
	is.True(err != nil)
	is.Equal(err.Error(), "400: invalid request")
}

func TestTemporary(t *testing.T) {
	is := is.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, "try again later\n")
		is.NoErr(err)
	}))
	defer srv.Close()
	client := http.Client{Timeout: 1 * time.Second}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	is.NoErr(err) // http.NewRequest
	_, err = httperr.Check(client.Do(req))
	is.True(err != nil)
	is.Equal(err.Error(), "500: try again later")
	is.Equal(httperr.Temporary(err), true)
}
