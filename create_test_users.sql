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
    TRUNCATE TABLE users;
    SET FOREIGN_KEY_CHECKS = 1;

    -- Create subscription plans
    INSERT INTO subscription_plans (name, storage_quota, price, billing_cycle) VALUES
    ('Basic', 5368709120, 0.00, 'monthly'),      -- 5GB free
    ('Premium', 53687091200, 8.99, 'monthly');   -- 50GB at $8.99

    -- Create test users
    -- Password is 'pw123' for all the test users
    INSERT INTO users (username, email, password, role, storage_quota, storage_used, subscription_status, two_factor_enabled) VALUES
    ('end_user', 'end_user@example.com', '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge', 'end_user', 5368709120, 4194304, 'free', false),
    ('premium_user', 'premium_user@example.com', '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge', 'premium_user', 53687091200, 10485760, 'premium', true),
    ('sys_admin', 'sys_admin@example.com', '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge', 'sys_admin', 5368709120, 0, 'free', true),
    ('super_admin', 'super_admin@example.com', '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge', 'super_admin', 5368709120, 0, 'free', true);


END //

DELIMITER ;

-- Execute the procedure
CALL create_test_data();