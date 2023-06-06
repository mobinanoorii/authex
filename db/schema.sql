

DROP table if exists "tokens" CASCADE;
CREATE table if not exists "tokens" (
    "address" char(42) NOT NULL,
    "symbol" varchar(10) NOT NULL UNIQUE,
    "first_block" int NOT NULL DEFAULT 0,
    "last_block" int NOT NULL DEFAULT 0,
    PRIMARY KEY ("address")
);


DROP table if exists "markets" CASCADE;
CREATE table if not exists "markets" (
    "id" varchar(20) NOT NULL,
    "base_symbol" varchar(10) NOT NULL,
    "quote_symbol" varchar(10) NOT NULL,
    "base_address" char(42),
    "quote_address" char(42),
    "active" boolean NOT NULL DEFAULT true,
    PRIMARY KEY ("id")
);

DROP table if exists "orders" CASCADE;
CREATE table IF NOT EXISTS "orders" (
    "id" char(36) NOT NULL PRIMARY KEY,
    "address" char(42) NOT NULL,
    "side" char(3) NOT NULL,
    "submitted_at" timestamp NOT NULL,
    "recorded_at" timestamp NOT NULL,
    "price" decimal(10,2),
    "quantity" decimal(10,2) NOT NULL,
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
    "price" decimal(10,2) NOT NULL
);

DROP table if exists "balances" CASCADE;
CREATE table if not exists "balances" (
    "address" char(42) NOT NULL,
    "symbol" varchar(10) NOT NULL REFERENCES "tokens" ("symbol"),
    "balance" decimal(10,2) NOT NULL,
    PRIMARY KEY ("address", "symbol")
);




