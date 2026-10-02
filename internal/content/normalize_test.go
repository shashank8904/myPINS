package content

import (
	"testing"
	"time"
)

// =============================================================================
// NormalizeURL tests
// =============================================================================

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase scheme and host",
			input:    "HTTPS://Example.COM/path",
			expected: "https://example.com/path",
		},
		{
			name:     "Remove URL fragments",
			input:    "https://example.com/post#comments",
			expected: "https://example.com/post",
		},
		{
			name:     "Remove trailing slash (except root)",
			input:    "https://example.com/post/",
			expected: "https://example.com/post",
		},
		{
			name:     "Keep root trailing slash",
			input:    "https://example.com/",
			expected: "https://example.com/",
		},
		{
			name:     "Remove utm_ tracking parameters",
			input:    "https://example.com/post?utm_source=twitter&utm_medium=social&utm_campaign=launch&utm_content=v1&utm_term=tech",
			expected: "https://example.com/post",
		},
		{
			name:     "Keep non-tracking parameters",
			input:    "https://example.com/post?utm_source=twitter&page=2&id=123",
			expected: "https://example.com/post?id=123&page=2",
		},
		{
			name:     "All normalizations combined",
			input:    "HTTP://EXAMPLE.COM/Path/To/Post/?utm_source=rss#section1",
			expected: "http://example.com/Path/To/Post?utm_source=rss", // NOTE: raw query parsing doesn't strip utm_source in this edge case without Decode/Encode, but we encode/decode properly in NormalizeURL. Wait, let's see. Encode sorts the keys. The output should be "http://example.com/Path/To/Post"
		},
		{
			name:     "Invalid URL returned as-is",
			input:    "://invalid-url",
			expected: "://invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fix the combined test expected value based on actual behavior
			if tt.name == "All normalizations combined" {
				tt.expected = "http://example.com/Path/To/Post"
			}
			result := NormalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Validation tests
// =============================================================================

func TestNormalizedItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    NormalizedItem
		wantErr bool
	}{
		{
			name: "Valid item",
			item: NormalizedItem{
				Title: "Valid Title",
				URL:   "https://example.com/valid",
			},
			wantErr: false,
		},
		{
			name: "Empty title",
			item: NormalizedItem{
				Title: "   ",
				URL:   "https://example.com/valid",
			},
			wantErr: true,
		},
		{
			name: "Empty URL",
			item: NormalizedItem{
				Title: "Valid Title",
				URL:   "   ",
			},
			wantErr: true,
		},
		{
			name: "Invalid URL format",
			item: NormalizedItem{
				Title: "Valid Title",
				URL:   "://invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Conversion tests
// =============================================================================

func TestNormalizedToContentItem(t *testing.T) {
	pubTime := time.Date(2024, 2, 6, 0, 0, 0, 0, time.UTC)

	t.Run("All fields populated", func(t *testing.T) {
		norm := NormalizedItem{
			SourceID:    42,
			Title:       "Test Title",
			URL:         "https://example.com/test",
			Author:      "Test Author",
			Summary:     "Test Summary",
			Body:        "Test Body",
			PublishedAt: pubTime,
			ContentType: "article",
		}

		ci := normalizedToContentItem(norm)

		if ci.SourceID != norm.SourceID {
			t.Errorf("SourceID = %d, want %d", ci.SourceID, norm.SourceID)
		}
		if ci.Title != norm.Title {
			t.Errorf("Title = %q, want %q", ci.Title, norm.Title)
		}
		if ci.URL != norm.URL {
			t.Errorf("URL = %q, want %q", ci.URL, norm.URL)
		}
		if ci.ContentType != norm.ContentType {
			t.Errorf("ContentType = %q, want %q", ci.ContentType, norm.ContentType)
		}
		if ci.Author == nil || *ci.Author != norm.Author {
			t.Errorf("Author = %v, want %q", ci.Author, norm.Author)
		}
		if ci.Summary == nil || *ci.Summary != norm.Summary {
			t.Errorf("Summary = %v, want %q", ci.Summary, norm.Summary)
		}
		if ci.Body == nil || *ci.Body != norm.Body {
			t.Errorf("Body = %v, want %q", ci.Body, norm.Body)
		}
		if ci.PublishedAt == nil || !ci.PublishedAt.Equal(norm.PublishedAt) {
			t.Errorf("PublishedAt = %v, want %v", ci.PublishedAt, norm.PublishedAt)
		}
	})

	t.Run("Optional fields empty", func(t *testing.T) {
		norm := NormalizedItem{
			SourceID:    1,
			Title:       "Minimal Title",
			URL:         "https://example.com/minimal",
			ContentType: "article",
		}

		ci := normalizedToContentItem(norm)

		if ci.Author != nil {
			t.Errorf("Author = %v, want nil", ci.Author)
		}
		if ci.Summary != nil {
			t.Errorf("Summary = %v, want nil", ci.Summary)
		}
		if ci.Body != nil {
			t.Errorf("Body = %v, want nil", ci.Body)
		}
		if ci.PublishedAt != nil {
			t.Errorf("PublishedAt = %v, want nil", ci.PublishedAt)
		}
	})
}
