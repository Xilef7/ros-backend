-- name: CreateTab :one
INSERT INTO "tab" DEFAULT VALUES
RETURNING "id", "created_at";

-- name: GetTabForShare :one
SELECT * FROM "tab" WHERE "id" = $1 FOR SHARE;

-- name: GetTabForNoKeyUpdate :one
SELECT * FROM "tab" WHERE "id" = $1 FOR NO KEY UPDATE;

-- name: GetOpenTabWithOrders :one
SELECT *
FROM "tab_with_orders"
WHERE "id" = $1 AND "closed_at" IS NULL;

-- name: GetTabWithOrdersForShare :one
SELECT "t".*, "o"."orders"
FROM "tab" AS "t",
(
    SELECT json_agg("o") AS "orders"
    FROM (
        SELECT "o".*, "oi"."items"
        FROM "order" AS "o"
        LEFT JOIN (
            SELECT "oi"."order_id", json_agg("oi") AS "items" FROM (
                SELECT "oi".*, "mi"."name", "mi"."description", "mi"."photo_pathinfo", "mi"."price", "mi"."portion_size", "mi"."modifiers_config"
                FROM "order_item" AS "oi"
                JOIN "menu_item" AS "mi" ON "oi"."menu_item_id" = "mi"."id"
                WHERE "oi"."tab_id" = $1
                FOR SHARE
            ) "oi"
            GROUP BY "oi"."order_id"
        ) "oi"
        ON "o"."scoped_id" = "oi"."order_id"
        WHERE "o"."tab_id" = $1
        FOR SHARE OF "o"
    ) "o"
) "o"
WHERE "t"."id" = $1
FOR SHARE OF "t";

-- name: GetVisitedTabsWithOrders :many
SELECT t.*
FROM "tab_with_orders" t
JOIN "visitation" v ON t."id" = v."tab_id"
WHERE v."customer_id" = $1;

-- name: UpdateTabTotalPrice :exec
UPDATE "tab" SET "total_price" = COALESCE((
    SELECT SUM(mi."price" * oi."quantity")
    FROM "order" o
    JOIN "order_item" oi ON o."tab_id" = oi."tab_id" AND o."scoped_id" = oi."order_id"
    JOIN "menu_item" mi ON oi."menu_item_id" = mi."id"
    WHERE o."tab_id" = $1 AND o."sent_at" IS NOT NULL
), 0)
WHERE "id" = $1;

-- name: CloseTab :one
UPDATE "tab" SET "closed_at" = NOW() WHERE "id" = $1
RETURNING "closed_at";

-- name: VisitTab :exec
INSERT INTO "visitation" ("tab_id", "customer_id")
VALUES ($1, $2)
ON CONFLICT ("tab_id", "customer_id") DO NOTHING;

-- name: IsVisitingCustomerIDs :many
SELECT "customer_id"
FROM "visitation"
WHERE "tab_id" = $1 AND "customer_id" = ANY(sqlc.arg('customer_ids')::UUID[]);

-- name: CreateGuestIDSequence :exec
INSERT INTO "guest_id_sequence" ("tab_id") VALUES ($1);

-- name: DeleteGuestIDSequence :exec
DELETE FROM "guest_id_sequence" WHERE "tab_id" = $1;

-- name: CreateGuest :one
UPDATE "guest_id_sequence" SET "value" = "value" + 1
WHERE "tab_id" = $1
RETURNING "value";

-- name: UpdateGuestName :exec
UPDATE "tab" SET "guest_names" = jsonb_set(
    COALESCE("guest_names", '{}'),
    ('{' || sqlc.arg('scoped_id')::SMALLINT || '}')::TEXT[],
    to_jsonb(sqlc.arg('name')::TEXT)
)
WHERE "id" = $1;

-- name: CreateOrderIDSequence :exec
INSERT INTO "order_id_sequence" ("tab_id") VALUES ($1);

-- name: DeleteOrderIDSequence :exec
DELETE FROM "order_id_sequence" WHERE "tab_id" = $1;

-- name: CreateOrder :one
WITH "seq" AS (
    UPDATE "order_id_sequence" SET "value" = "value" + 1
    WHERE "tab_id" = $1
    RETURNING "value"
)
INSERT INTO "order" ("tab_id", "scoped_id")
SELECT $1, "seq"."value"
FROM "seq"
RETURNING "scoped_id";

-- name: GetOrderForShare :one
SELECT * FROM "order" WHERE "tab_id" = $1 AND "scoped_id" = $2 FOR SHARE;

-- name: GetOrderForNoKeyUpdate :one
SELECT * FROM "order" WHERE "tab_id" = $1 AND "scoped_id" = $2 FOR NO KEY UPDATE;

-- name: GetOrderWithItems :one
SELECT * FROM "order_with_items" WHERE "tab_id" = $1 AND "scoped_id" = $2;

-- name: SendOrder :exec
UPDATE "order" SET "sent_at" = NOW() WHERE "tab_id" = $1 AND "scoped_id" = $2;

