import React from 'react';
import { useNavigate } from 'react-router-dom';
import { logout } from '../services/authService';

const SysAdminDashboard = ({ user }) => {
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="container mx-auto py-8">
      <h1 className="text-2xl font-bold mb-4">System Admin Dashboard</h1>
      <div className="bg-white p-6 rounded shadow">
        <p>Welcome, {user.username}!</p>
        <p>Your role is: {user.role}</p>

        <div className="mt-4">
          <h2 className="text-xl font-bold">User Management</h2>
          {/* Add system admin-specific user management features here */}
          <p>Manage users and their roles.</p>
        </div>

        <div className="mt-4">
          <h2 className="text-xl font-bold">System Settings</h2>
          {/* Add system admin-specific system settings here */}
          <p>Configure system settings.</p>
        </div>

        <div className="mt-4">
          <h2 className="text-xl font-bold">Monitoring</h2>
          {/* Add system admin-specific monitoring features here */}
          <p>Monitor system performance and activities.</p>
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

export default SysAdminDashboard;