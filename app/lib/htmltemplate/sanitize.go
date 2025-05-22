package htmltemplate

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	blackfriday "github.com/russross/blackfriday/v2"
	"golang.org/x/net/html"
)

// SanitizedHTML returns a sanitized content html.
func (te *Engine) SanitizedHTML(content string, isPost bool) []byte {
	// Ensure unit line endings are used when pulling out of JSON.
	markdownWithUnixLineEndings := strings.Replace(content, "\r\n", "\n", -1)
	htmlCode := blackfriday.Run([]byte(markdownWithUnixLineEndings))

	// Sanitize by removing HTML if true.
	if !te.allowUnsafeHTML {
		htmlCode = bluemonday.UGCPolicy().SanitizeBytes(htmlCode)
	}

	// Add table of contents if this is a blog post
	if isPost {
		// Add IDs to headings if they don't have them
		htmlCode = addIDsToHeadings(htmlCode)

		// Generate table of contents
		toc := generateTableOfContents(htmlCode)

		// Add back to top button
		backToTop := []byte(`<div id="back-to-top" title="Back to Top">↑</div>`)

		// Combine TOC + content + back to top button
		if len(toc) > 0 {
			htmlCode = append(toc, htmlCode...)
		}
		htmlCode = append(htmlCode, backToTop...)
	}

	return htmlCode
}

// sanitizedContent returns a sanitized content block or an error is one occurs.
func (te *Engine) sanitizedContent(t *template.Template, name, markdown string, isBlogPost bool) (*template.Template, error) {
	// Determine if we should generate TOC
	// Default to false if not specified
	addTOC := false
	if isBlogPost && name == "content" {
		addTOC = true
	}
	htmlCode := te.SanitizedHTML(markdown, addTOC)

	// Change delimiters temporarily so code samples can use Go blocks.
	safeContent := fmt.Sprintf(`[{[{define "%s"}]}]%s[{[{end}]}]`, name, htmlCode)
	t = t.Delims("[{[{", "}]}]")
	var err error
	t, err = t.Parse(safeContent)
	if err != nil {
		return nil, err
	}
	// Reset delimiters
	t = t.Delims("{{", "}}")
	return t, nil
}

// renderText extracts text content from an HTML node, ignoring nested tags
func renderText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	
	var result string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += renderText(c)
	}
	
	return strings.TrimSpace(result)
}

// slugify creates a URL-friendly slug from a string
func slugify(text string) string {
	// Convert to lowercase
	slug := strings.ToLower(text)
	// Replace non-alphanumeric characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

// generateTableOfContents parses HTML content and generates a table of contents
func generateTableOfContents(htmlContent []byte) []byte {
	// If no headings, return empty TOC
	if !bytes.Contains(htmlContent, []byte("<h")) {
		return []byte("")
	}
	
	// Parse the HTML content
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return []byte("")
	}
	
	// Define a heading structure
	type heading struct {
		level    int
		id       string
		title    string
		position int // Position in the document
	}
	
	var headings []heading
	var posCounter int
	
	// Walk the HTML tree to find headings
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Check if this is a heading element (h1-h6)
			if len(n.Data) == 2 && n.Data[0] == 'h' && n.Data[1] >= '1' && n.Data[1] <= '6' {
				level := int(n.Data[1] - '0') // Convert '1'..'6' to 1..6
				
				// Look for an ID attribute
				var id string
				for _, attr := range n.Attr {
					if attr.Key == "id" {
						id = attr.Val
						break
					}
				}
				
				// If no ID found, generate one from the heading text
				if id == "" {
					id = slugify(renderText(n))
				}
				
				// Get the heading text content
				title := renderText(n)
				
				// Add to our headings slice
				headings = append(headings, heading{
					level:    level,
					id:       id,
					title:    title,
					position: posCounter,
				})
				
				posCounter++
			}
		}
		
		// Recursively process child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	
	// Start the walk from the document root
	walk(doc)
	
	if len(headings) == 0 {
		return []byte("")
	}
	
	// Generate TOC HTML
	var toc bytes.Buffer
	toc.WriteString(`<div class="table-of-contents">`)
	toc.WriteString(`<div class="toc-header">`)
	toc.WriteString(`<h2>Table of Contents</h2>`)
	toc.WriteString(`<button id="toc-toggle" aria-expanded="true" aria-controls="toc-list">−</button>`)
	toc.WriteString(`</div>`)
	toc.WriteString(`<div id="toc-list" class="toc-content">`)
	toc.WriteString(`<ul>`)
	
	for _, h := range headings {
		// Remove any HTML tags from the title
		titleText := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(h.title, "")
		
		// Add indentation based on heading level
		indent := ""
		if h.level > 1 {
			indent = fmt.Sprintf(`style="margin-left: %dem;"`, (h.level-1)*2)
		}
		
		toc.WriteString(fmt.Sprintf(`<li %s><a href="#%s">%s</a></li>`, indent, h.id, titleText))
	}
	
	toc.WriteString(`</ul>`)
	toc.WriteString(`</div>`)
	toc.WriteString(`</div>`)
	
	return toc.Bytes()
}

// addIDsToHeadings adds IDs to headings that don't have them
func addIDsToHeadings(htmlContent []byte) []byte {
	// Process each heading level separately since Go regex doesn't support backreferences
	for level := 1; level <= 6; level++ {
		// Find all h{level} headings
		headingPattern := fmt.Sprintf(`<h%d([^>]*)>(.*?)</h%d>`, level, level)
		allHeadingsRe := regexp.MustCompile(headingPattern)
		
		// Find headings with IDs
		headingsWithIDRe := regexp.MustCompile(`id=["']([^"']+)["']`)
		
		// Replace each heading with a version that has an ID
		htmlContent = allHeadingsRe.ReplaceAllFunc(htmlContent, func(match []byte) []byte {
			// Skip if it already has an ID
			if headingsWithIDRe.Match(match) {
				return match
			}
			
			submatches := allHeadingsRe.FindSubmatch(match)
			attrs := string(submatches[1])
			content := string(submatches[2])
			
			// Remove any HTML tags for the ID
			plainContent := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(content, "")
			
			// Create a slug from the content
			slug := strings.ToLower(plainContent)
			slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
			slug = strings.Trim(slug, "-")
			
			// Create a new heading with the ID
			return []byte(fmt.Sprintf(`<h%d id="%s"%s>%s</h%d>`, level, slug, attrs, content, level))
		})
	}
	
	return htmlContent
}
