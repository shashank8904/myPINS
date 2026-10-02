-- =============================================================================
-- Migration: 003_add_source_quality_and_dedup.sql
-- Project:   myPersonalTechRadar
-- Purpose:   Implement Phase 3 features:
--            1. Source quality tracking
--            2. Layer 2 deduplication (normalized title + source)
-- =============================================================================

-- 1. Add source quality fields
ALTER TABLE sources
    ADD COLUMN base_quality_score NUMERIC(4,3) NOT NULL DEFAULT 1.000
        CHECK (base_quality_score >= 0.000 AND base_quality_score <= 1.000),
    ADD COLUMN quality_explanation TEXT;

-- Update existing sources to have an explanation
UPDATE sources SET quality_explanation = 'Initial default score.' WHERE quality_explanation IS NULL;

-- 2. Layer 2 Deduplication
-- Add a unique constraint to prevent the exact same title from the same source
-- from being inserted multiple times, even if the URL changes.
-- Using lower() and trim() to catch minor formatting differences.
CREATE UNIQUE INDEX uq_content_items_source_title 
    ON content_items (source_id, lower(trim(title)));
