CREATE TABLE IF NOT EXISTS music (
    id UUID NOT NULL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist_id UUID NOT NULL,
    cover_s3_key VARCHAR(1024),
    track_s3_key VARCHAR(1024),
    created_at TIMESTAMP DEFAULT NOW()
);
