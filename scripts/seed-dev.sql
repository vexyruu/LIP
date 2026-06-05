-- Local dev seed for MLIP (Postgres). Run after migrations:
-- Get-Content scripts\seed-dev.sql | docker compose exec -T postgres psql -U postgres -d mlip

-- Clean seller (Listing Lab default) + fraud demo seller + moderator
INSERT INTO users (id, email, display_name, status)
VALUES
    ('00000000-0000-0000-0000-000000000002', 'clean@mlip.dev', 'Clean Seller', 'ACTIVE'),
    ('11111111-1111-1111-1111-111111111111', 'seller@mlip.dev', 'Fraud Ring Seller', 'ACTIVE'),
    ('22222222-2222-2222-2222-222222222222', 'mod@mlip.dev', 'Moderator', 'ACTIVE')
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    status = EXCLUDED.status;

-- Category id=1 for Postman / Listing Lab body "category_id": 1
INSERT INTO categories (id, name, path)
OVERRIDING SYSTEM VALUE
VALUES (1, 'Athletic', 'Men/Shoes/Athletic')
ON CONFLICT (id) DO NOTHING;

-- Demo listing on fraud seller for moderation queue (re-run safe)
INSERT INTO listings (
    id,
    user_id,
    title,
    description,
    price_ask,
    condition,
    category_id,
    status,
    risk_score,
    risk_tier,
    suggested_price,
    price_lower_bound,
    price_upper_bound,
    extracted_brand,
    extracted_product,
    extracted_size,
    policy_violation,
    images
)
VALUES (
    '33333333-3333-3333-3333-333333333333',
    '11111111-1111-1111-1111-111111111111',
    'Air Jordan 1 Retro High OG',
    'Size 10. Worn twice. Comes with original box. Chicago colorway.',
    285.00,
    2,
    1,
    'UNDER_REVIEW',
    0.92,
    'HIGH',
    260.00,
    220.00,
    310.00,
    'Nike',
    'Air Jordan 1',
    '10',
    false,
    ARRAY[
        'https://images.unsplash.com/photo-1556906781-9a412961c28c?auto=format&fit=crop&w=900&q=80',
        'https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=900&q=80'
    ]
)
ON CONFLICT (id) DO UPDATE SET
    images = EXCLUDED.images,
    status = EXCLUDED.status,
    risk_score = EXCLUDED.risk_score,
    risk_tier = EXCLUDED.risk_tier;
