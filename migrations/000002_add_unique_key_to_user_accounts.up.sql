BEGIN;

ALTER TABLE db.user_accounts ADD CONSTRAINT username_unique_idx UNIQUE (username);

COMMIT;
