BEGIN;

ALTER TABLE db.user_accounts DROP CONSTRAINT username_unique_idx;

COMMIT;
