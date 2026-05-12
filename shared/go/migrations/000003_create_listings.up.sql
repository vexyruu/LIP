CREATE TYPE listing_status AS ENUM ('DRAFT', 'LIVE', 'UNDER_REVIEW', 'REJECTED');

CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    price_ask DECIMAL(10, 2) NOT NULL,
    condition SMALLINT NOT NULL,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    status listing_status NOT NULL,
    suggested_price DECIMAL(10, 2),
    price_lower_bound DECIMAL(10, 2),
    price_upper_bound DECIMAL(10, 2),
    risk_score DECIMAL(10, 2),
    risk_tier TEXT,
    extracted_brand TEXT,
    extracted_product TEXT,
    extracted_size TEXT,
    policy_violation BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);