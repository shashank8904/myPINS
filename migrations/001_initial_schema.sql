-- =============================================================================
-- Migration: 001_initial_schema.sql
-- Project:   myPersonalTechRadar
-- Purpose:   V0 initial schema — all core domain entities.
--
-- Tables created (in dependency order):
--   1. sources
--   2. topics
--   3. content_items
--   4. content_item_topics
--   5. interests
--   6. roadmap_items
--   7. user_item_interactions
--
-- Design notes:
--   - All timestamps are TIMESTAMPTZ (timezone-aware).
--   - source_type / content_type use TEXT + CHECK rather than ENUM so new
--     values can be added without a DDL ALTER TYPE migration.
--   - relevance_score uses NUMERIC(4,3) for exact decimal storage (0.000–1.000).
--   - V1 note: wherever user_id belongs it is marked with a comment.
-- =============================================================================


-- =============================================================================
-- 1. sources
--    Represents where content originates (blog, GitHub, YouTube, etc.)
-- =============================================================================

CREATE TABLE sources (
    id          BIGSERIAL    PRIMARY KEY,
    name        TEXT         NOT NULL,
    url         TEXT         NOT NULL,
    source_type TEXT         NOT NULL
                             CHECK (source_type IN (
                                 'rss_feed',
                                 'github',
                                 'youtube',
                                 'hackernews',
                                 'conference',
                                 'other'
                             )),
    description TEXT,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- A source URL uniquely identifies the source (e.g. the RSS feed endpoint).
CREATE UNIQUE INDEX uq_sources_url ON sources (url);

-- Fast look-up by source type when filtering the feed by origin category.
CREATE INDEX idx_sources_source_type ON sources (source_type);


-- =============================================================================
-- 2. topics
--    A technology or subject area (Go, Kubernetes, Distributed Systems, etc.)
--    Central concept that links content, interests, and the learning roadmap.
-- =============================================================================

CREATE TABLE topics (
    id          BIGSERIAL    PRIMARY KEY,
    name        TEXT         NOT NULL,
    slug        TEXT         NOT NULL,   -- URL-safe: "distributed-systems"
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Both name and slug must be unique across topics.
CREATE UNIQUE INDEX uq_topics_name ON topics (name);
CREATE UNIQUE INDEX uq_topics_slug ON topics (slug);


-- =============================================================================
-- 3. content_items
--    A single piece of technology content (article, release, video, etc.)
--    Belongs to exactly one source.
-- =============================================================================

CREATE TABLE content_items (
    id            BIGSERIAL    PRIMARY KEY,
    source_id     BIGINT       NOT NULL REFERENCES sources (id) ON DELETE RESTRICT,
    title         TEXT         NOT NULL,
    url           TEXT         NOT NULL,
    content_type  TEXT         NOT NULL
                               CHECK (content_type IN (
                                   'article',
                                   'github_release',
                                   'youtube_video',
                                   'conference_talk',
                                   'research_paper',
                                   'hackernews_discussion',
                                   'other'
                               )),
    summary       TEXT,
    body          TEXT,
    author        TEXT,
    published_at  TIMESTAMPTZ,           -- nullable: may be unknown
    ingested_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Primary deduplication key: the canonical URL of the content.
-- If the same article appears through two sources, the URL collision
-- prevents a duplicate row.
CREATE UNIQUE INDEX uq_content_items_url ON content_items (url);

-- Feed queries will almost always filter by source.
CREATE INDEX idx_content_items_source_id ON content_items (source_id);

-- Feed ordering by ingestion or publication date.
CREATE INDEX idx_content_items_ingested_at ON content_items (ingested_at DESC);
CREATE INDEX idx_content_items_published_at ON content_items (published_at DESC);

-- Filter feed by content type (e.g. show only GitHub releases).
CREATE INDEX idx_content_items_content_type ON content_items (content_type);


-- =============================================================================
-- 4. content_item_topics
--    Junction table: many-to-many between content_items and topics.
--    relevance_score indicates how strongly an item relates to a topic
--    (0.000 = weak signal, 1.000 = primary topic).
-- =============================================================================

CREATE TABLE content_item_topics (
    id               BIGSERIAL    PRIMARY KEY,
    content_item_id  BIGINT       NOT NULL REFERENCES content_items (id) ON DELETE CASCADE,
    topic_id         BIGINT       NOT NULL REFERENCES topics (id) ON DELETE RESTRICT,
    -- NUMERIC(4,3): stores exactly 0.000–1.000 with no floating-point drift.
    relevance_score  NUMERIC(4,3) NOT NULL DEFAULT 1.000
                                  CHECK (relevance_score >= 0.000 AND relevance_score <= 1.000),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- An item can be linked to a topic only once.
CREATE UNIQUE INDEX uq_content_item_topics ON content_item_topics (content_item_id, topic_id);

-- Ranking query: find all items for a topic, ordered by relevance.
CREATE INDEX idx_content_item_topics_topic_id ON content_item_topics (topic_id, relevance_score DESC);


-- =============================================================================
-- 5. interests
--    The user's explicit interest level in a topic.
--    weight: 0.000 (no interest) → 1.000 (deeply interested).
--
--    V1 note: add `user_id BIGINT NOT NULL REFERENCES users(id)` here.
--    The UNIQUE constraint will become UNIQUE(user_id, topic_id).
-- =============================================================================

CREATE TABLE interests (
    id         BIGSERIAL    PRIMARY KEY,
    topic_id   BIGINT       NOT NULL REFERENCES topics (id) ON DELETE RESTRICT,
    -- NUMERIC(4,3): same reasoning as relevance_score.
    weight     NUMERIC(4,3) NOT NULL DEFAULT 1.000
                            CHECK (weight >= 0.000 AND weight <= 1.000),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- For a single-user system there can be at most one interest record per topic.
-- V1: change to UNIQUE(user_id, topic_id).
CREATE UNIQUE INDEX uq_interests_topic_id ON interests (topic_id);


-- =============================================================================
-- 6. roadmap_items
--    Something the user is currently learning or plans to learn.
--    status:   current | next | exploring | completed | paused
--    priority: explicit ordering within the same status group (lower = higher priority).
--
--    V1 note: add `user_id BIGINT NOT NULL REFERENCES users(id)` here.
--    The UNIQUE constraint will become UNIQUE(user_id, topic_id).
-- =============================================================================

CREATE TABLE roadmap_items (
    id         BIGSERIAL    PRIMARY KEY,
    topic_id   BIGINT       NOT NULL REFERENCES topics (id) ON DELETE RESTRICT,
    status     TEXT         NOT NULL DEFAULT 'exploring'
                            CHECK (status IN (
                                'current',
                                'next',
                                'exploring',
                                'completed',
                                'paused'
                            )),
    -- Optional explicit ordering within a status group.
    -- NULL means unordered; lower integer = higher priority.
    priority   INT,
    notes      TEXT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- For a single-user system a topic can only appear once on the roadmap.
-- V1: change to UNIQUE(user_id, topic_id).
CREATE UNIQUE INDEX uq_roadmap_items_topic_id ON roadmap_items (topic_id);

-- Roadmap list view: order by status then explicit priority.
CREATE INDEX idx_roadmap_items_status ON roadmap_items (status, priority ASC NULLS LAST);


-- =============================================================================
-- 7. user_item_interactions
--    Records a user action on a ContentItem.
--    action: read | saved | dismissed
--
--    The UNIQUE(content_item_id, action) constraint means one record per
--    (item, action) pair. A user can both READ and SAVE an item — those are
--    separate rows. But they cannot SAVE the same item twice.
--    Use an upsert (INSERT ... ON CONFLICT ... DO UPDATE) to refresh interacted_at.
--
--    V1 note: add `user_id BIGINT NOT NULL REFERENCES users(id)` here.
--    The UNIQUE constraint will become UNIQUE(user_id, content_item_id, action).
-- =============================================================================

CREATE TABLE user_item_interactions (
    id              BIGSERIAL    PRIMARY KEY,
    content_item_id BIGINT       NOT NULL REFERENCES content_items (id) ON DELETE CASCADE,
    action          TEXT         NOT NULL
                                 CHECK (action IN ('read', 'saved', 'dismissed')),
    interacted_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Enforces one record per (item, action) pair for the single user.
-- V1: change to UNIQUE(user_id, content_item_id, action).
CREATE UNIQUE INDEX uq_user_item_interactions ON user_item_interactions (content_item_id, action);

-- Fetch all interactions for a specific item (e.g. has the user saved this?).
CREATE INDEX idx_user_item_interactions_item ON user_item_interactions (content_item_id);

-- Feed personalization: find all items the user has dismissed or saved.
CREATE INDEX idx_user_item_interactions_action ON user_item_interactions (action);
