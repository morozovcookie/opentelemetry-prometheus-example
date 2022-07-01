BEGIN;

CREATE TABLE user_accounts (
    row_id BIGINT AUTO_INCREMENT NOT NULL COMMENT 'row unique identifier',

    user_account_id VARCHAR(32) NOT NULL
        COMMENT 'user account unique identifier',
    username VARCHAR(255) NOT NULL COMMENT 'user username',
    user_id VARCHAR(32) NOT NULL COMMENT 'user unique identifier',
    created_at BIGINT NOT NULL COMMENT 'time when record was created',

    PRIMARY KEY (row_id ASC),
    INDEX user_account_id_hash_idx USING HASH (user_account_id),
    INDEX user_id_hash_idx USING HASH (user_id)
) COMMENT='stores information about user accounts' ENGINE=InnoDB;

CREATE TABLE users (
    row_id BIGINT AUTO_INCREMENT NOT NULL COMMENT 'row unique identifier',

    user_id VARCHAR(32) NOT NULL COMMENT 'user unique identifier',
    first_name VARCHAR(255) NOT NULL COMMENT 'user first name',
    last_name VARCHAR(255) NOT NULL COMMENT 'user last name',
    created_at BIGINT NOT NULL COMMENT 'time when record was created',

    PRIMARY KEY (row_id ASC),
    INDEX user_id_hash_idx USING HASH (user_id)
) COMMENT='stores information about users' ENGINE=InnoDB;

COMMIT;
