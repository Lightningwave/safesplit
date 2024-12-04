-- safesplit_db.sql

-- Drop database if exists and create new one
DROP DATABASE IF EXISTS safesplit;
CREATE DATABASE safesplit;
USE safesplit;

-- Create users table
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role ENUM('end_user', 'premium_user', 'sys_admin', 'super_admin') NOT NULL DEFAULT 'end_user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);