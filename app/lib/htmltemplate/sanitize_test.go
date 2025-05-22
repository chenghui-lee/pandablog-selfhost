package htmltemplate

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateTableOfContents_NoHeadings(t *testing.T) {
	html := []byte(`<p>No headings here</p>`)
	updated, toc := generateTableOfContents(html)
	if len(toc) != 0 {
		t.Errorf("Expected empty TOC, got: %s", toc)
	}
	if !bytes.Equal(updated, html) {
		t.Errorf("Expected HTML unchanged, got: %s", updated)
	}
}

func TestGenerateTableOfContents_SingleHeading(t *testing.T) {
	html := []byte(`<h1>Hello World</h1>`)
	updated, toc := generateTableOfContents(html)
	if !strings.Contains(string(toc), "Table of Contents") {
		t.Errorf("TOC missing header: %s", toc)
	}
	if !strings.Contains(string(toc), "#hello-world") {
		t.Errorf("TOC missing anchor link: %s", toc)
	}
	if !strings.Contains(string(updated), "id=\"hello-world\"") {
		t.Errorf("Heading missing ID: %s", updated)
	}
}

func TestGenerateTableOfContents_MultipleHeadings(t *testing.T) {
	html := []byte(`<h1>First</h1><h2>Second</h2><h3>Third</h3>`)
	updated, toc := generateTableOfContents(html)
	if !strings.Contains(string(toc), "#first") ||
		!strings.Contains(string(toc), "#second") ||
		!strings.Contains(string(toc), "#third") {
		t.Errorf("TOC missing anchors: %s", toc)
	}
	if !strings.Contains(string(updated), "id=\"first\"") ||
		!strings.Contains(string(updated), "id=\"second\"") ||
		!strings.Contains(string(updated), "id=\"third\"") {
		t.Errorf("Updated HTML missing heading IDs: %s", updated)
	}
}

func TestGenerateTableOfContents_HeadingsWithSpecialChars(t *testing.T) {
	html := []byte(`<h2>Go & Rust: A Comparison!</h2>`)
	updated, toc := generateTableOfContents(html)
	if !strings.Contains(string(toc), "#go-rust-a-comparison") {
		t.Errorf("TOC missing slugified anchor: %s", toc)
	}
	if !strings.Contains(string(updated), "id=\"go-rust-a-comparison\"") {
		t.Errorf("Heading missing slugified ID: %s", updated)
	}
}

func TestGenerateTableOfContents_ExistingIDs(t *testing.T) {
	html := []byte(`<h1 id="custom-id">Custom Heading</h1>`)
	updated, toc := generateTableOfContents(html)
	if !strings.Contains(string(toc), "#custom-id") {
		t.Errorf("TOC should use existing ID: %s", toc)
	}
	if !strings.Contains(string(updated), "id=\"custom-id\"") {
		t.Errorf("Heading should retain existing ID: %s", updated)
	}
}
