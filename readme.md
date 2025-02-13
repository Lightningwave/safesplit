# Safesplit

<p align="center">
  <img src="/frontend/public/safesplit-logo.png" alt="Safesplit Logo" width="300"/>
</p>

A secure file sharing and recovery system built with Go (Gin) backend and React frontend.

## Features

- 🔒 Secure file sharing with end-to-end encryption
- 📱 Mobile compatibility
- 🔑 JWT Authentication with password hashing
- 🔐 AES Encryption for files
- 🧩 Shamir secret sharing for encrypted key
- 📦 Reed-Solomon code for file splitting
- 🗜️ Zstd Compression
- ☁️ Distributed storage via Amazon S3 API
- 🔍 Two-factor authentication (2FA)
- 💳 Payment integration with PayPal Braintree (test mode)

## System Requirements

- Debian GNU/Linux 12 (bookworm)
- 2 vCPU
- 2GB RAM minimum
- 10GB storage minimum

## Tech Stack

### Backend

- Go 1.16+
- Gin Web Framework
- GORM
- MySQL
- JWT Authentication

### Frontend

- React
- Tailwind CSS

## Getting Started

### Prerequisites

- Go 1.16+
- MySQL
- Node.js v16.20+

### Prerequisites Installation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install essential tools
sudo apt install git nginx mysql-server golang-go nodejs npm -y
```

### MySQL Setup

```bash
# Secure MySQL installation
sudo mysql_secure_installation

# Create database and user
sudo mysql -u root -p
CREATE DATABASE safesplit;
USE safesplit;
exit;

# Load schema
sudo mysql safesplit < database-setup.sql
```

### Backend Setup

1. Clone repository and set up environment:

```bash
# Clone repository
git clone -b beta https://github.com/Lightningwave/safesplit.git
cd safesplit/backend

# Install Go dependencies
go mod init safesplit

# Install required packages
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u github.com/golang-jwt/jwt/v5
go get -u golang.org/x/crypto/bcrypt
go get github.com/klauspost/compress/zstd
go get github.com/klauspost/reedsolomon
go get github.com/hashicorp/vault/shamir
go get golang.org/x/crypto/chacha20poly1305
go get golang.org/x/crypto/twofish
go get github.com/braintree-go/braintree-go
go get github.com/joho/godotenv

# Tidy up dependencies
go mod tidy
```

2. Configure environment:

```bash
# Create .env file
cat > .env << 'EOL'
# Database settings
DB_HOST=localhost
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=safesplit
DB_PORT=3306

# SMTP settings
SMTP_HOST=mail.privateemail.com
SMTP_PORT=465
SMTP_USERNAME=noreply@safesplit.xyz
SMTP_PASSWORD=your_email_password
SMTP_FROM_NAME=Safesplit
SMTP_FROM_EMAIL=noreply@safesplit.xyz

# S3 Configuration
Singapore
S3_REGION_1=ap-southeast-1
S3_BUCKET_1=safesplit-sg-node
Japan
S3_REGION_2=ap-northeast-1
S3_BUCKET_2=safesplit-jp-node
Australia
S3_REGION_3=ap-southeast-2
S3_BUCKET_3=safesplit-aus-node

# AWS Credentials
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
EOL

# Set proper permissions
chmod 600 .env
```

3. Build and deploy:

```bash
# Build backend
go build -o main

# Create systemd service
sudo tee /etc/systemd/system/safesplit.service << 'EOL'
[Unit]
Description=Safesplit Backend Service
After=network.target mysql.service

[Service]
Type=simple
User=vaangyn
WorkingDirectory=/home/vaangyn/safesplit/backend
ExecStart=/home/vaangyn/safesplit/backend/main
EnvironmentFile=/home/vaangyn/safesplit/backend/.env
Restart=always
RestartSec=10
StartLimitInterval=0
StandardOutput=append:/var/log/safesplit.log
StandardError=append:/var/log/safesplit.error.log

[Install]
WantedBy=multi-user.target
EOL

