-- Wallets created before the create path stamped this kept the zero date, and a balance
-- stated in year one counts every transaction the wallet ever carried toward it.
-- Creation is the honest moment: the opening balance was given then.
UPDATE wallets
   SET balance_updated_at = created_at
 WHERE balance_updated_at < TIMESTAMPTZ '1970-01-01';
