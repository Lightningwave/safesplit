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

    -- Create folders for users
    INSERT INTO folders (user_id, name, parent_folder_id) VALUES
    (1, 'Documents', NULL),
    (1, 'Photos', NULL),
    (1, 'Work', NULL),
    (2, 'Premium Documents', NULL),
    (2, 'Projects', NULL),
    (2, 'Archives', NULL);

    -- Create some nested folders
    INSERT INTO folders (user_id, name, parent_folder_id) VALUES
    (1, 'Personal', 1),
    (2, 'Client Files', 4);

    -- Create sample files with encryption data
    INSERT INTO files (
        user_id, folder_id, name, original_name, file_path, 
        size, mime_type, encryption_iv, encryption_salt, file_hash,
        is_archived, is_deleted, deleted_at
    ) VALUES
    -- Active files
    (1, 1, 'enc_resume.pdf', 'resume.pdf', '/storage/user1/enc_resume.pdf', 
        1048576, 'application/pdf', RANDOM_BYTES(16), RANDOM_BYTES(32), 
        SHA2('resume_content', 256), FALSE, FALSE, NULL),
    (1, 2, 'enc_vacation.jpg', 'vacation.jpg', '/storage/user1/enc_vacation.jpg', 
        2097152, 'image/jpeg', RANDOM_BYTES(16), RANDOM_BYTES(32), 
        SHA2('vacation_content', 256), FALSE, FALSE, NULL),
    -- Archived file
    (2, 4, 'enc_archived_doc.docx', 'archived_doc.docx', '/storage/user2/archived/enc_archived_doc.docx', 
        1572864, 'application/msword', RANDOM_BYTES(16), RANDOM_BYTES(32), 
        SHA2('archived_content', 256), TRUE, FALSE, NULL),
    -- Deleted file (for premium user)
    (2, 4, 'enc_deleted_doc.txt', 'deleted_doc.txt', '/storage/user2/deleted/enc_deleted_doc.txt', 
        512000, 'text/plain', RANDOM_BYTES(16), RANDOM_BYTES(32), 
        SHA2('deleted_content', 256), FALSE, TRUE, NOW());

    -- Create key fragments for each file
    INSERT INTO key_fragments (file_id, fragment_index, encrypted_fragment, holder_type) VALUES
    (1, 1, 'encrypted_fragment_1_owner', 'owner'),
    (1, 2, 'encrypted_fragment_1_system', 'system'),
    (2, 1, 'encrypted_fragment_2_owner', 'owner'),
    (2, 2, 'encrypted_fragment_2_system', 'system'),
    (3, 1, 'encrypted_fragment_3_owner', 'owner'),
    (3, 2, 'encrypted_fragment_3_system', 'system'),
    (4, 1, 'encrypted_fragment_4_owner', 'owner'),
    (4, 2, 'encrypted_fragment_4_system', 'system');

-- Create file shares with password protection
    INSERT INTO file_shares (
        file_id, shared_by, share_link, password_hash, password_salt, 
        encrypted_key_fragment, expires_at, max_downloads, download_count
    ) VALUES
    (1, 1, 'share_link_123', 
        '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge', -- hash of 'share123'
        'a1b2c3d4e5f6g7h8', 'encrypted_share_fragment_1', 
        DATE_ADD(NOW(), INTERVAL 7 DAY), 5, 0),
    (2, 1, 'share_link_456', 
        '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge', -- hash of 'share123'
        'h8g7f6e5d4c3b2a1', 'encrypted_share_fragment_2', 
        DATE_ADD(NOW(), INTERVAL 30 DAY), NULL, 0);

    -- Add share access logs
    INSERT INTO share_access_logs (share_id, ip_address, status, failure_reason) VALUES
    (1, '192.168.1.100', 'success', NULL),
    (1, '192.168.1.101', 'failed', 'Invalid password'),
    (2, '192.168.1.102', 'success', NULL);

    -- Add activity logs with encryption operations
    INSERT INTO activity_logs (user_id, activity_type, file_id, ip_address, status) VALUES
    (1, 'upload', 1, '192.168.1.1', 'success'),
    (1, 'encrypt', 1, '192.168.1.1', 'success'),
    (2, 'upload', 2, '192.168.1.2', 'success'),
    (2, 'encrypt', 2, '192.168.1.2', 'success'),
    (1, 'download', 2, '192.168.1.1', 'success'),
    (1, 'decrypt', 2, '192.168.1.1', 'success'),
    (2, 'archive', 3, '192.168.1.2', 'success'),
    (2, 'delete', 4, '192.168.1.2', 'success');

    -- Create user subscriptions
    INSERT INTO user_subscriptions (user_id, plan_id, start_date, end_date, status) VALUES
    (2, 2, NOW(), DATE_ADD(NOW(), INTERVAL 1 MONTH), 'active');

    -- Add feedback entries
    INSERT INTO feedback (user_id, type, subject, message, status) VALUES
    (1, 'feedback', 'Great File Sharing Feature', 'The password protection for shared files works perfectly!', 'pending'),
    (1, 'feedback', 'UI Suggestion', 'Would be great to have a dark mode option.', 'in_review'),
    (2, 'suspicious_activity', 'Unknown Access Attempt', 'Received email about login attempt from unknown location', 'pending'),
    (2, 'feedback', 'Premium Features', 'Love the additional storage space!', 'resolved');

END //

DELIMITER ;

-- Execute the procedure
CALL create_test_data();