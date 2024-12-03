import React from 'react';
import { Navigate } from 'react-router-dom';

function Dashboard({ user }) {
  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-4">Welcome, {user.username}!</h1>
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4">Your Dashboard</h2>
        {/* Add dashboard content here */}
      </div>
    </div>
  );
}

export default Dashboard;