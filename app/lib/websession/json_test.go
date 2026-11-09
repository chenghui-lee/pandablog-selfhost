package websession_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chenghui-lee/pandablog-selfhost/app/lib/datastorage"
	"github.com/chenghui-lee/pandablog-selfhost/app/lib/websession"
)

func TestNewJSONSession(t *testing.T) {
	// Use local filesytem when developing.
	f := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(f, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create file %q: %v", f, err)
	}
	ss := datastorage.NewLocalStorage(f)

	// Set up the session storage provider.
	secretkey := "82a18fbbfed2694bb15d512a70c53b1a088e669966918d3d474564b2ac44349b"
	en := websession.NewEncryptedStorage(secretkey)
	store, err := websession.NewJSONSession(ss, en)
	if err != nil {
		t.Fatalf("Failed to create json session: %v", err)
	}

	token := "abc"
	data := "hello"
	expiry := time.Now().Add(time.Hour)

	if err := store.Commit(token, []byte(data), expiry); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	b, exists, err := store.Find(token)
	if err != nil {
		t.Fatalf("store.Find failed: %v", err)
	}
	if !exists {
		t.Error("store.Find returned false on exists")
	}
	if got, want := string(b), data; got != data {
		t.Errorf("store.Find got %q want %q", got, want)
	}

	if err := store.Delete(token); err != nil {
		t.Fatalf("store.Delete failed: %v", err)
	}

	_, exists, err = store.Find(token)
	if err != nil {
		t.Fatalf("store.Find failed: %v", err)
	}
	if exists {
		t.Errorf("store.Find returned true on exits")
	}
}

func TestJSONSessionExpiry(t *testing.T) {
	f := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(f, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create file %q: %v", f, err)
	}
	ss := datastorage.NewLocalStorage(f)

	secretkey := "82a18fbbfed2694bb15d512a70c53b1a088e669966918d3d474564b2ac44349b"
	en := websession.NewEncryptedStorage(secretkey)
	store, err := websession.NewJSONSession(ss, en)
	if err != nil {
		t.Fatalf("Failed to create json session: %v", err)
	}

	// An expired session must not be returned by Find.
	if err := store.Commit("expired", []byte("old"), time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("Failed to commit expired session: %v", err)
	}
	_, exists, err := store.Find("expired")
	if err != nil {
		t.Fatalf("store.Find failed: %v", err)
	}
	if exists {
		t.Error("store.Find returned true for an expired session")
	}

	// A valid session must still be returned.
	if err := store.Commit("valid", []byte("new"), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Failed to commit valid session: %v", err)
	}
	_, exists, err = store.Find("valid")
	if err != nil {
		t.Fatalf("store.Find failed: %v", err)
	}
	if !exists {
		t.Error("store.Find returned false for a valid session")
	}

	// Committing another session must purge expired records.
	if err := store.Commit("valid2", []byte("new"), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Failed to commit second valid session: %v", err)
	}
	sd := new(websession.SessionDatabase)
	if err := sd.Load(ss, en); err != nil {
		t.Fatalf("Failed to load session database: %v", err)
	}
	if _, ok := sd.Records["expired"]; ok {
		t.Error("expired session was not purged from the store")
	}
	if _, ok := sd.Records["valid"]; !ok {
		t.Error("valid session was removed from the store")
	}
}
