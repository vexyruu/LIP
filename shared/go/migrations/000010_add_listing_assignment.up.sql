ALTER TABLE listings
    ADD COLUMN assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN assigned_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_listings_assigned_to ON listings (assigned_to);
