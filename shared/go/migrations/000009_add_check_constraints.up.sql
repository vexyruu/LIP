ALTER TABLE listings
    ADD CONSTRAINT listings_condition_check CHECK (condition BETWEEN 1 AND 5);

ALTER TABLE users
    ADD CONSTRAINT users_status_check CHECK (status IN ('ACTIVE', 'BANNED'));
