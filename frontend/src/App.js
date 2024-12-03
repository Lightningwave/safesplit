import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import NavigationBar from './components/NavigationBar';
import HomePage from './components/HomePage';
import Dashboard from './components/Dashboard';
import LoginForm from './components/auth/LoginForm';
import RegisterForm from './components/auth/RegisterForm';
import AboutPage from './components/AboutPage';  // Add this import
import { logout } from './services/authService';

function App() {
  const [user, setUser] = useState(() => {
    const savedUser = localStorage.getItem('user');
    return savedUser ? JSON.parse(savedUser) : null;
  });

  const handleLogout = () => {
    logout();
    setUser(null);
  };

  return (
    <Router>
      <div className="flex flex-col min-h-screen">
        <NavigationBar user={user} onLogout={handleLogout} />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/login" element={<LoginForm onLogin={setUser} />} />
          <Route path="/register" element={<RegisterForm />} />
          <Route path="/dashboard" element={<Dashboard user={user} />} />
          <Route path="/about" element={<AboutPage />} />  {/* Add this route */}
        </Routes>
      </div>
    </Router>
  );
}

export default App;