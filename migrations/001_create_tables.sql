-- migrations/001_create_tables.sql
CREATE TABLE IF NOT EXISTS "customer" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "login_id" VARCHAR(16) UNIQUE NOT NULL,
    "email" TEXT UNIQUE NOT NULL,
    "password_hash" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "phone_number" TEXT,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "menu_tag_dimension" (
    "id" SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    "value" TEXT NOT NULL,
    "description" TEXT,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "menu_tag" (
    "id" SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    "value" TEXT NOT NULL,
    "description" TEXT,
    "dimension" SMALLINT,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY ("dimension") REFERENCES "menu_tag_dimension"("id") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS "menu_tag_prerequisite" (
    "menu_tag_id" SMALLINT,
    "prerequisite_tag_id" SMALLINT,
    PRIMARY KEY ("menu_tag_id", "prerequisite_tag_id"),
    FOREIGN KEY ("menu_tag_id") REFERENCES "menu_tag"("id") ON DELETE CASCADE,
    FOREIGN KEY ("prerequisite_tag_id") REFERENCES "menu_tag"("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "menu_item" (
    "id" SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "photo_pathinfo" TEXT,
    "price" INTEGER NOT NULL,
    "portion_size" SMALLINT NOT NULL DEFAULT 1,
    "available" BOOLEAN NOT NULL DEFAULT TRUE,
    "modifiers_config" JSONB,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "menu_item_tag" (
    "menu_item_id" SMALLINT,
    "menu_tag_id" SMALLINT,
    PRIMARY KEY ("menu_item_id", "menu_tag_id"),
    FOREIGN KEY ("menu_item_id") REFERENCES "menu_item"("id") ON DELETE CASCADE,
    FOREIGN KEY ("menu_tag_id") REFERENCES "menu_tag"("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "tab" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "total_price" INTEGER NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "closed_at" TIMESTAMP,
    "guest_names" JSONB
);

CREATE TABLE IF NOT EXISTS "order_id_sequence" (
    "tab_id" UUID PRIMARY KEY,
    "value" SMALLINT NOT NULL DEFAULT 0,
    FOREIGN KEY ("tab_id") REFERENCES "tab"("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "guest_id_sequence" (
    "tab_id" UUID PRIMARY KEY,
    "value" INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY ("tab_id") REFERENCES "tab"("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "order" (
    "tab_id" UUID,
    "scoped_id" SMALLINT,
    "sent_at" TIMESTAMP,
    PRIMARY KEY ("tab_id", "scoped_id"),
    FOREIGN KEY ("tab_id") REFERENCES "tab"("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "order_item_id_sequence" (
    "tab_id" UUID,
    "order_id" SMALLINT,
    "value" INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY ("tab_id", "order_id"),
    FOREIGN KEY ("tab_id", "order_id") REFERENCES "order"("tab_id", "scoped_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "order_item" (
    "tab_id" UUID,
    "order_id" SMALLINT,
    "scoped_id" SMALLINT,
    "menu_item_id" SMALLINT NOT NULL,
    "quantity" SMALLINT NOT NULL DEFAULT 1,
    "modifiers" JSONB,
    "guest_owners" SMALLINT ARRAY,
    "customer_owners" UUID ARRAY,
    PRIMARY KEY ("tab_id", "order_id", "scoped_id"),
    FOREIGN KEY ("tab_id", "order_id") REFERENCES "order"("tab_id", "scoped_id") ON DELETE CASCADE,
    FOREIGN KEY ("menu_item_id") REFERENCES "menu_item"("id") ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS "visitation" (
    "tab_id" UUID,
    "customer_id" UUID,
    PRIMARY KEY ("tab_id", "customer_id"),
    FOREIGN KEY ("tab_id") REFERENCES "tab"("id") ON DELETE CASCADE,
    FOREIGN KEY ("customer_id") REFERENCES "customer"("id") ON DELETE CASCADE
);

CREATE OR REPLACE VIEW "order_item_with_menu" AS
SELECT "oi".*, "mi"."name", "mi"."description", "mi"."photo_pathinfo", "mi"."price", "mi"."portion_size", "mi"."modifiers_config"
FROM "order_item" AS "oi"
JOIN "menu_item" AS "mi" ON "oi"."menu_item_id" = "mi"."id";

CREATE OR REPLACE VIEW "order_with_items" AS
SELECT "o".*, json_agg("oi") AS "items"
FROM "order" AS "o"
LEFT JOIN "order_item_with_menu" AS "oi" ON "o"."tab_id" = "oi"."tab_id" AND "o"."scoped_id" = "oi"."order_id"
GROUP BY "o"."tab_id", "o"."scoped_id";

CREATE OR REPLACE VIEW "tab_with_orders" AS
SELECT "t".*, json_agg("o") AS "orders"
FROM "tab" AS "t"
LEFT JOIN "order_with_items" AS "o" ON "t"."id" = "o"."tab_id"
GROUP BY "t"."id";
