INSERT INTO master_categories (id, name, type, icon, color) VALUES
    ('00000000-0000-0000-0000-000000000007', 'Others', 'expense', 'bag', 'c7')
ON CONFLICT DO NOTHING;
