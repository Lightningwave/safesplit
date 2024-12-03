# Safesplit

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
│   │   ├── auth_controller.go
│   │   └── user_controller.go
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
- Node.js 14+
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

# Tidy up dependencies
go mod tidy
```

3. Set up the database
```bash
mysql -u root -p < database-setup.sql
```

4. Start the server
```bash
go run main.go
```

The server will start on `http://localhost:8080`

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
npx tailwindcss init -p       # Initialize Tailwind CSS

# Optional: if you want animations
npm install framer-motion     # For animations
```

3. Configure Tailwind CSS - Update `tailwind.config.js`:
```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
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

The application will be available at `http://localhost:3000`

## API Endpoints

### Public Routes
- POST `/api/login` - User login
- POST `/api/users` - User registration

### Protected Routes
- GET `/api/users/me` - Get current user profile

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](https://choosealicense.com/licenses/mit/)
