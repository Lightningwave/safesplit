import React from 'react';
import { useNavigate } from 'react-router-dom';
import { logout } from '../services/authService';

const PremiumUserDashboard = ({ user }) => {
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="container mx-auto py-8">
      <h1 className="text-2xl font-bold mb-4">Premium User Dashboard</h1>
      <div className="bg-white p-6 rounded shadow">
        <p>Welcome, {user.username}!</p>
        <p>Your role is: {user.role}</p>

        <div className="mt-4">
          <h2 className="text-xl font-bold">Your Premium Features</h2>
          {/* Add premium user-specific features here */}
          <p>You have access to premium features.</p>
        </div>

        <div className="mt-4">
          <h2 className="text-xl font-bold">Your Activities</h2>
          {/* Add premium user-specific activities here */}
          <p>No activities to display.</p>
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

export default PremiumUserDashboard;