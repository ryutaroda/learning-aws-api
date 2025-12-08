CREATE TABLE bookmarks (
    id SERIAL PRIMARY KEY,
    url VARCHAR(2048) NOT NULL UNIQUE,
    title VARCHAR(500),
    description TEXT,
    image_url VARCHAR(2048),
    favicon_url VARCHAR(2048),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    tags TEXT[],
    fetched_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- インデックス
CREATE INDEX idx_bookmarks_status ON bookmarks(status);
CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at DESC);
CREATE INDEX idx_bookmarks_fetched_at ON bookmarks(fetched_at);

-- GINインデックス（配列検索用）
CREATE INDEX idx_bookmarks_tags ON bookmarks USING GIN(tags);

-- 全文検索用インデックス（PostgreSQL）
CREATE INDEX idx_bookmarks_title_search ON bookmarks USING GIN(to_tsvector('english', title));
CREATE INDEX idx_bookmarks_description_search ON bookmarks USING GIN(to_tsvector('english', description));