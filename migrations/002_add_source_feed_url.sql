-- =============================================================================
-- Migration: 002_add_source_feed_url.sql
-- Project:   myPersonalTechRadar
-- Purpose:   Add a nullable feed_url column to sources.
--
-- Rationale:
--   The sources.url column stores the human-facing homepage of a source
--   (e.g. https://go.dev/blog/). The ingestion layer needs a separate URL
--   that points to the machine-readable feed (e.g. https://go.dev/blog/feed.atom).
--   These are different things and must not be collapsed.
--
--   feed_url is nullable because:
--     - Not every source type has a feed (e.g. hackernews uses an API).
--     - Sources can be registered before their feed URL is known.
--
-- No data migration is needed: existing rows will have feed_url = NULL.
-- =============================================================================

ALTER TABLE sources
    ADD COLUMN feed_url TEXT;
