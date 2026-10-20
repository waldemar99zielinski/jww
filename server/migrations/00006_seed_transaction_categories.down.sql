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