# Create log files
sudo touch /var/log/safesplit.log /var/log/safesplit.error.log
sudo chown vaangyn:vaangyn /var/log/safesplit.log /var/log/safesplit.error.log

# Start service
sudo systemctl enable safesplit
sudo systemctl start safesplit
```

### Frontend Setup

1. Install dependencies:

```bash
cd ~/safesplit/frontend

# Core dependencies
npm install react-router-dom   # For routing
npm install lucide-react      # For icons

# Development dependencies
npm install -D tailwindcss postcss autoprefixer
npm install @tailwindcss/aspect-ratio
npx tailwindcss init -p

# Optional: for animations
npm install framer-motion
```

2. Configure Tailwind CSS - Update `tailwind.config.js`:

```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/aspect-ratio")],
};
```

3. Update `src/index.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

4. Build and deploy:

```bash
npm install
npm run build

# Set permissions
sudo chown -R www-data:www-data build/
sudo chmod -R 755 build/
```

### Nginx Configuration

```bash
# Create Nginx config
sudo tee /etc/nginx/sites-available/safesplit << 'EOL'
server {
    listen 80;
    server_name safesplit.xyz www.safesplit.xyz;

    client_max_body_size 500M;

    root /home/vaangyn/safesplit/frontend/build;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:8080/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
EOL

# Enable site
sudo ln -s /etc/nginx/sites-available/safesplit /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default

# Test and restart Nginx
sudo nginx -t
sudo systemctl restart nginx
```

### SSL/TLS Setup

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx -y

