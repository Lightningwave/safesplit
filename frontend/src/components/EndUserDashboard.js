// src/components/EndUserDashboard.js
import React from 'react';
import { Activity, CreditCard, Users, TrendingUp } from 'lucide-react';
import SideNavigation from './SideNavigation';

const EndUserDashboard = ({ user, onLogout }) => {
  const stats = [
    { id: 1, name: 'Total Transactions', value: '24', icon: <Activity className="w-6 h-6" /> },
    { id: 2, name: 'Active Splits', value: '3', icon: <Users className="w-6 h-6" /> },
    { id: 3, name: 'Monthly Spending', value: '$1,234', icon: <CreditCard className="w-6 h-6" /> },
    { id: 4, name: 'Split Success Rate', value: '95%', icon: <TrendingUp className="w-6 h-6" /> },
  ];

  const recentActivities = [
    { id: 1, type: 'Payment', description: 'Split bill at Restaurant XYZ', amount: '$45.00', date: '2024-12-05' },
    { id: 2, type: 'Split', description: 'Created new split with John and Sarah', amount: '$120.00', date: '2024-12-04' },
    { id: 3, type: 'Payment', description: 'Utilities split with roommates', amount: '$85.00', date: '2024-12-03' },
  ];

  return (
    <div className="flex h-screen bg-gray-100">
      <SideNavigation user={user} onLogout={onLogout} />
      
      {/* Main Content */}
      <div className="flex-1 ml-64 p-8">
        <div className="space-y-6">
          {/* Welcome Section */}
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold">Welcome back, {user.username}!</h1>
              <p className="text-gray-600">Here's what's happening with your splits</p>
            </div>
          </div>

          {/* Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {stats.map((stat) => (
              <div key={stat.id} className="bg-white p-6 rounded-lg shadow-sm">
                <div className="flex items-center justify-between">
                  <div className="text-gray-500">{stat.icon}</div>
                  <span className="text-2xl font-bold">{stat.value}</span>
                </div>
                <div className="mt-2 text-sm text-gray-600">{stat.name}</div>
              </div>
            ))}
          </div>

          {/* Recent Activities */}
          <div className="bg-white rounded-lg shadow-sm">
            <div className="p-6">
              <h2 className="text-lg font-bold mb-4">Recent Activities</h2>
              <div className="space-y-4">
                {recentActivities.map((activity) => (
                  <div key={activity.id} className="flex items-center justify-between border-b pb-4 last:border-0">
                    <div>
                      <p className="font-medium">{activity.description}</p>
                      <p className="text-sm text-gray-600">{activity.date}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-medium">{activity.amount}</p>
                      <p className="text-sm text-gray-600">{activity.type}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* User Profile Section */}
          <div className="bg-white rounded-lg shadow-sm">
            <div className="p-6">
              <h2 className="text-lg font-bold mb-4">Your Profile</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-gray-600">Username</p>
                  <p className="font-medium">{user.username}</p>
                </div>
                <div>
                  <p className="text-gray-600">Email</p>
                  <p className="font-medium">{user.email}</p>
                </div>
                <div>
                  <p className="text-gray-600">Account Type</p>
                  <p className="font-medium">{user.role.replace('_', ' ').toUpperCase()}</p>
                </div>
                <div>
                  <p className="text-gray-600">Member Since</p>
                  <p className="font-medium">
                    {new Date(user.created_at).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric'
                    })}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="bg-white rounded-lg shadow-sm">
            <div className="p-6">
              <h2 className="text-lg font-bold mb-4">Quick Actions</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <button className="p-4 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors">
                  Create New Split
                </button>
                <button className="p-4 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors">
                  Send Payment
                </button>
                <button className="p-4 bg-purple-500 text-white rounded-lg hover:bg-purple-600 transition-colors">
                  View All Transactions
                </button>
              </div>
            </div>
          </div>

          {/* Recent Split Groups */}
          <div className="bg-white rounded-lg shadow-sm">
            <div className="p-6">
              <h2 className="text-lg font-bold mb-4">Active Split Groups</h2>
              <div className="space-y-4">
                {[
                  { id: 1, name: 'Roommates', members: 4, total: '$350', due: '2024-12-10' },
                  { id: 2, name: 'Trip to Vegas', members: 6, total: '$1,200', due: '2024-12-15' },
                  { id: 3, name: 'Weekly Lunch', members: 3, total: '$45', due: '2024-12-08' },
                ].map((group) => (
                  <div key={group.id} className="flex items-center justify-between border-b pb-4 last:border-0">
                    <div>
                      <p className="font-medium">{group.name}</p>
                      <p className="text-sm text-gray-600">{group.members} members</p>
                    </div>
                    <div className="text-right">
                      <p className="font-medium">{group.total}</p>
                      <p className="text-sm text-gray-600">Due: {group.due}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  );
};

export default EndUserDashboard;