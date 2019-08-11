# httperr

HTTP error wrapper that returns an error if the HTTP request failed (i.e. 404, 500, etc.) as well as
if any network issues occurred.

This is useful for cases when you don't care why an HTTP request failed, and would like to treat 
network errors and API errors once.

## Usage

```go
req, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
if err != nil {
	return errors.Wrap(err, "new request")
}
resp, err := httperr.Check(client.Do(req))
if err != nil {
	return errors.Wrap(err, "HTTP error")
}
defer resp.Body.Close()
// use resp
```
