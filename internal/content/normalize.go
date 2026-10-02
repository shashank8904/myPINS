package content

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// NormalizedItem is the common intermediate representation for all content
// entering the system, regardless of its original source format.
//
// Every source type (RSS, Atom, GitHub, YouTube, etc.) must produce
// NormalizedItem values. This creates a single funnel point where all
// external content enters the system in the same shape before being
// validated, deduplicated, and persisted.
//
// Flow:
//
//	SourceFormat (e.g. RSSItem) → NormalizedItem → ContentItem → Database
type NormalizedItem struct {
	SourceID    int64
	Title       string
	URL         string // will be normalized for dedup
	Author      string
	Summary     string
	Body        string
	PublishedAt time.Time
	ContentType string // "article", "github_release", etc.
}

// trackingParams are URL query parameters commonly used for marketing
// attribution. They do not change the content and should be stripped
// to prevent the same article from appearing as multiple items.
var trackingParams = map[string]bool{
	"utm_source":   true,
	"utm_medium":   true,
	"utm_campaign": true,
	"utm_content":  true,
	"utm_term":     true,
}

// NormalizeURL applies conservative URL normalization to reduce trivial
// duplicates without risk of breaking valid URLs.
//
// What it does:
//   - Lowercase scheme and host (RFC 3986 §3.1, §3.2.2)
//   - Remove URL fragments (#...)
//   - Remove trailing slash (except bare domain root "/")
//   - Remove common tracking query parameters (utm_*)
//
// What it does NOT do:
//   - Rewrite paths (case-sensitive on most servers)
//   - Remove unknown query parameters
//   - Follow redirects
//   - Resolve relative URLs
//
// If the URL cannot be parsed, it is returned unchanged.
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Lowercase scheme and host per RFC 3986.
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove fragment — it refers to a client-side anchor, not different content.
	u.Fragment = ""

	// Remove trailing slash (except the root path "/").
	// https://example.com/post/ and https://example.com/post are the same page
	// on most servers.
	if len(u.Path) > 1 && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimRight(u.Path, "/")
	}

	// Remove tracking query parameters.
	if u.RawQuery != "" {
		q := u.Query()
		for param := range trackingParams {
			q.Del(param)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// Validate checks whether a NormalizedItem has enough information to be
// useful. Returns nil if valid, or a descriptive error explaining why
// the item was rejected.
//
// Items that fail validation are logged and counted (IngestResult.Filtered)
// but do not stop processing of other items.
func (n *NormalizedItem) Validate() error {
	if strings.TrimSpace(n.Title) == "" {
		return fmt.Errorf("empty title")
	}
	if strings.TrimSpace(n.URL) == "" {
		return fmt.Errorf("empty URL")
	}
	if _, err := url.Parse(n.URL); err != nil {
		return fmt.Errorf("invalid URL %q: %w", n.URL, err)
	}
	return nil
}

// normalizedToContentItem converts a validated NormalizedItem into a
// ContentItem ready for database persistence.
//
// Optional fields (Author, Summary, Body, PublishedAt) are stored as
// pointers: nil means the source did not provide the value, which maps
// to NULL in PostgreSQL.
func normalizedToContentItem(item NormalizedItem) ContentItem {
	ci := ContentItem{
		SourceID:    item.SourceID,
		Title:       item.Title,
		URL:         item.URL,
		ContentType: item.ContentType,
	}

	if item.Author != "" {
		ci.Author = &item.Author
	}
	if item.Summary != "" {
		ci.Summary = &item.Summary
	}
	if item.Body != "" {
		ci.Body = &item.Body
	}
	if !item.PublishedAt.IsZero() {
		ci.PublishedAt = &item.PublishedAt
	}

	return ci
}
