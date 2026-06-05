CREATE TYPE upload_status AS ENUM ('PENDING', 'READY', 'REJECTED');

CREATE TABLE uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    object_key TEXT NOT NULL,
    public_url TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    status upload_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_uploads_user_status ON uploads (user_id, status);
CREATE INDEX idx_uploads_public_url ON uploads (public_url);
