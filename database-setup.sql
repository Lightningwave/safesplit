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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create some test data (optional)
INSERT INTO users (username, email, password) VALUES
    ('testuser', 'test@example.com', 'password123'),
    ('demouser', 'demo@example.com', 'password123');

-- Show the created table structure
DESCRIBE users;

-- Show test data
SELECT * FROM users;
