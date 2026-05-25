CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    seller_id UUID NOT NULL REFERENCES users(id),
    
    product_name TEXT NOT NULL,
    description TEXT NOT NULL,

    base_price FLOAT NOT NULL,
    auction_end TIMESTAMPTZ NOT NULL,

    is_sold BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
---- create above / drop below ----
DROP TABLE IF EXISTS products;