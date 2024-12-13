-- Drop database if exists and create new one
DROP DATABASE IF EXISTS safesplit;
CREATE DATABASE safesplit;
USE safesplit;

-- Users table
-- Purpose: Stores user account information, authentication details, and subscription status
-- Note: Handles both basic user data and premium features
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,         -- Unique username for login
    email VARCHAR(255) UNIQUE NOT NULL,            -- Unique email for notifications and recovery
    password VARCHAR(255) NOT NULL,                -- Hashed password
    role ENUM('end_user', 'premium_user', 'sys_admin', 'super_admin') NOT NULL DEFAULT 'end_user',  -- User access level
    read_access BOOLEAN NOT NULL DEFAULT TRUE,     -- User's read permission status
    write_access BOOLEAN NOT NULL DEFAULT TRUE,    -- User's write permission status
    two_factor_enabled BOOLEAN DEFAULT FALSE,      -- Whether 2FA is enabled
    two_factor_secret VARCHAR(255),                -- Secret key for 2FA if enabled
    storage_quota BIGINT DEFAULT 5368709120,       -- Storage limit (5GB for free users, 50GB for premium)
    storage_used BIGINT DEFAULT 0,                 -- Current storage usage in bytes
    subscription_status ENUM('free', 'premium', 'cancelled') DEFAULT 'free',  -- Current subscription level
    is_active BOOLEAN DEFAULT TRUE,                -- Whether account is active
    last_login TIMESTAMP NULL,                     -- Last successful login timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Billing-related fields
    payment_method VARCHAR(50) DEFAULT 'none',                     -- Payment method used by the user
    billing_cycle ENUM('monthly', 'yearly') DEFAULT 'monthly',      -- Billing cycle preference
    next_billing_date TIMESTAMP NULL,                               -- Next scheduled billing date
    last_billing_date TIMESTAMP NULL,                               -- Last successful billing date
    billing_status ENUM('active', 'pending', 'failed', 'cancelled') DEFAULT 'pending',  -- Current billing status
    customer_id VARCHAR(255)                                        -- Identifier for the customer in the payment system
);


-- Folders table
-- Purpose: Manages hierarchical organization of files
-- Note: Supports nested folders through parent_folder_id
CREATE TABLE folders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- Owner of the folder
    name VARCHAR(255) NOT NULL,                   -- Folder name
    parent_folder_id INT,                         -- NULL for root folders, points to parent for nested
    is_archived BOOLEAN DEFAULT FALSE,            -- Whether folder is archived
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Files table
-- Purpose: Stores file metadata and encryption details
-- Note: Handles both file information and encryption requirements
CREATE TABLE files (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- Owner of the file
    folder_id INT,                                -- NULL if in root, folder ID if in folder
    name VARCHAR(255) NOT NULL,                   -- Encrypted file name
    original_name VARCHAR(255) NOT NULL,          -- Original file name before encryption
    file_path VARCHAR(1024) NOT NULL,             -- Path to encrypted file on storage
    size BIGINT NOT NULL,                         -- File size in bytes
    mime_type VARCHAR(127),                       -- File type for download handling
    is_archived BOOLEAN DEFAULT FALSE,            -- Whether file is archived
    is_deleted BOOLEAN DEFAULT FALSE,             -- Soft delete flag
    deleted_at TIMESTAMP NULL,                    -- When file was soft deleted
    encryption_iv BINARY(16),            -- AES initialization vector
    encryption_salt BINARY(32),          -- Salt for key derivation
    file_hash VARCHAR(64) NOT NULL,               -- Hash for integrity verification
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);

-- Key fragments table
-- Purpose: Implements Shamir's Secret Sharing for encryption keys
-- Note: Splits encryption keys between owner and system
CREATE TABLE key_fragments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_id INT NOT NULL,                         -- Associated file
    fragment_index INT NOT NULL,                  -- Position in the key reconstruction sequence
    encrypted_fragment TEXT NOT NULL,             -- The encrypted key fragment
    holder_type ENUM('owner', 'system') NOT NULL, -- Who holds this fragment
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    UNIQUE KEY unique_fragment (file_id, fragment_index)
);

-- File shares table
-- Purpose: Manages password-protected file sharing
-- Note: Implements secure sharing with password protection and download limits
CREATE TABLE file_shares (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_id INT NOT NULL,                         -- File being shared
    shared_by INT NOT NULL,                       -- User who created the share
    share_link VARCHAR(255) UNIQUE NOT NULL,      -- Unique share link
    password_hash VARCHAR(255) NOT NULL,          -- Hash of share password
    password_salt VARCHAR(32) NOT NULL,           -- Salt for password hashing
    encrypted_key_fragment TEXT NOT NULL,         -- Fragment encrypted with share password
    expires_at TIMESTAMP NULL,                    -- Optional expiration date
    max_downloads INT NULL,                       -- Optional download limit
    download_count INT DEFAULT 0,                 -- Current download count
    is_active BOOLEAN DEFAULT TRUE,               -- Whether share is active
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Share access logs table
-- Purpose: Tracks access attempts to shared files
-- Note: Used for security monitoring and access control
CREATE TABLE share_access_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    share_id INT NOT NULL,                        -- Associated share
    ip_address VARCHAR(45) NOT NULL,              -- Access attempt IP
    access_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('success', 'failed') NOT NULL,    -- Access attempt result
    failure_reason VARCHAR(255),                  -- Reason for failure if failed
    FOREIGN KEY (share_id) REFERENCES file_shares(id) ON DELETE CASCADE
);

-- Activity logs table
-- Purpose: Tracks all user actions for auditing
-- Note: Comprehensive logging for security and troubleshooting
CREATE TABLE activity_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- User performing action
    activity_type ENUM('upload', 'download', 'delete', 'share', 'login', 
                      'logout', 'archive', 'restore', 'encrypt', 'decrypt') NOT NULL,
    file_id INT,                                  -- Associated file if any
    folder_id INT,                                -- Associated folder if any
    ip_address VARCHAR(45),                       -- User's IP address
    status ENUM('success', 'failure') NOT NULL,   -- Operation outcome
    error_message TEXT,                           -- Error details if failed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE SET NULL,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);

-- Subscription plans table
-- Purpose: Defines available subscription tiers
-- Note: Currently supports free (5GB) and premium (50GB) plans
CREATE TABLE subscription_plans (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,                   -- Plan name
    storage_quota BIGINT NOT NULL,                -- Storage limit in bytes
    price DECIMAL(10,2) NOT NULL,                 -- Monthly price
    billing_cycle ENUM('monthly', 'yearly') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- User subscriptions table
-- Purpose: Tracks user subscription history
-- Note: Maintains subscription status and billing periods
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

-- Feedback and reports table
-- Purpose: Stores user feedback and suspicious activity reports
-- Note: Handles both general feedback and security concerns
CREATE TABLE feedback (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,                         -- User submitting feedback
    type ENUM('feedback', 'suspicious_activity') NOT NULL,  -- Type of report
    subject VARCHAR(255) NOT NULL,                -- Brief description
    message TEXT NOT NULL,                        -- Detailed feedback content
    status ENUM('pending', 'in_review', 'resolved') DEFAULT 'pending',  -- Processing status
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
CREATE INDEX idx_feedback_user_id ON feedback(user_id);
CREATE INDEX idx_feedback_status ON feedback(status);