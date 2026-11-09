package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"

	"github.com/chenghui-lee/pandablog-selfhost/app/lib/websession"
	"github.com/chenghui-lee/pandablog-selfhost/app/middleware"
)

func newAnonHandler(t *testing.T) (http.Handler, *websession.Session, *scs.SessionManager) {
	t.Helper()

	manager := scs.New()
	sess := websession.New("session", manager)
	h := middleware.NewHandler(nil, sess, nil, "", "")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return h.DisallowAnon(next), sess, manager
}

func TestDisallowAnonNormalizesSlashes(t *testing.T) {
	mw, _, manager := newAnonHandler(t)

	ctx, err := manager.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("manager.Load failed: %v", err)
	}

	tests := []struct {
		path string
		want int
	}{
		// The router trims slashes, so these all resolve to the dashboard and
		// must be protected.
		{"/dashboard", http.StatusFound},
		{"/dashboard/", http.StatusFound},
		{"//dashboard", http.StatusFound},
		{"///dashboard", http.StatusFound},
		{"//dashboard/posts", http.StatusFound},
		{"///dashboard/posts/new", http.StatusFound},
		// A post whose slug starts with "dashboard" is not the dashboard.
		{"/dashboardfoo", http.StatusOK},
		{"/blog", http.StatusOK},
	}

	for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, tc.path, nil).WithContext(ctx)
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, r)
		if got := w.Code; got != tc.want {
			t.Errorf("path %q: status got %v want %v", tc.path, got, tc.want)
		}
	}
}

func TestDisallowAnonAllowsAuthenticatedUser(t *testing.T) {
	mw, sess, manager := newAnonHandler(t)

	ctx, err := manager.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("manager.Load failed: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "//dashboard/posts", nil).WithContext(ctx)
	sess.SetUser(r, "admin")

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("status got %v want %v", got, want)
	}
}
