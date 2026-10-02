package content

import (
	"context"
	"fmt"
)

// Service coordinates content operations.
// It sits between the HTTP/CLI layer and the Repository, owning the
// conversion from raw feed data into the ContentItem domain model.
type Service struct {
	repository *Repository
	fetcher    *RSSFetcher
}

func NewService(repository *Repository, fetcher *RSSFetcher) *Service {
	return &Service{
		repository: repository,
		fetcher:    fetcher,
	}
}

// Create persists a single ContentItem, returning the inserted row.
// It does not perform deduplication — use IngestRSS for feed ingestion.
func (s *Service) Create(ctx context.Context, item ContentItem) (ContentItem, error) {
	return s.repository.Create(ctx, item)
}

// IngestResult summarises the outcome of a single IngestRSS call.
type IngestResult struct {
	Discovered int // total items found in the feed
	Filtered   int // items rejected by validation
	Inserted   int // items newly written to the database
	Ignored    int // items skipped because the URL already exists
}

// IngestRSS fetches the RSS/Atom feed at feedURL and persists any new items
// under the given sourceID.
//
// Pipeline:
//  1. Convert source-specific struct to common intermediate form (NormalizedItem)
//  2. Normalize deduplication key (canonical URL)
//  3. Validate content quality
//  4. Convert to persistence model (ContentItem)
//  5. Persist with deduplication (ON CONFLICT DO NOTHING)
//
// What IngestRSS does NOT do (V0 scope):
//   - AI summarization
//   - Topic classification
//   - Relevance ranking
//
// These will be layered on in future milestones.
func (s *Service) IngestRSS(ctx context.Context, sourceID int64, feedURL string) (IngestResult, error) {
	items, err := s.fetcher.Fetch(feedURL)
	if err != nil {
		return IngestResult{}, fmt.Errorf("ingest RSS %s: %w", feedURL, err)
	}

	result := IngestResult{Discovered: len(items)}

	for _, rssItem := range items {
		// 1. Convert to intermediate representation
		norm := rssItemToNormalizedItem(sourceID, rssItem)

		// 2. Normalize URL for deduplication
		norm.URL = NormalizeURL(norm.URL)

		// 3. Validation
		if err := norm.Validate(); err != nil {
			// Log filtered items for debugging but continue processing
			// log.Printf("Filtered item %q: %v", norm.URL, err)
			result.Filtered++
			continue
		}

		// 4. Convert to database format
		ci := normalizedToContentItem(norm)

		// 5. Persist
		_, inserted, err := s.repository.CreateOrIgnore(ctx, ci)
		if err != nil {
			return IngestResult{}, fmt.Errorf("persist item %q: %w", norm.URL, err)
		}

		if inserted {
			result.Inserted++
		} else {
			result.Ignored++
		}
	}

	return result, nil
}

// rssItemToNormalizedItem converts a parsed RSSItem into a NormalizedItem.
// This is the boundary where source-specific formats enter the common pipeline.
func rssItemToNormalizedItem(sourceID int64, item RSSItem) NormalizedItem {
	return NormalizedItem{
		SourceID:    sourceID,
		Title:       item.Title,
		URL:         item.URL,
		Author:      item.Author,
		Summary:     item.Description,
		PublishedAt: item.PublishedAt,
		ContentType: "article", // RSS/Atom feeds are articles in V0
	}
}

// GetByID returns a single ContentItem by ID.
func (s *Service) GetByID(ctx context.Context, id int64) (ContentItem, error) {
	return s.repository.GetByID(ctx, id)
}

// List returns the personalized feed (ranked, explained).
func (s *Service) List(ctx context.Context) ([]ContentItem, error) {
	// For V0, hardcode a limit of 50 items.
	return s.repository.ListPersonalized(ctx, 50)
}

// Interact records a user action (e.g. read, saved, dismissed) for an item.
func (s *Service) Interact(ctx context.Context, itemID int64, action string) error {
	return s.repository.RecordInteraction(ctx, itemID, action)
}
