

DROP table if exists "assets" CASCADE;
CREATE table if not exists "assets" (
    "address" char(42) PRIMARY KEY,
    "symbol" varchar(10) NOT NULL,
    "first_block" int NOT NULL DEFAULT 0,
    "last_block" int NOT NULL DEFAULT 0,
    "class" varchar(10) NOT NULL -- is it a token or a native asset?
);


DROP table if exists "markets" CASCADE;
CREATE table if not exists "markets" (
    "address" char(42) PRIMARY KEY,
    "base_address" char(42) NOT NULL REFERENCES "assets" ("address"),
    "quote_address" char(42) NOT NULL REFERENCES "assets" ("address"),
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
    "price" numeric(78) NOT NULL, -- TODO: what precision do we need for price?
    "size" int NOT NULL,
    "market_address" char(42) NOT NULL REFERENCES "markets" ("address")
);

DROP table if exists "matches" CASCADE;
CREATE table if not exists "matches" (
    "id" char(36) NOT NULL,
    "order_id" char(36) NOT NULL REFERENCES "orders" ("id"),
    "price" numeric(78) NOT NULL,
    "size" int NOT NULL,
    "side" char(10) NOT NULL,
    "matched_at" timestamp NOT NULL,
    "status" varchar(10) NOT NULL,
    PRIMARY KEY ("id", "order_id")
);

CREATE INDEX "matches_index_order_id" ON "matches" USING btree ("order_id");
CREATE INDEX "matches_index_status" ON "matches" USING btree ("status");

DROP table if exists "balances" CASCADE;
CREATE table if not exists "balances" (
    "address" char(42) NOT NULL,
    "asset_address" char(42) NOT NULL REFERENCES "assets" ("address"),
    "balance" numeric(78) NOT NULL,
    PRIMARY KEY ("address", "asset_address")
);


DROP TABLE IF EXISTS "accounts" CASCADE;
CREATE TABLE IF NOT EXISTS "accounts" (
    "address" char(42) PRIMARY KEY,
    "active" boolean NOT NULL DEFAULT true
);



