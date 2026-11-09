package websession

import (
	"crypto/rand"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

// Sessionstorer reads and writes data to an object.
type Sessionstorer interface {
	Save([]byte) error
	Load() ([]byte, error)
}

// Session stores session level information
type Session struct {
	Name    string
	manager *scs.SessionManager
}

// New returns a new session cookie store.
func New(name string, manager *scs.SessionManager) *Session {
	return &Session{
		manager: manager,
		Name:    name,
	}
}

// RememberMe -
func (s *Session) RememberMe(r *http.Request, value bool) {
	s.manager.Cookie.Persist = value
}

// Logout -
func (s *Session) Logout(r *http.Request) {
	s.manager.Destroy(r.Context())
}

// RenewToken generates a new session token for the request. Call it after a
// successful login to prevent session fixation.
func (s *Session) RenewToken(r *http.Request) error {
	return s.manager.RenewToken(r.Context())
}

// User -
func (s *Session) User(r *http.Request) (string, bool) {
	u := s.manager.GetString(r.Context(), "user")
	return u, len(u) > 0
}

// SetUser -
func (s *Session) SetUser(r *http.Request, value string) {
	s.manager.Put(r.Context(), "user", value)
}

// String -
func (s *Session) String(r *http.Request, name string) string {
	return s.manager.GetString(r.Context(), name)
}

// SetString -
func (s *Session) SetString(r *http.Request, name string, value string) {
	s.manager.Put(r.Context(), name, value)
}

// SetCSRFForPath generates a CSRF token for the given path.
func (s *Session) SetCSRFForPath(r *http.Request, path string) string {
	token := generate(32)
	s.SetString(r, "csrf_"+path, token)
	return token
}

// SetCSRF generates a CSRF token for the request path.
func (s *Session) SetCSRF(r *http.Request) string {
	return s.SetCSRFForPath(r, r.URL.Path)
}

// CSRF -
func (s *Session) CSRF(r *http.Request) bool {
	token := r.FormValue("token")
	path := "csrf_" + r.URL.Path
	v := s.String(r, path)

	if len(v) > 0 {
		s.manager.Remove(r.Context(), path)
		if v == token && len(token) > 0 {
			return true
		}
	}

	return false
}

// Generate a token.
// Source: https://devpy.wordpress.com/2013/10/24/create-random-string-in-golang/
func generate(length int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Never silently fall back to a predictable all-zero token.
		panic("websession: failed to read random bytes: " + err.Error())
	}
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}

	return string(bytes)
}