# Get certificate
sudo certbot --nginx -d safesplit.xyz -d www.safesplit.xyz
```

## Configuration

### Email Testing with Mailtrap

1. Create an account with Mailtrap
2. Update `.env` file in backend folder with Mailtrap credentials:

```bash
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=465
SMTP_USERNAME=your_username
SMTP_PASSWORD=your_password
SMTP_FROM_NAME=Safesplit
SMTP_FROM_EMAIL=noreply@safesplit.com
```

### Payment Testing with PayPal Braintree

Use test credit card numbers from [Braintree's testing guide](https://developer.paypal.com/braintree/docs/guides/credit-cards/testing-go-live/php)

Example test card:

```
Number: 4111111111111111
Expiry: 12/25
CVV: 123
```

## API Endpoints

### Public Routes

- POST `/api/login` - User login
- POST `/api/super-login` - Super admin login
- POST `/api/register` - User registration
- GET `/api/files/share/:shareLink` - Access shared file
- POST `/api/files/share/:shareLink` - Access shared file with password
- POST `/api/files/share/:shareLink/verify` - Verify 2FA and download shared file
- GET `/api/premium/shares/:shareLink` - Access premium shared file
- POST `/api/premium/shares/:shareLink` - Access premium shared file with password
- POST `/api/premium/shares/:shareLink/verify` - Verify 2FA and download premium shared file
- GET `/api/health` - Health check endpoint

### Protected Routes

- GET `/api/me` - Get current user profile

#### Two-Factor Authentication

- POST `/api/2fa/enable` - Enable email-based 2FA
- POST `/api/2fa/disable` - Disable email-based 2FA
- GET `/api/2fa/status` - Get 2FA status

#### File Management

- GET `/api/files` - List user files
- GET `/api/files/:id/download` - Download specific file
- POST `/api/files/mass-download` - Download multiple files
- GET `/api/files/mass-download/:id` - Get status of mass download
- POST `/api/files/upload` - Upload single file
- POST `/api/files/mass-upload` - Upload multiple files
- GET `/api/files/encryption/options` - Get available encryption options
- DELETE `/api/files/:id` - Delete specific file
- POST `/api/files/mass-delete` - Delete multiple files
- PUT `/api/files/:id/archive` - Archive specific file
- PUT `/api/files/:id/unarchive` - Unarchive specific file
- POST `/api/files/mass-archive` - Archive multiple files
- POST `/api/files/mass-unarchive` - Unarchive multiple files
- POST `/api/files/:id/share` - Create file share

#### Folder Management

- GET `/api/folders` - List root folders
- GET `/api/folders/:id` - Get folder contents
- POST `/api/folders` - Create new folder
- DELETE `/api/folders/:id` - Delete folder

#### Storage Management

- GET `/api/storage/info` - Get storage usage information

#### Payment & Subscription

- POST `/api/payment/upgrade` - Process payment for upgrade
- GET `/api/payment/status` - Get payment status
- POST `/api/payment/cancel` - Cancel subscription

#### Feedback System

- POST `/api/feedback` - Submit feedback
- GET `/api/feedback` - Get user's feedback
- GET `/api/feedback/categories` - Get feedback categories

#### Report System

- POST `/api/reports/file/:id` - Report a file
- POST `/api/reports/share/:shareLink` - Report a share
- GET `/api/reports` - Get user's reports

### Premium Routes

#### Fragment Management

- GET `/api/premium/fragments/files/:fileId` - Get user fragments for file

#### File Recovery

- GET `/api/premium/recovery/files` - List recoverable files
- POST `/api/premium/recovery/files/:fileId` - Recover specific file

#### Advanced Sharing

- POST `/api/premium/shares/files/:id` - Create advanced share

### Admin Routes

#### System Administration

- GET `/api/admin/me` - Get super admin profile
- POST `/api/admin/create-sysadmin` - Create system admin
- GET `/api/admin/sysadmins` - List system admins
- DELETE `/api/admin/sysadmins/:id` - Delete system admin
- GET `/api/admin/system-logs` - Get system logs

### System Admin Routes

#### User Management

- GET `/api/system/users` - List all users
- GET `/api/system/users/:id` - Get user details
- PUT `/api/system/users/:id` - Update user account
- DELETE `/api/system/users/:id` - Delete user account
- GET `/api/system/users/deleted` - Get deleted users
- POST `/api/system/users/deleted/:id/restore` - Restore deleted user

#### Storage Administration

- GET `/api/system/storage/stats` - Get storage statistics

#### Feedback Management

- GET `/api/system/feedback` - Get all feedbacks
- GET `/api/system/feedback/:id` - Get specific feedback
- PUT `/api/system/feedback/:id/status` - Update feedback status
- GET `/api/system/feedback/stats` - Get feedback statistics

#### Report Management

- GET `/api/system/reports` - Get all reports
- GET `/api/system/reports/:id` - Get report details
- PUT `/api/system/reports/:id/status` - Update report status
- GET `/api/system/reports/stats` - Get report statistics

## Maintenance Commands

### Service Management

```bash
# Backend service
sudo systemctl status safesplit
sudo systemctl start safesplit
sudo systemctl stop safesplit
sudo systemctl restart safesplit

# Nginx
sudo systemctl status nginx
sudo systemctl restart nginx

# MySQL
sudo systemctl status mysql
```

### Logs

```bash
# Backend logs
tail -f /var/log/safesplit.log
tail -f /var/log/safesplit.error.log

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# System logs
sudo journalctl -u safesplit -f
```

### Updates

```bash
# Pull latest code
cd ~/safesplit
git pull origin beta

# Update backend
cd backend
go build -o main
sudo systemctl restart safesplit

# Update frontend
cd ../frontend
sudo rm -rf build/
npm install
npm run build
sudo chown -R www-data:www-data build/
sudo chmod -R 755 build/
sudo systemctl restart nginx
```

## Project Structure example

```
safesplit/
├── backend/
│   ├── config/
│   │   └── database.go
│   ├── controllers/
│   │   ├── LoginController.go
│   │   └── LogoutController.go
│   │   └── CreateAccountController.go
│   ├── models/
│   │   └── user.go
│   ├── routes/
│   │   └── routes.go
│   └── main.go
└── frontend/
    ├── src/
    │   ├── components/
    │   │   └── auth/
    │   │       ├── LoginForm.js
    │   │       └── RegisterForm.js
    │   ├── services/
    │   │   └── authService.js
    │   └── App.js
    └── package.json
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
