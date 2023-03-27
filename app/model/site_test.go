package model_test

import (
	"testing"

	"github.com/chenghui-lee/pandablog-selfhost/app/model"
)

func TestSiteURL(t *testing.T) {
	s := new(model.Site)
	s.Scheme = "http"
	s.URL = "localhost"
	if got, want := s.SiteURL(), "http://localhost"; got != want {
		t.Errorf("SiteURL() got %q want %q", got, want)
	}
}
