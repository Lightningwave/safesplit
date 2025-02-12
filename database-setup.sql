-- Drop database if exists and create new one
DROP DATABASE IF EXISTS safesplit;
CREATE DATABASE safesplit;
USE safesplit;

-- Users table
-- Purpose: Stores user account information, authentication details, and subscription status
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    master_key_salt BINARY(32) NOT NULL,           -- Salt for deriving master key
    master_key_nonce BINARY(16) NOT NULL,          -- Nonce for master key encryption
    encrypted_master_key BINARY(64) NOT NULL,      -- Encrypted user master key
    master_key_version INT NOT NULL DEFAULT 1,     -- Current version of master key
    key_last_rotated TIMESTAMP NULL,              -- Last key rotation timestamp
    role ENUM('end_user', 'premium_user', 'sys_admin', 'super_admin') NOT NULL DEFAULT 'end_user',
    read_access BOOLEAN NOT NULL DEFAULT TRUE,
    write_access BOOLEAN NOT NULL DEFAULT TRUE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    storage_quota BIGINT DEFAULT 5368709120,       -- Default 5GB in bytes
    storage_used BIGINT DEFAULT 0,
    subscription_status ENUM('free', 'premium', 'cancelled') DEFAULT 'free',
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP NULL,
    last_password_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    failed_login_attempts INT DEFAULT 0,
    account_locked_until TIMESTAMP NULL,
    force_password_change BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Server master keys table
-- Purpose: Manages server-side encryption keys
CREATE TABLE server_master_keys (
    id INT AUTO_INCREMENT PRIMARY KEY,
    key_id VARCHAR(64) NOT NULL UNIQUE,           -- Unique identifier for the key
    encrypted_key BINARY(32) NOT NULL,            -- Encrypted server master key
    key_nonce BINARY(16) NOT NULL,               -- Nonce for key encryption
    is_active BOOLEAN DEFAULT TRUE,              -- Whether this is the current active key
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    activated_at TIMESTAMP NULL,                 -- When the key became active
    retired_at TIMESTAMP NULL,                   -- When the key was retired
    UNIQUE INDEX idx_active_key (is_active, retired_at),
    INDEX idx_key_id (key_id)
);

-- Key rotation history table
-- Purpose: Tracks key rotation events
CREATE TABLE key_rotation_histories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    old_key_version INT NOT NULL,
    new_key_version INT NOT NULL,
    rotation_type ENUM('automatic', 'manual', 'forced') NOT NULL,
    rotated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_rotation (user_id, rotated_at)
);

-- Password history table
CREATE TABLE password_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Billing profiles table
CREATE TABLE billing_profiles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    customer_id VARCHAR(255) NOT NULL,           -- Payment processor's customer ID
    billing_email VARCHAR(255),                  -- Can be different from user email
    billing_name VARCHAR(255),                   -- Full name for billing
    billing_address TEXT,                        -- Full billing address
    country_code VARCHAR(2),                     -- ISO country code
    default_payment_method ENUM('credit_card', 'bank_account', 'paypal') DEFAULT 'credit_card',
    billing_cycle ENUM('monthly', 'yearly') DEFAULT 'monthly',
    currency VARCHAR(3) DEFAULT 'USD',
    next_billing_date TIMESTAMP NULL,
    last_billing_date TIMESTAMP NULL,
    billing_status ENUM('active', 'pending', 'failed', 'cancelled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_customer (customer_id)
);

-- Folders table
CREATE TABLE folders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- Owner of the folder
    name VARCHAR(255) NOT NULL,                   -- Folder name
    parent_folder_id INT,                         -- NULL for root folders
    is_archived BOOLEAN DEFAULT FALSE,            -- Whether folder is archived
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Files table
CREATE TABLE files (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- Owner of the file
    folder_id INT,                                -- NULL if in root
    name VARCHAR(255) NOT NULL,                   -- Encrypted file name
    original_name VARCHAR(255) NOT NULL,          -- Original file name
    file_path VARCHAR(1024) NOT NULL,             -- Path to encrypted file
    size BIGINT NOT NULL,                         -- File size in bytes
    compressed_size BIGINT,                       -- Size after compression
    is_compressed BOOLEAN DEFAULT FALSE,          -- Whether file is compressed
    compression_ratio DOUBLE PRECISION,           -- Compression ratio
    mime_type VARCHAR(127),                       -- File type
    is_archived BOOLEAN DEFAULT FALSE,            -- Whether file is archived
    is_deleted BOOLEAN DEFAULT FALSE,             -- Soft delete flag
    is_shared BOOLEAN DEFAULT FALSE,              -- Whether file is shared
    deleted_at TIMESTAMP NULL,                    -- Soft delete timestamp
    encryption_iv VARBINARY(24),                  -- Initialization vector
    encryption_salt BINARY(32),                   -- Salt for key derivation
    encryption_type VARCHAR(20) DEFAULT 'standard',-- Type of encryption used
    encryption_version INT DEFAULT 1,             -- Version of encryption
    master_key_version INT NOT NULL DEFAULT 1,    -- Version of master key used
    server_key_id VARCHAR(64) NULL,               -- ID of server key used
    share_count INTEGER NOT NULL DEFAULT 2,       -- Shamir's scheme shares
    threshold INTEGER NOT NULL DEFAULT 2,         -- Shamir's scheme threshold
    file_hash VARCHAR(64) NOT NULL,               -- Integrity verification
    data_shard_count INTEGER NOT NULL DEFAULT 4,  -- Reed-Solomon data shards
    parity_shard_count INTEGER NOT NULL DEFAULT 2,-- Reed-Solomon parity shards
    is_sharded BOOLEAN DEFAULT FALSE,             -- Uses Reed-Solomon
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);

