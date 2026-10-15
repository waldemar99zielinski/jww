-- +goose Up
-- Bootstrap the category picker so a fresh install is usable before anyone
-- creates categories by hand. These are global reference rows, not user data.
--
-- IDs are fixed UUIDv5 values rather than gen_random_uuid() so a category has
-- the same id in every environment (reproducible seeds; a Down that targets
-- exactly what Up inserted). Derived as:
--   ns = uuid5(NAMESPACE_DNS, 'personal-finances.transaction_categories')
--      = 2153f979-c01d-5319-a66a-c03d9b9bcde9
--   id = uuid5(ns, '<type>:<name>')
--
-- ON CONFLICT keeps this safe on a database that already has some of these
-- names (unique index transaction_categories_type_name_uq).

INSERT INTO transaction_categories (id, type, name, active, sort_order) VALUES
    ('2787fb30-c4f9-5ec1-9c16-ccc256389c9c', 'expense', 'Groceries',     true,  0),
    ('4541c2c2-45ff-5dd9-90d9-e2421fcb7e3e', 'expense', 'Transport',     true,  1),
    ('3fed16e0-bf58-5794-8ae4-4041941ed8e7', 'expense', 'Entertainment', true,  2),
    ('b0d3a56f-5a86-54f3-a12b-bab556f5992c', 'expense', 'Health',        true,  3),
    ('fcd00418-c0e3-5377-ac1f-ec65e6fd2ffc', 'expense', 'Rent',          true,  4),
    ('282036a3-e0bc-5bab-bad1-5d8fec1b348b', 'expense', 'Clothes',       true,  5),
    ('638342b0-0b9f-56e9-9143-94b35f76e0b2', 'expense', 'Subscriptions', true,  6),
    ('eb09f5ec-a8bc-5bc5-8b98-9f35de3d40f6', 'expense', 'Travel',        true,  7),
    ('e100ee3a-7061-5c93-a9c5-80bdf7529ea0', 'expense', 'Utilities',     true,  8),
    ('6852cd7e-74ad-510e-baea-cedc2a7e23db', 'expense', 'Education',     true,  9),
    ('95fa10f1-3d93-5341-a8be-b890084e44f6', 'expense', 'Personal care', true, 10),
    ('6f17195a-7730-5a3d-b9d2-50e840168e26', 'expense', 'Gifts',         true, 11),
    ('32d17d97-785f-58ff-872e-abd1cc1e1191', 'expense', 'Dining out',    true, 12),
    ('663a127b-89af-5acb-9db4-4d6668780186', 'expense', 'Baby',          true, 13),
    ('162e6f9b-8658-5248-a153-061d6d886136', 'expense', 'Home',          true, 14),
    ('bb7abd90-3565-5e86-a1e4-cc72e6dd4daf', 'income',  'Salary',        true,  0),
    ('680c99ba-6285-51e0-b877-e1eede0cfd83', 'income',  'Freelance',     true,  1),
    ('1158118e-3ca3-5439-afcb-fb41b46023dc', 'income',  'Investments',   true,  2),
    ('fda81c3f-5fff-5a00-8c3d-834aa3443778', 'income',  'Refunds',       true,  3),
    ('e7c15778-21af-5830-9c9a-0e46975dcc06', 'income',  'Bonus',         true,  4),
    ('b3313992-4a4a-5609-807b-314a61fe49e9', 'income',  'Gifts',         true,  5)
ON CONFLICT (type, name) DO NOTHING;

-- +goose Down
-- Remove exactly the rows Up inserted. Fails loudly if transactions still
-- reference them (FK, no ON DELETE) — intended.
DELETE FROM transaction_categories
WHERE id IN (
    '2787fb30-c4f9-5ec1-9c16-ccc256389c9c',
    '4541c2c2-45ff-5dd9-90d9-e2421fcb7e3e',
    '3fed16e0-bf58-5794-8ae4-4041941ed8e7',
    'b0d3a56f-5a86-54f3-a12b-bab556f5992c',
    'fcd00418-c0e3-5377-ac1f-ec65e6fd2ffc',
    '282036a3-e0bc-5bab-bad1-5d8fec1b348b',
    '638342b0-0b9f-56e9-9143-94b35f76e0b2',
    'eb09f5ec-a8bc-5bc5-8b98-9f35de3d40f6',
    'e100ee3a-7061-5c93-a9c5-80bdf7529ea0',
    '6852cd7e-74ad-510e-baea-cedc2a7e23db',
    '95fa10f1-3d93-5341-a8be-b890084e44f6',
    '6f17195a-7730-5a3d-b9d2-50e840168e26',
    '32d17d97-785f-58ff-872e-abd1cc1e1191',
    '663a127b-89af-5acb-9db4-4d6668780186',
    '162e6f9b-8658-5248-a153-061d6d886136',
    'bb7abd90-3565-5e86-a1e4-cc72e6dd4daf',
    '680c99ba-6285-51e0-b877-e1eede0cfd83',
    '1158118e-3ca3-5439-afcb-fb41b46023dc',
    'fda81c3f-5fff-5a00-8c3d-834aa3443778',
    'e7c15778-21af-5830-9c9a-0e46975dcc06',
    'b3313992-4a4a-5609-807b-314a61fe49e9'
);
