import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import NavigationBar from './components/NavigationBar';
import HomePage from './components/HomePage';
import AboutPage from './components/AboutPage';
import EndUserDashboard from './components/EndUserDashboard';
import PremiumUserDashboard from './components/PremiumUserDashboard';
import SuperAdminDashboard from './components/SuperAdminDashboard';
import SysAdminDashboard from './components/SysAdminDashboard';
import LoginForm from './components/auth/LoginForm';
import SuperAdminLogin from './components/auth/SuperAdminLogin';  
import RegisterForm from './components/auth/RegisterForm';
import ProtectedRoute from './components/auth/ProtectedRoute';
import SharedFileAccess from './components/EndUser/SharedFileAccess';
import { logout, getCurrentUser, getDashboardByRole } from './services/authService';

function App() {
  const [user, setUser] = useState(() => getCurrentUser());

  const handleLogout = () => {
    logout();
    setUser(null);
  };

  // Helper function to check if user has access to current route
  const isRouteAccessible = (allowedRoles) => {
    return user && allowedRoles.includes(user.role);
  };

  return (
    <Router>
      <div className="flex flex-col min-h-screen">
        {/* Show NavigationBar only when user is not logged in */}
        {!user && <NavigationBar user={user} onLogout={handleLogout} />}
        
        <Routes>
          {/* Public Routes */}
          <Route 
            path="/" 
            element={user ? <Navigate to={getDashboardByRole(user.role)} /> : <HomePage />} 
          />
          
          <Route path="/about" element={<AboutPage />} />
          
          {/* Authentication Routes */}
          <Route 
            path="/login" 
            element={
              user ? (
                <Navigate to={getDashboardByRole(user.role)} />
              ) : (
                <LoginForm onLogin={setUser} />
              )
            } 
          />
          
          <Route 
            path="/super-login" 
            element={
              user ? (
                <Navigate to={getDashboardByRole(user.role)} />
              ) : (
                <SuperAdminLogin onLogin={setUser} />
              )
            } 
          />
          
          <Route 
            path="/register" 
            element={
              user ? (
                <Navigate to={getDashboardByRole(user.role)} />
              ) : (
                <RegisterForm />
              )
            } 
          />

          <Route path="/share/:shareLink" element={<SharedFileAccess />} />

          {/* Protected Routes */}
          <Route 
            path="/dashboard" 
            element={
              <ProtectedRoute 
                element={(props) => <EndUserDashboard {...props} user={user} onLogout={handleLogout} />}
                allowedRoles={['end_user']} 
                isAuthenticated={isRouteAccessible(['end_user'])}
              />
            } 
          />
          
          <Route 
            path="/premium-dashboard" 
            element={
              <ProtectedRoute 
                element={(props) => <PremiumUserDashboard {...props} user={user} onLogout={handleLogout} />}
                allowedRoles={['premium_user']} 
                isAuthenticated={isRouteAccessible(['premium_user'])}
              />
            } 
          />
          
          <Route 
            path="/admin-dashboard" 
            element={
              <ProtectedRoute 
                element={(props) => <SysAdminDashboard {...props} user={user} onLogout={handleLogout} />}
                allowedRoles={['sys_admin']} 
                isAuthenticated={isRouteAccessible(['sys_admin'])}
              />
            } 
          />
          
          <Route 
            path="/super-dashboard" 
            element={
              <ProtectedRoute 
                element={(props) => <SuperAdminDashboard {...props} user={user} onLogout={handleLogout} />}
                allowedRoles={['super_admin']} 
                isAuthenticated={isRouteAccessible(['super_admin'])}
              />
            } 
          />

          {/* Catch all route */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;