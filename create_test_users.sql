DELIMITER //

DROP PROCEDURE IF EXISTS create_test_data //
CREATE PROCEDURE create_test_data()
BEGIN
    -- Clear existing data
    SET FOREIGN_KEY_CHECKS = 0;
    TRUNCATE TABLE feedback;
    TRUNCATE TABLE user_subscriptions;
    TRUNCATE TABLE subscription_plans;
    TRUNCATE TABLE share_access_logs;
    TRUNCATE TABLE activity_logs;
    TRUNCATE TABLE file_shares;
    TRUNCATE TABLE key_fragments;
    TRUNCATE TABLE files;
    TRUNCATE TABLE folders;
    TRUNCATE TABLE key_rotation_history;
    TRUNCATE TABLE password_history;
    TRUNCATE TABLE server_master_keys;
    TRUNCATE TABLE users;
    SET FOREIGN_KEY_CHECKS = 1;

    -- Create subscription plans
    INSERT INTO subscription_plans (name, storage_quota, price, billing_cycle) VALUES
    ('Basic', 5368709120, 0.00, 'monthly'),      -- 5GB free
    ('Premium', 53687091200, 8.99, 'monthly');   -- 50GB at $8.99

    -- Create test users with master key fields
    -- Password is 'pw123' for all test users
    INSERT INTO users (
        username, 
        email, 
        password, 
        master_key_salt,           -- 32-byte salt
        master_key_nonce,          -- 16-byte nonce
        encrypted_master_key,      -- 64-byte encrypted key
        master_key_version,
        role, 
        storage_quota, 
        storage_used, 
        subscription_status, 
        two_factor_enabled
    ) VALUES
    (
        'end_user',
        'end_user@example.com',
        '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge',
        UNHEX('CC31AD132A4D1F34A9FA9B6C3A693BB146D1BA449F2B573EBAB79DB9053EC715'), -- salt
        UNHEX('CC31AD132A4D1F34A9FA9B6C3A693BB1'),                                 -- nonce
        UNHEX('CC31AD132A4D1F34A9FA9B6C3A693BB146D1BA449F2B573EBAB79DB9053EC71500000000000000000000000000000000000000000000000000000000000000'), -- key
        1,
        'end_user',
        5368709120,
        4194304,
        'free',
        false
    ),
    (
        'premium_user',
        'premium_user@example.com',
        '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge',
        UNHEX('9A3254645BFD269D14C146680E49927B51EBC20ADA1EB6C94E4F704986DCC33A'), -- salt
        UNHEX('9A3254645BFD269D14C146680E49927B'),                                 -- nonce
        UNHEX('9A3254645BFD269D14C146680E49927B51EBC20ADA1EB6C94E4F704986DCC33A00000000000000000000000000000000000000000000000000000000000000'), -- key
        1,
        'premium_user',
        53687091200,
        10485760,
        'premium',
        false
    ),
    (
        'sys_admin',
        'sys_admin@example.com',
        '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge',
        UNHEX('6293E61742A9A26D16ABC91564FE26157923B855F58535AD73E9720C60F94C22'), -- salt
        UNHEX('6293E61742A9A26D16ABC91564FE2615'),                                 -- nonce
        UNHEX('6293E61742A9A26D16ABC91564FE26157923B855F58535AD73E9720C60F94C2200000000000000000000000000000000000000000000000000000000000000'), -- key
        1,
        'sys_admin',
        5368709120,
        0,
        'free',
        true
    ),
    (
        'super_admin',
        'super_admin@example.com',
        '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge',
        UNHEX('B5F83100C6B4F1FF3865A9FB3A32CBB1EF4770A734026D0AE0451737109750C7'), -- salt
        UNHEX('B5F83100C6B4F1FF3865A9FB3A32CBB1'),                                 -- nonce
        UNHEX('B5F83100C6B4F1FF3865A9FB3A32CBB1EF4770A734026D0AE0451737109750C700000000000000000000000000000000000000000000000000000000000000'), -- key
        1,
        'super_admin',
        5368709120,
        0,
        'free',
        true
    );

    -- Insert active server master key
    INSERT INTO server_master_keys (
        key_id,
        encrypted_key,
        key_nonce,
        is_active,
        activated_at
    ) VALUES (
        'server_key_1',
        UNHEX('CC31AD132A4D1F34A9FA9B6C3A693BB146D1BA449F2B573EBAB79DB9053EC71500000000000000000000000000000000000000000000000000000000000000'),
        UNHEX('CC31AD132A4D1F34A9FA9B6C3A693BB1'),
        true,
        CURRENT_TIMESTAMP
    );

END //

DELIMITER ;

-- Execute the procedure
CALL create_test_data();