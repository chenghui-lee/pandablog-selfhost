package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/josephspurrier/polarbearblog/app/lib/datastorage"
	"github.com/josephspurrier/polarbearblog/app/lib/envdetect"
	"github.com/josephspurrier/polarbearblog/app/lib/htmltemplate"
	"github.com/josephspurrier/polarbearblog/app/lib/websession"
	"github.com/josephspurrier/polarbearblog/app/middleware"
	"github.com/josephspurrier/polarbearblog/app/route"
	"github.com/josephspurrier/polarbearblog/html"
)

var (
	storageSitePath    = "storage/site.json"
	storageSessionPath = "storage/session.bin"
	sessionName        = "session"
)

// Boot -
func Boot() (http.Handler, error) {
	// Set the storage and session environment variables.
	sitePath := os.Getenv("PBB_SITE_PATH")
	if len(sitePath) > 0 {
		storageSitePath = sitePath
	}

	sessionPath := os.Getenv("PBB_SESSION_PATH")
	if len(sessionPath) > 0 {
		storageSessionPath = sessionPath
	}

	sname := os.Getenv("PBB_SESSION_NAME")
	if len(sname) > 0 {
		sessionName = sname
	}

	// Get the environment variables.
	secretKey := os.Getenv("PBB_SESSION_KEY")
	if len(secretKey) == 0 {
		return nil, fmt.Errorf("environment variable missing: %v", "PBB_SESSION_KEY")
	}

	bucket := os.Getenv("PBB_GCP_BUCKET_NAME")
	if len(bucket) == 0 {
		return nil, fmt.Errorf("environment variable missing: %v", "PBB_GCP_BUCKET_NAME")
	}

	allowHTML, err := strconv.ParseBool(os.Getenv("PBB_ALLOW_HTML"))
	if err != nil {
		return nil, fmt.Errorf("environment variable not able to parse as bool: %v", "PBB_ALLOW_HTML")
	}

	// Create new store object with the defaults.
	var ds datastorage.Datastorer
	var ss websession.Sessionstorer

	if !envdetect.RunningLocalDev() {
		// Use Google when running in GCP.
		ds = datastorage.NewGCPStorage(bucket, storageSitePath)
		ss = datastorage.NewGCPStorage(bucket, storageSessionPath)
	} else {
		// Use local filesytem when developing.
		ds = datastorage.NewLocalStorage(storageSitePath)
		ss = datastorage.NewLocalStorage(storageSessionPath)
	}

	// Set up the data storage provider.
	storage, err := datastorage.New(ds)
	if err != nil {
		return nil, err
	}

	// Set up the session storage provider.
	en := websession.NewEncryptedStorage(secretKey)
	store, err := websession.NewJSONSession(ss, en)
	if err != nil {
		return nil, err
	}

	// Initialize a new session manager and configure the session lifetime.
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = false
	sessionManager.Store = store
	sess := websession.New(sessionName, sessionManager)

	// Set up the template engine.
	tm := html.NewTemplateManager(storage, sess)
	tmpl := htmltemplate.New(tm, allowHTML)

	// Setup the routes.
	c, err := route.Register(storage, sess, tmpl)
	if err != nil {
		return nil, err
	}

	// Set up the router and middleware.
	site, err := c.Storage.Site.Load(context.Background())
	if err != nil {
		return nil, err
	}
	var mw http.Handler
	mw = c.Router
	h := middleware.NewHandler(c.Render, c.Sess, c.Router, site.URL, site.Scheme)
	mw = h.Redirect(mw)
	mw = middleware.Head(mw)
	mw = h.DisallowAnon(mw)
	mw = sessionManager.LoadAndSave(mw)
	mw = middleware.Gzip(mw)
	mw = h.LogRequest(mw)

	return mw, nil
}
