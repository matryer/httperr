package httperr_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	is.Equal(httperr.Body(err), nil)
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
	is.Equal(string(httperr.Body(err)), "invalid request\n")
}

func TestTruncate(t *testing.T) {
	is := is.New(t)
	const (
		// truncateAfter is the maximum length of the body to include.
		truncateAfter = 50
		// truncatePadding is the status code, colon and dots etc
		truncatePadding = 8
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, strings.Repeat("b", truncateAfter*2)+"\n")
		is.NoErr(err)
	}))
	defer srv.Close()
	client := http.Client{Timeout: 1 * time.Second}
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	is.NoErr(err) // http.NewRequest
	_, err = httperr.Check(client.Do(req))
	is.True(err != nil)
	t.Log(err.Error())
	is.Equal(len(err.Error()), truncateAfter+truncatePadding)
}

func TestTemporary(t *testing.T) {
	is := is.New(t)
	status := http.StatusInternalServerError
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
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

	normalError := errors.New("some other error")
	is.Equal(httperr.Temporary(normalError), false)
}

func TestErrFromClient(t *testing.T) {
	is := is.New(t)
	resperr := errors.New("response error")
	_, err := httperr.Check(nil, resperr)
	is.Equal(err, resperr)
}

func TestErrReading(t *testing.T) {
	is := is.New(t)
	readerr := errors.New("read error")
	resp := &http.Response{
		Body:       errReader{err: readerr},
		StatusCode: http.StatusBadRequest,
	}
	_, err := httperr.Check(resp, nil)
	is.True(err != nil)
	is.Equal(err.Error(), `400: read error`)
}

type errReader struct {
	err error
}

func (e errReader) Read(b []byte) (int, error) {
	return 0, e.err
}

func (errReader) Close() error {
	return nil
}
