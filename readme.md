# Safesplit

<p align="center">
  <img src="/frontend/public/safesplit-logo.png" alt="Safesplit Logo" width="300"/>
</p>

A web application built with Go (Gin) backend and React frontend for secure file sharing and recovery system.

## Tech Stack

### Backend

- Go
- Gin Web Framework
- GORM
- MySQL
- JWT Authentication

### Frontend

- React
- Tailwind CSS

## Project Structure Model-View(React frontend)-Controller

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

## Features

- User Registration
- User Login
- Protected Routes
- JWT Authentication
- Password Hashing

## Getting Started

### Prerequisites

- Go 1.16+
- MySQL

### Backend Setup

1. Clone the repository

```bash
git clone https://github.com/Lightningwave/safesplit.git
cd safesplit/backend
```

2. Install Go dependencies

```bash
# Initialize Go module
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

3. Set up the database

```bash
mysql -u root -p < database-setup.sql
# Initialize the database connection
cd ../backend/config/database.go
# change root:admin123 to your mysql username and password
```

4. Start the server

```bash
go run main.go
```

The server will start on ``

### Frontend Setup

1. Navigate to frontend directory

```bash
cd ../frontend
```

2. Install dependencies

```bash
# Core dependencies
npm install react-router-dom   # For routing
npm install lucide-react       # For icons

# Development dependencies
npm install -D tailwindcss postcss autoprefixer
npm install @tailwindcss/aspect-ratio
npx tailwindcss init -p

# Optional: if you want animations
npm install framer-motion
```

3. Configure Tailwind CSS - Update `tailwind.config.js`:

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

4. Update `src/index.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

5. Start the development server

```bash
npm start
```

The application will be available at `https://safesplit.xyz`

## API Endpoints

### Public Routes

- POST `/api/login` - User login
- POST `/api/users` - User registration

### Protected Routes

- GET `/api/users/me` - Get current user profile

## Testing

### Email testing with mailtrap

```bash
# Create new account with Mailtrap
Get your username and password details
# Create .env file on backend folder
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=465
SMTP_USERNAME= your username from Mailtrap
SMTP_PASSWORD= your password from Mailtrap
SMTP_FROM_NAME=Safesplit
SMTP_FROM_EMAIL=noreply@safesplit.com
```

### Payment testing with Paypal braintree

```bash
Use any credit card details from:
https://developer.paypal.com/braintree/docs/guides/credit-cards/testing-go-live/php
# Example
4111111111111111
12/25
123

```

### Amazon s3 servers for storage

```bash
# Update .env with your aws details

# AWS Credentials (shared across regions)
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=

# S3 Region 1
S3_REGION_1=us-east-1
S3_BUCKET_1=safesplit1

# S3 Region 2
S3_REGION_2=us-east-1
S3_BUCKET_2=safesplit2

# S3 Region 3
S3_REGION_3=us-east-1
S3_BUCKET_3=safesplit3
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