-- Key fragments table
CREATE TABLE key_fragments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_id INT NOT NULL,
    fragment_index INT NOT NULL,                    -- Shamir share index (1-255)
    fragment_path VARCHAR(1024) NOT NULL,           -- Path in node storage
    node_index INT NOT NULL,                        -- Which node stores this fragment
    encryption_nonce BINARY(16) NOT NULL,           -- GCM nonce
    holder_type ENUM('user', 'server') NOT NULL,    -- Whether server or user holds this fragment
    master_key_version INT,                         -- Version of master key used (for user fragments)
    server_key_id VARCHAR(64),                      -- Server key ID (for server fragments)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    UNIQUE KEY unique_fragment (file_id, fragment_index)    
);

-- File shares table
CREATE TABLE file_shares (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_id INT NOT NULL,                         -- File being shared
    shared_by INT NOT NULL,                       -- User who created share
    share_link VARCHAR(255) UNIQUE NOT NULL,      -- Unique share link
    password_hash VARCHAR(255) NOT NULL,          -- Share password hash
    password_salt VARCHAR(32) NOT NULL,           -- Share password salt
    encrypted_key_fragment MEDIUMBLOB NOT NULL,   -- Password-encrypted fragment
    fragment_index INT NOT NULL DEFAULT 1,
    expires_at TIMESTAMP NULL,                    -- Optional expiration
    max_downloads INT NULL,                       -- Optional download limit
    download_count INT DEFAULT 0,                 -- Current downloads
    is_active BOOLEAN DEFAULT TRUE,               -- Share status
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    share_type VARCHAR(20) NOT NULL DEFAULT 'normal',  -- Share type (normal/recipient)
    email VARCHAR(255) NULL,                      -- Recipient email for recipient shares
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Share access logs table
CREATE TABLE share_access_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    share_id INT NOT NULL,                        -- Associated share
    ip_address VARCHAR(45) NOT NULL,              -- Access IP
    access_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('success', 'failed') NOT NULL,    -- Access result
    failure_reason VARCHAR(255),                  -- Failure reason
    FOREIGN KEY (share_id) REFERENCES file_shares(id) ON DELETE CASCADE
);

-- Activity logs table
CREATE TABLE activity_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- User performing action
    activity_type ENUM('upload', 'download', 'delete', 'share', 'login', 
                      'logout', 'archive', 'restore', 'encrypt', 'decrypt','unarchive') NOT NULL,
    file_id INT,                                  -- Associated file
    folder_id INT,                                -- Associated folder
    ip_address VARCHAR(45),                       -- User's IP
    status ENUM('success', 'failure') NOT NULL,   -- Operation result
    error_message TEXT,                           -- Error details
    details TEXT,                                 -- Additional details
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE SET NULL,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);

-- Subscription plans table
CREATE TABLE subscription_plans (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,                   -- Plan name
    storage_quota BIGINT NOT NULL,                -- Storage limit
    price DECIMAL(10,2) NOT NULL,                 -- Monthly price
    billing_cycle ENUM('monthly', 'yearly') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- User subscriptions table
CREATE TABLE user_subscriptions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- Subscribed user
    plan_id INT NOT NULL,                         -- Selected plan
    start_date TIMESTAMP NOT NULL,                -- Subscription start
    end_date TIMESTAMP NOT NULL,                  -- Subscription end
    status ENUM('active', 'cancelled', 'expired') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (plan_id) REFERENCES subscription_plans(id)
);

-- Feedback table
CREATE TABLE feedbacks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- User submitting feedback
    type ENUM('feedback', 'suspicious_activity') NOT NULL,
    subject VARCHAR(255) NOT NULL,                -- Brief description
    message TEXT NOT NULL,                        -- Detailed content
    details TEXT,
    status ENUM('pending', 'in_review', 'resolved') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_files_user_id ON files(user_id);
CREATE INDEX idx_files_folder_id ON files(folder_id);
CREATE INDEX idx_key_fragments_file_id ON key_fragments(file_id);
CREATE INDEX idx_file_shares_link ON file_shares(share_link);
CREATE INDEX idx_share_access_logs_share_id ON share_access_logs(share_id);
CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at);
CREATE INDEX idx_feedback_user_id ON feedbacks(user_id);
CREATE INDEX idx_feedback_status ON feedbacks(status);
CREATE INDEX idx_files_is_shared ON files(is_shared);
CREATE INDEX idx_files_key_version ON files(master_key_version);
CREATE INDEX idx_key_fragments_key_version ON key_fragments(master_key_version);
CREATE INDEX idx_server_master_keys_active ON server_master_keys(is_active);
