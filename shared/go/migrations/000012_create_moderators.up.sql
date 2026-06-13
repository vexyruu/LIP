CREATE TABLE moderators (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    role        TEXT NOT NULL CHECK (role IN ('MODERATOR', 'ANALYST', 'ADMIN')),
    password_hash TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO moderators (id, email, display_name, role, password_hash) VALUES
    ('22222222-2222-2222-2222-222222222222', 'mod@mlip.dev',     'Moderator', 'MODERATOR', '$2b$10$HKTuUf10h9ecIInNnVYxm.rlXFOkD/Kw3Da061ts/15BwMepvey4i'),
    ('44444444-4444-4444-4444-444444444444', 'analyst@mlip.dev', 'Analyst',   'ANALYST',   '$2b$10$XomwP8AWsj1DAPc86iJD8uC1I8YP4ot6WmwzLGYjaIEQcsHyP3hCC'),
    ('55555555-5555-5555-5555-555555555555', 'admin@mlip.dev',   'Admin',     'ADMIN',     '$2b$10$QFQ4TzWfzpjrbFbut0JWC.8g6sDn/QNStMUNidDg2.IAVgFOUkx8G');
