package route_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chenghui-lee/pandablog-selfhost/app/lib/router"
	"github.com/chenghui-lee/pandablog-selfhost/app/route"
)

func setupRouter() *router.Mux {
	// Set the handling of all responses.
	customServeHTTP := func(w http.ResponseWriter, r *http.Request, status int, err error) {}

	// Send all 404 to the customer handler.
	notFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customServeHTTP(w, r, http.StatusNotFound, nil)
	})

	// Set up the router.
	return router.New(customServeHTTP, notFound)
}

func TestXML(t *testing.T) {
	mux := setupRouter()

	// Create core app.
	c := &route.Core{}
	x := &route.XMLUtil{c}
	mux.Get("/robots.txt", x.Robots)
	r := httptest.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	b, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Errorf("Failed to read body: %v", err)
	}
	if got, want := string(b), "User-agent: *\nDisallow: /login/*\nDisallow: /login"; got != want {
		t.Errorf("body got %q want %q", got, want)
	}
}
