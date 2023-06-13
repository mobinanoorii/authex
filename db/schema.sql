

DROP table if exists "tokens" CASCADE;
CREATE table if not exists "tokens" (
    "address" char(42) PRIMARY KEY,
    "symbol" varchar(10) NOT NULL,
    "first_block" int NOT NULL DEFAULT 0,
    "last_block" int NOT NULL DEFAULT 0,
    "asset_type" varchar(10) NOT NULL -- is it a token or a native asset?
);


DROP table if exists "markets" CASCADE;
CREATE table if not exists "markets" (
    "id" char(42) PRIMARY KEY,
    "base_address" char(42) NOT NULL REFERENCES "tokens" ("address"),
    "quote_address" char(42) NOT NULL REFERENCES "tokens" ("address"),
    "recorded_at" timestamp NOT NULL,
    "active" boolean NOT NULL DEFAULT true
);

DROP table if exists "orders" CASCADE;
CREATE table IF NOT EXISTS "orders" (
    "id" char(36) PRIMARY KEY,
    "from_address" char(42) NOT NULL, 
    "side" char(3) NOT NULL,
    "submitted_at" timestamp NOT NULL,
    "recorded_at" timestamp NOT NULL,
    "price" numeric(10, 2), -- TODO: what precision do we need for price?
    "quantity" int NOT NULL,
    "market_id" varchar(20) NOT NULL REFERENCES "markets" ("id")
);

-- DROP INDEX IF EXISTS "orders_index_address" ON "orders";
-- CREATE INDEX IF NOT EXISTS "orders_index_address" ON "orders" USING HASH (address);

-- DROP INDEX IF EXISTS "orders_index_side" ON orders;
-- CREATE INDEX IF NOT EXISTS "orders_index_side" ON "orders" USING HASH (side);


DROP table if exists "matches" CASCADE;
CREATE table if not exists "matches" (
    "top" char(36) NOT NULL,
    "bottom" char(36) NOT NULL,
    "price" numeric(10, 2) NOT NULL -- TODO: what precision do we need for price?
);

DROP table if exists "balances" CASCADE;
CREATE table if not exists "balances" (
    "address" char(42) NOT NULL,
    "token_address" char(42) NOT NULL REFERENCES "tokens" ("address"),
    "balance" numeric(78) NOT NULL,
    PRIMARY KEY ("address", "token_address")
);




