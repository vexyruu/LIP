DROP INDEX IF EXISTS idx_listings_assigned_to;

ALTER TABLE listings
    DROP COLUMN IF EXISTS assigned_to,
    DROP COLUMN IF EXISTS assigned_at;
