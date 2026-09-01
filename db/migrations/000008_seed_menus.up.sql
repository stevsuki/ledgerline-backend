-- Mirrors NAV_GROUPS in the web UI (lib/nav.ts): five groups, fourteen pages.
-- Fixed ids so children can reference their parent and every environment matches.
-- path stays NULL until the frontend routes are wired to this table.
INSERT INTO menus (id, parent_id, code, name, path, icon, sort_order) VALUES
    ('a0000000-0000-0000-0000-000000000001', NULL, 'overview',  'Overview',       NULL, NULL, 1),
    ('a0000000-0000-0000-0000-000000000002', NULL, 'money',     'Money',          NULL, NULL, 2),
    ('a0000000-0000-0000-0000-000000000003', NULL, 'analysis',  'Analysis',       NULL, NULL, 3),
    ('a0000000-0000-0000-0000-000000000004', NULL, 'access',    'Access control', NULL, NULL, 4),
    ('a0000000-0000-0000-0000-000000000005', NULL, 'workspace', 'Workspace',      NULL, NULL, 5)
ON CONFLICT (id) DO NOTHING;

INSERT INTO menus (id, parent_id, code, name, path, icon, sort_order) VALUES
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'dashboard',    'Dashboard',   NULL, 'grid',   1),

    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000002', 'transactions', 'Transactions', NULL, 'swap',   1),
    ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000002', 'budgets',      'Budgets',      NULL, 'target', 2),
    ('b0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000002', 'wallets',      'Wallets',      NULL, 'wallet', 3),
    ('b0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000002', 'goals',        'Goals',        NULL, 'flag',   4),
    ('b0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000002', 'recurring',    'Recurring',    NULL, 'repeat', 5),

    ('b0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000003', 'insights',     'Insights',     NULL, 'chart',  1),
    ('b0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000003', 'shared',       'Shared',       NULL, 'users',  2),
    ('b0000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000003', 'reminders',    'Reminders',    NULL, 'bell',   3),

    ('b0000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000004', 'users',        'Users',        NULL, 'users',  1),
    ('b0000000-0000-0000-0000-000000000011', 'a0000000-0000-0000-0000-000000000004', 'roles',        'Roles',        NULL, 'shield', 2),
    ('b0000000-0000-0000-0000-000000000012', 'a0000000-0000-0000-0000-000000000004', 'audit',        'Audit log',    NULL, 'search', 3),

    ('b0000000-0000-0000-0000-000000000013', 'a0000000-0000-0000-0000-000000000005', 'settings',     'Settings',     NULL, 'gear',   1),
    ('b0000000-0000-0000-0000-000000000014', 'a0000000-0000-0000-0000-000000000005', 'mobile',       'Mobile',       NULL, 'phone',  2)
ON CONFLICT (id) DO NOTHING;
