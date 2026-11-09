package middleware

import (
	"net/http"
	"strings"
)

// DisallowAuth does not allow authenticated users to access the page.
func (c *Handler) DisallowAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If user is authenticated, don't allow them to access the page.
		if _, loggedIn := c.Sess.User(r); !loggedIn {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// isDashboardPath reports whether the request path is handled by the
// dashboard. The router (github.com/matryer/way) trims leading and trailing
// slashes from the path before splitting it into segments, so paths such as
// "//dashboard" and "///dashboard/posts" are routed to the dashboard even
// though they do not have the "/dashboard" prefix. Mirror that normalization
// here so the ACL cannot be bypassed with extra slashes.
func isDashboardPath(p string) bool {
	segs := strings.Split(strings.Trim(p, "/"), "/")
	return len(segs) > 0 && segs[0] == "dashboard"
}

// DisallowAnon does not allow anonymous users to access the page.
func (c *Handler) DisallowAnon(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't allow anon users to access the dashboard.
		if isDashboardPath(r.URL.Path) {
			// If user is not authenticated, don't allow them to access the page.
			if _, loggedIn := c.Sess.User(r); !loggedIn {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