-- name: DeleteNotSentOrders :exec
DELETE FROM "order" WHERE "tab_id" = $1 AND "sent_at" IS NULL;

-- name: DeleteOrderItems :exec
DELETE FROM "order_item" WHERE "tab_id" = $1 AND "order_id" = $2;

-- name: CreateOrderItems :copyfrom
INSERT INTO "order_item" ("tab_id", "order_id", "scoped_id", "menu_item_id", "quantity", "modifiers", "guest_owners", "customer_owners")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateOrderItemQuantity :exec
UPDATE "order_item" SET "quantity" = $4
WHERE "tab_id" = $1 AND "order_id" = $2 AND "scoped_id" = $3;

-- name: UpdateOrderItemModifiers :exec
UPDATE "order_item" SET "modifiers" = $4
WHERE "tab_id" = $1 AND "order_id" = $2 AND "scoped_id" = $3;

-- name: AddOrderItemGuestOwner :exec
UPDATE "order_item" SET "guest_owners" = array_append("guest_owners", sqlc.arg('guest_id')::SMALLINT)
WHERE "tab_id" = $1 AND "order_id" = $2 AND "scoped_id" = $3 AND sqlc.arg('guest_id')::SMALLINT != ANY("guest_owners");

-- name: RemoveOrderItemGuestOwner :exec
UPDATE "order_item" SET "guest_owners" = array_remove("guest_owners", sqlc.arg('guest_id')::SMALLINT)
WHERE "tab_id" = $1 AND "order_id" = $2 AND "scoped_id" = $3 AND sqlc.arg('guest_id')::SMALLINT = ANY("guest_owners");;

-- name: AddOrderItemCustomerOwner :exec
UPDATE "order_item" SET "customer_owners" = array_append("customer_owners", sqlc.arg('customer_id')::UUID)
WHERE "tab_id" = $1 AND "order_id" = $2 AND "scoped_id" = $3 AND sqlc.arg('customer_id')::UUID != ANY("customer_owners");

-- name: RemoveOrderItemCustomerOwner :exec
UPDATE "order_item" SET "customer_owners" = array_remove("customer_owners", sqlc.arg('customer_id')::UUID)
WHERE "tab_id" = $1 AND "order_id" = $2 AND "scoped_id" = $3 AND sqlc.arg('customer_id')::UUID = ANY("customer_owners");

-- name: DeleteOrderItem :exec
DELETE FROM "order_item"
WHERE "tab_id" = $1 AND "order_id" = $2 AND "scoped_id" = $3;

-- name: CreateCustomer :one
INSERT INTO "customer" ("login_id", "email", "password_hash", "name", "phone_number")
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCustomerByID :one
SELECT * FROM "customer" WHERE "id" = $1;

-- name: GetCustomerByLogin :one
SELECT * FROM "customer" WHERE "login_id" = $1;

-- name: UpdateCustomerLoginID :one
UPDATE "customer" SET "login_id" = $2 WHERE "id" = $1
RETURNING *;

-- name: UpdateCustomerEmail :one
UPDATE "customer" SET "email" = $2 WHERE "id" = $1
RETURNING *; 

-- name: UpdateCustomerPassword :one
UPDATE "customer" SET "password_hash" = $2 WHERE "id" = $1
RETURNING *;

-- name: UpdateCustomerInfo :one
UPDATE "customer" SET "name" = $2, "phone_number" = $3 WHERE "id" = $1
RETURNING *;

-- name: CreateMenuItem :one
INSERT INTO "menu_item" ("name", "description", "photo_pathinfo", "price", "portion_size", "available", "modifiers_config")
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetMenuItem :one
SELECT * FROM "menu_item" WHERE "id" = $1;

-- name: GetNotDeletedMenuItem :one
SELECT * FROM "menu_item" WHERE "id" = $1 AND "deleted_at" IS NULL;

-- name: ListMenuItems :many
SELECT * FROM "menu_item" WHERE "deleted_at" IS NULL ORDER BY "name";

-- name: UpdateMenuItem :one
UPDATE "menu_item" SET "name" = $2, "description" = $3, "photo_pathinfo" = $4, "price" = $5, "portion_size" = $6, "available" = $7, "modifiers_config" = $8, "updated_at" = NOW()
WHERE "id" = $1
RETURNING *;

-- name: SoftDeleteMenuItem :exec
UPDATE "menu_item" SET "deleted_at" = COALESCE("deleted_at", NOW()) WHERE "id" = $1;

-- name: CreateMenuTag :one
INSERT INTO "menu_tag" ("value", "description", "dimension")
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListMenuTags :many
SELECT * FROM "menu_tag" ORDER BY "value";

-- name: CreateMenuTagDimension :one
INSERT INTO "menu_tag_dimension" ("value", "description")
VALUES ($1, $2)
RETURNING *;

-- name: ListMenuTagDimensions :many
SELECT * FROM "menu_tag_dimension" ORDER BY "value";
