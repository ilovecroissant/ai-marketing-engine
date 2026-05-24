-- Enable pgvector extension (needed for RAG in Phase 7)
CREATE EXTENSION IF NOT EXISTS vector;

-- Users table
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,
    plan       VARCHAR(50) DEFAULT 'free',   -- 'free', 'pro', 'enterprise'
    created_at TIMESTAMP DEFAULT NOW()
);

-- Campaigns table
CREATE TABLE campaigns (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    product     VARCHAR(255) NOT NULL,
    website     VARCHAR(500),
    industry    VARCHAR(100) NOT NULL,
    goal        TEXT,
    tone        VARCHAR(50),
    platform    VARCHAR(50),   -- 'facebook', 'linkedin', 'blog', 'email'
    status      VARCHAR(50) DEFAULT 'pending',  -- 'pending', 'processing', 'done', 'failed'
    created_at  TIMESTAMP DEFAULT NOW()
);

-- Contents table
CREATE TABLE contents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id  UUID REFERENCES campaigns(id) ON DELETE CASCADE,
    title        VARCHAR(500),
    body         TEXT,
    score        FLOAT DEFAULT 0,      -- ranking score from C++ engine
    seo_score    FLOAT DEFAULT 0,
    engagement   FLOAT DEFAULT 0,
    readability  FLOAT DEFAULT 0,
    is_duplicate BOOLEAN DEFAULT FALSE,
    status       VARCHAR(50) DEFAULT 'draft',   -- 'draft', 'scheduled', 'published'
    scheduled_at TIMESTAMP,
    published_at TIMESTAMP,
    created_at   TIMESTAMP DEFAULT NOW()
);

-- Analytics table
CREATE TABLE analytics (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id   UUID REFERENCES contents(id) ON DELETE CASCADE,
    views        INT DEFAULT 0,
    clicks       INT DEFAULT 0,
    ctr          FLOAT DEFAULT 0,      -- click-through rate
    api_latency  FLOAT,                -- ms
    recorded_at  TIMESTAMP DEFAULT NOW()
);

-- Embeddings table (for RAG in Phase 7 — create now, use later)
CREATE TABLE embeddings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id  UUID REFERENCES contents(id) ON DELETE CASCADE,
    chunk_text  TEXT NOT NULL,
    embedding   vector(1536),          -- OpenAI ada-002 dimension
    created_at  TIMESTAMP DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX idx_campaigns_user_id  ON campaigns(user_id);
CREATE INDEX idx_contents_campaign  ON contents(campaign_id);
CREATE INDEX idx_analytics_content  ON analytics(content_id);
CREATE INDEX idx_contents_status    ON contents(status);