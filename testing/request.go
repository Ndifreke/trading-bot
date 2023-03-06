package testing

import (
	"net/http"
	"net/http/httptest"
)

func TestRequest(handler func(r *http.Request) string) (*http.Response, error) {
	req := httptest.NewRequest("", "", nil)
	w := httptest.NewRecorder()
	handle := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(handler(r)))
	}
	handle(w, req)
	return req.Response, nil
}
