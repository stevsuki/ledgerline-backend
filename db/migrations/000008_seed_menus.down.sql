-- Children first: menus_parent_id_fkey is ON DELETE RESTRICT.
DELETE FROM menus WHERE code IN (
    'dashboard', 'transactions', 'budgets', 'wallets', 'goals', 'recurring',
    'insights', 'shared', 'reminders', 'users', 'roles', 'audit',
    'settings', 'mobile'
);

DELETE FROM menus WHERE code IN ('overview', 'money', 'analysis', 'access', 'workspace');
