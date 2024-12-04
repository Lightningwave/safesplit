import React from 'react';
import { useNavigate } from 'react-router-dom';
import { logout } from '../services/authService';

const SuperAdminDashboard = ({ user }) => {
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="container mx-auto py-8">
      <h1 className="text-2xl font-bold mb-4">Super Admin Dashboard</h1>
      <div className="bg-white p-6 rounded shadow">
        <p>Welcome, {user.username}!</p>
        <p>Your role is: {user.role}</p>

        <div className="mt-4">
          <h2 className="text-xl font-bold">User Management</h2>
          {/* Add super admin-specific user management features here */}
          <p>Manage all user accounts and roles.</p>
        </div>

        <div className="mt-4">
          <h2 className="text-xl font-bold">System Configuration</h2>
          {/* Add super admin-specific system configuration features here */}
          <p>Configure all system settings and parameters.</p>
        </div>

        <div className="mt-4">
          <h2 className="text-xl font-bold">Reporting and Analytics</h2>
          {/* Add super admin-specific reporting and analytics features here */}
          <p>Access comprehensive reports and analytics.</p>
        </div>

        <button
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded mt-4"
          onClick={handleLogout}
        >
          Logout
        </button>
      </div>
    </div>
  );
};

export default SuperAdminDashboard;