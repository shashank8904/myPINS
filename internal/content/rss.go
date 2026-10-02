package content

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RSSFetcher fetches and parses feed URLs.
// It supports both RSS 2.0 and Atom 1.0 formats, which are the two most
// common syndication formats used by technology blogs and news sources.
type RSSFetcher struct {
	client *http.Client
}

func NewRSSFetcher() *RSSFetcher {
	return &RSSFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// -------------------------------------------------------------------------
// RSS 2.0 structures
// -------------------------------------------------------------------------

// rss20Feed models the outer envelope of an RSS 2.0 document.
// Root element: <rss version="2.0"><channel>...</channel></rss>
type rss20Feed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []rss20Item `xml:"item"`
	} `xml:"channel"`
}

type rss20Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	PubDate     string `xml:"pubDate"`
}

// -------------------------------------------------------------------------
// Atom 1.0 structures
// -------------------------------------------------------------------------

// atomFeed models the outer envelope of an Atom 1.0 document.
// Root element: <feed xmlns="http://www.w3.org/2005/Atom">...</feed>
//
// Key structural differences from RSS 2.0:
//   - Entries use <entry> not <item>.
//   - The link URL is stored in an href *attribute*, not element text.
//   - Author is a nested struct: <author><name>...</name></author>.
//   - Publication time uses <published> in RFC 3339 format.
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	Author    struct {
		Name string `xml:"name"`
	} `xml:"author"`
}

// atomLink represents an Atom <link> element whose URL is an attribute.
// An entry may carry multiple links (e.g. "alternate" for the article page
// and "self" for the entry's own feed URL). We prefer "alternate".
type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// alternateHref returns the href of the "alternate" link, falling back to
// the first link found if no "alternate" relation is present.
func (e *atomEntry) alternateHref() string {
	for _, l := range e.Links {
		if l.Rel == "alternate" {
			return l.Href
		}
	}
	if len(e.Links) > 0 {
		return e.Links[0].Href
	}
	return ""
}

// -------------------------------------------------------------------------
// Date parsing helpers
// -------------------------------------------------------------------------

// rss20Formats lists the date layouts commonly found in RSS 2.0 pubDate fields.
// RSS 2.0 specifies RFC 1123 but many feeds use variations of it.
var rss20Formats = []string{
	time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 MST",
}

func parseRSS20Date(s string) time.Time {
	for _, layout := range rss20Formats {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// parseAtomDate parses an Atom <published> or <updated> timestamp.
// Atom mandates RFC 3339 (a profile of ISO 8601), which time.RFC3339 covers.
func parseAtomDate(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// -------------------------------------------------------------------------
// Fetch
// -------------------------------------------------------------------------

// Fetch downloads and parses the feed at the given URL.
//
// Auto-detection strategy:
// We read the entire response body into a buffer so we can decode it once
// to detect the root element name (either "rss" or "feed"), then decode it
// again with the correct struct. The body is small (a feed), so buffering is
// acceptable. This avoids third-party dependencies.
func (f *RSSFetcher) Fetch(url string) ([]RSSItem, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch feed %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed %s returned status %d", url, resp.StatusCode)
	}

	// Buffer the body so we can decode it twice: once to sniff the root
	// element, and once with the correct typed struct.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read feed body %s: %w", url, err)
	}

	// Sniff the root XML element name.
	var root struct {
		XMLName xml.Name
	}
	if err := xml.NewDecoder(bytes.NewReader(body)).Decode(&root); err != nil {
		return nil, fmt.Errorf("detect feed format %s: %w", url, err)
	}

	switch root.XMLName.Local {
	case "rss":
		return parseRSS20(body)
	case "feed":
		return parseAtom(body)
	default:
		return nil, fmt.Errorf("unsupported feed format %q at %s", root.XMLName.Local, url)
	}
}

// parseRSS20 decodes an RSS 2.0 body and returns RSSItems.
func parseRSS20(body []byte) ([]RSSItem, error) {
	var feed rss20Feed
	if err := xml.NewDecoder(bytes.NewReader(body)).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode RSS 2.0 feed: %w", err)
	}

	items := make([]RSSItem, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		rssItem := RSSItem{
			Title:       item.Title,
			URL:         item.Link,
			Author:      item.Author,
			Description: item.Description,
			PublishedAt: parseRSS20Date(item.PubDate),
		}
		items = append(items, rssItem)
	}
	return items, nil
}

// parseAtom decodes an Atom 1.0 body and returns RSSItems.
// We reuse RSSItem as the common intermediate representation regardless of
// source format — the field names describe the data, not the source format.
func parseAtom(body []byte) ([]RSSItem, error) {
	var feed atomFeed
	if err := xml.NewDecoder(bytes.NewReader(body)).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode Atom feed: %w", err)
	}

	items := make([]RSSItem, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		// Prefer <published>; fall back to <updated> if <published> is absent.
		pubStr := entry.Published
		if pubStr == "" {
			pubStr = entry.Updated
		}

		rssItem := RSSItem{
			Title:       entry.Title,
			URL:         entry.alternateHref(),
			Author:      entry.Author.Name,
			Description: entry.Summary,
			PublishedAt: parseAtomDate(pubStr),
		}
		items = append(items, rssItem)
	}
	return items, nil
}
