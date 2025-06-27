-- Seed menu_tag_dimensions
INSERT INTO "menu_tag_dimension" (value, description, created_at, updated_at)
VALUES
    ('Spiciness', 'Level of spiciness', NOW(), NOW()),
    ('Dietary', 'Dietary restrictions', NOW(), NOW());

-- Seed menu_tags
INSERT INTO "menu_tag" (value, description, dimension, created_at, updated_at)
VALUES
    ('Vegan', 'No animal products', 2, NOW(), NOW()),
    ('Gluten-Free', 'No gluten', 2, NOW(), NOW()),
    ('Mild', 'Not spicy', 1, NOW(), NOW()),
    ('Hot', 'Very spicy', 1, NOW(), NOW());

-- Seed menu_items
INSERT INTO "menu_item" (name, description, photo_pathinfo, price, portion_size, available, modifiers_config, created_at, updated_at, deleted_at)
VALUES
    ('Margherita Pizza', 'Classic pizza with tomato, mozzarella, and basil', '/images/pizza1.jpg', 1200, 1, TRUE, NULL, NOW(), NOW(), NULL),
    ('Spicy Vegan Curry', 'Chickpea curry with extra spice', '/images/curry1.jpg', 1000, 1, TRUE, NULL, NOW(), NOW(), NULL),
    ('Gluten-Free Pasta', 'Pasta made with rice flour', '/images/pasta1.jpg', 1100, 1, TRUE, NULL, NOW(), NOW(), NULL);

-- Seed customers
INSERT INTO "customer" (id, login_id, email, password_hash, name, phone_number, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'alice', 'alice@example.com', 'hash1', 'Alice Smith', '123-456-7890', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000002', 'bob', 'bob@example.com', 'hash2', 'Bob Jones', '234-567-8901', NOW(), NOW());

-- Seed tabs
INSERT INTO "tab" (id, total_price, created_at, closed_at, guest_names)
VALUES
    ('11111111-1111-1111-1111-111111111111', 2300, NOW(), NULL, '{"1":"Alice","2":"Bob"}'::jsonb);

-- Seed visitations
INSERT INTO "visitation" (tab_id, customer_id)
VALUES
    ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000001'),
    ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000002');

-- Seed orders
INSERT INTO "order" (tab_id, scoped_id, sent_at)
VALUES
    ('11111111-1111-1111-1111-111111111111', 1, NOW());

-- Seed order_items
INSERT INTO "order_item" (tab_id, order_id, scoped_id, menu_item_id, quantity, modifiers, guest_owners, customer_owners)
VALUES
    ('11111111-1111-1111-1111-111111111111', 1, 1, 1, 1, NULL, NULL, ARRAY['00000000-0000-0000-0000-000000000001'::uuid]),
    ('11111111-1111-1111-1111-111111111111', 1, 2, 2, 2, NULL, NULL, ARRAY['00000000-0000-0000-0000-000000000002'::uuid]);

-- Note: Adjust UUIDs and IDs as needed for your environment.
