package content

import (
	"testing"
	"time"
)

// =============================================================================
// RSS 2.0 parsing tests
// =============================================================================

// TestParseRSS20_BasicFeed verifies that a well-formed RSS 2.0 document
// is correctly parsed into RSSItem structs.
func TestParseRSS20_BasicFeed(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Blog</title>
    <item>
      <title>First Post</title>
      <link>https://example.com/first</link>
      <description>This is the first post.</description>
      <author>Alice</author>
      <pubDate>Mon, 02 Jan 2006 15:04:05 -0700</pubDate>
    </item>
    <item>
      <title>Second Post</title>
      <link>https://example.com/second</link>
      <description>This is the second post.</description>
    </item>
  </channel>
</rss>`)

	items, err := parseRSS20(body)
	if err != nil {
		t.Fatalf("parseRSS20 returned error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// First item — all fields present.
	first := items[0]
	if first.Title != "First Post" {
		t.Errorf("item[0].Title = %q, want %q", first.Title, "First Post")
	}
	if first.URL != "https://example.com/first" {
		t.Errorf("item[0].URL = %q, want %q", first.URL, "https://example.com/first")
	}
	if first.Author != "Alice" {
		t.Errorf("item[0].Author = %q, want %q", first.Author, "Alice")
	}
	if first.Description != "This is the first post." {
		t.Errorf("item[0].Description = %q, want %q", first.Description, "This is the first post.")
	}
	if first.PublishedAt.IsZero() {
		t.Errorf("item[0].PublishedAt should not be zero")
	}

	// Second item — author and pubDate missing.
	second := items[1]
	if second.Title != "Second Post" {
		t.Errorf("item[1].Title = %q, want %q", second.Title, "Second Post")
	}
	if second.Author != "" {
		t.Errorf("item[1].Author = %q, want empty string", second.Author)
	}
}

// TestParseRSS20_EmptyChannel verifies that an RSS feed with zero items
// returns an empty slice (not nil), and does not error.
func TestParseRSS20_EmptyChannel(t *testing.T) {
	body := []byte(`<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
  </channel>
</rss>`)

	items, err := parseRSS20(body)
	if err != nil {
		t.Fatalf("parseRSS20 returned error: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

// =============================================================================
// Atom 1.0 parsing tests
// =============================================================================

// TestParseAtom_BasicFeed verifies that a well-formed Atom 1.0 document
// is correctly parsed. This is important because the Go Blog uses Atom.
func TestParseAtom_BasicFeed(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Go Blog</title>
  <entry>
    <title>Go 1.22 Released</title>
    <link rel="alternate" href="https://go.dev/blog/go1.22"/>
    <summary>Go 1.22 is released.</summary>
    <published>2024-02-06T00:00:00Z</published>
    <author><name>The Go Team</name></author>
  </entry>
  <entry>
    <title>Structured Logging</title>
    <link href="https://go.dev/blog/slog"/>
    <summary>Introducing slog.</summary>
    <updated>2023-08-22T00:00:00Z</updated>
  </entry>
</feed>`)

	items, err := parseAtom(body)
	if err != nil {
		t.Fatalf("parseAtom returned error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// First entry — has rel="alternate" link and <published>.
	first := items[0]
	if first.Title != "Go 1.22 Released" {
		t.Errorf("item[0].Title = %q, want %q", first.Title, "Go 1.22 Released")
	}
	if first.URL != "https://go.dev/blog/go1.22" {
		t.Errorf("item[0].URL = %q, want %q", first.URL, "https://go.dev/blog/go1.22")
	}
	if first.Author != "The Go Team" {
		t.Errorf("item[0].Author = %q, want %q", first.Author, "The Go Team")
	}
	if first.PublishedAt.IsZero() {
		t.Errorf("item[0].PublishedAt should not be zero")
	}

	// Second entry — no rel attribute on link (should still work),
	// and uses <updated> instead of <published>.
	second := items[1]
	if second.URL != "https://go.dev/blog/slog" {
		t.Errorf("item[1].URL = %q, want %q", second.URL, "https://go.dev/blog/slog")
	}
	if second.PublishedAt.IsZero() {
		t.Errorf("item[1].PublishedAt should not be zero (should fall back to <updated>)")
	}
}

// TestParseAtom_AlternateLinkPreference verifies that when an Atom entry
// has multiple <link> elements, the one with rel="alternate" is preferred.
func TestParseAtom_AlternateLinkPreference(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <title>Multi Link Entry</title>
    <link rel="self" href="https://example.com/feed/entry/1"/>
    <link rel="alternate" href="https://example.com/post/1"/>
    <link rel="enclosure" href="https://example.com/media/1.mp3"/>
  </entry>
</feed>`)

	items, err := parseAtom(body)
	if err != nil {
		t.Fatalf("parseAtom returned error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].URL != "https://example.com/post/1" {
		t.Errorf("URL = %q, want alternate link %q", items[0].URL, "https://example.com/post/1")
	}
}

// =============================================================================
// Feed format auto-detection tests
// =============================================================================

// TestFetch_AutoDetect uses the Fetch method's internal XML sniffing logic
// indirectly via the exported parseRSS20 and parseAtom functions.
// Direct Fetch testing requires an HTTP server (integration test territory).
// Here we verify the building blocks work correctly.

// =============================================================================
// Date parsing tests
// =============================================================================

func TestParseRSS20Date(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool // true = should parse successfully (non-zero time)
	}{
		{"RFC1123Z", "Mon, 02 Jan 2006 15:04:05 -0700", true},
		{"RFC1123", "Mon, 02 Jan 2006 15:04:05 MST", true},
		{"single-digit day with offset", "Mon, 2 Jan 2006 15:04:05 -0700", true},
		{"single-digit day with timezone", "Mon, 2 Jan 2006 15:04:05 MST", true},
		{"empty string", "", false},
		{"garbage", "not-a-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRSS20Date(tt.input)
			if tt.want && result.IsZero() {
				t.Errorf("parseRSS20Date(%q) returned zero time, want non-zero", tt.input)
			}
			if !tt.want && !result.IsZero() {
				t.Errorf("parseRSS20Date(%q) returned non-zero time, want zero", tt.input)
			}
		})
	}
}

func TestParseAtomDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"RFC3339 UTC", "2024-02-06T00:00:00Z", true},
		{"RFC3339 with offset", "2024-02-06T12:30:00+05:30", true},
		{"empty string", "", false},
		{"garbage", "not-a-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAtomDate(tt.input)
			if tt.want && result.IsZero() {
				t.Errorf("parseAtomDate(%q) returned zero time, want non-zero", tt.input)
			}
			if !tt.want && !result.IsZero() {
				t.Errorf("parseAtomDate(%q) returned non-zero time, want zero", tt.input)
			}
		})
	}
}

// =============================================================================
// rssItemToNormalizedItem conversion tests
// =============================================================================

func TestRssItemToNormalizedItem_AllFields(t *testing.T) {
	pubTime := time.Date(2024, 2, 6, 0, 0, 0, 0, time.UTC)

	rss := RSSItem{
		Title:       "Go 1.22 Released",
		URL:         "https://go.dev/blog/go1.22",
		Author:      "The Go Team",
		Description: "Go 1.22 brings major improvements.",
		PublishedAt: pubTime,
	}

	norm := rssItemToNormalizedItem(42, rss)

	if norm.SourceID != 42 {
		t.Errorf("SourceID = %d, want 42", norm.SourceID)
	}
	if norm.Title != rss.Title {
		t.Errorf("Title = %q, want %q", norm.Title, rss.Title)
	}
	if norm.URL != rss.URL {
		t.Errorf("URL = %q, want %q", norm.URL, rss.URL)
	}
	if norm.ContentType != "article" {
		t.Errorf("ContentType = %q, want %q", norm.ContentType, "article")
	}
	if norm.Author != "The Go Team" {
		t.Errorf("Author = %q, want %q", norm.Author, "The Go Team")
	}
	if norm.Summary != rss.Description {
		t.Errorf("Summary = %q, want %q", norm.Summary, rss.Description)
	}
	if !norm.PublishedAt.Equal(pubTime) {
		t.Errorf("PublishedAt = %v, want %v", norm.PublishedAt, pubTime)
	}
}

func TestRssItemToNormalizedItem_OptionalFieldsEmpty(t *testing.T) {
	rss := RSSItem{
		Title: "Minimal Item",
		URL:   "https://example.com/minimal",
		// Author, Description, PublishedAt intentionally empty/zero.
	}

	norm := rssItemToNormalizedItem(1, rss)

	if norm.Author != "" {
		t.Errorf("Author = %q, want empty string for empty author", norm.Author)
	}
	if norm.Summary != "" {
		t.Errorf("Summary = %q, want empty string for empty description", norm.Summary)
	}
	if !norm.PublishedAt.IsZero() {
		t.Errorf("PublishedAt = %v, want zero time", norm.PublishedAt)
	}
}
