
import React, { useState, useEffect } from 'react';
import { Database, Users, Loader2, AlertCircle, Search } from 'lucide-react';

const ViewStorage = () => {
  const [storageStats, setStorageStats] = useState(null);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchStorageStats = async () => {
    try {
      const response = await fetch('/api/system/storage/stats', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch storage statistics');
      const data = await response.json();
      setStorageStats(data.storage_stats);
      setUsers(data.storage_stats.users || []); // Ensure users array exists
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStorageStats();
  }, []);

  // Helper function to convert bytes to GB with two decimal places
  const bytesToGB = (bytes) => {
    return (bytes / 1024 / 1024 / 1024).toFixed(2);
  };

  const filteredUsers = users.filter(user =>
    user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.id.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="animate-spin mr-2" />
        <span>Loading storage statistics...</span>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header Section */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold">Storage Statistics</h2>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={18} />
            <input
              type="text"
              placeholder="Search by UserID or Username"
              className="pl-10 pr-4 py-2 border rounded-md"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 m-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center text-red-600">
            <AlertCircle className="mr-2" size={20} />
            <span>{error}</span>
          </div>
        </div>
      )}

      {/* Overall Storage Statistics */}
      {storageStats && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 p-6">
          {/* Total Users */}
          <div className="bg-white rounded-lg shadow p-6 flex items-center">
            <Users className="h-8 w-8 text-blue-500" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Total Users</p>
              <p className="text-2xl font-semibold">{storageStats.total_users}</p>
              <p className="text-sm text-gray-500">Active: {storageStats.active_users}</p>
            </div>
          </div>

          {/* Total Storage Used */}
          <div className="bg-white rounded-lg shadow p-6 flex items-center">
            <Database className="h-8 w-8 text-green-500" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Total Storage Used</p>
              <p className="text-2xl font-semibold">
                {bytesToGB(storageStats.storage_used)} GB
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Per-User Storage Details */}
      <div className="p-6">
        <h3 className="text-lg font-semibold mb-4">Per-User Storage Details</h3>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-sm font-medium text-gray-500 uppercase tracking-wider"
                >
                  UserID
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-sm font-medium text-gray-500 uppercase tracking-wider"
                >
                  Username
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-sm font-medium text-gray-500 uppercase tracking-wider"
                >
                  Account Type
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-sm font-medium text-gray-500 uppercase tracking-wider"
                >
                  Storage Used
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-sm font-medium text-gray-500 uppercase tracking-wider"
                >
                  Storage Total
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredUsers.map((user) => (
                <tr key={user.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{user.id}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{user.username}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                      user.subscription_status === 'premium' 
                        ? 'bg-purple-100 text-purple-800' 
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {user.subscription_status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {bytesToGB(user.storage_used)} GB
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {bytesToGB(user.storage_total)} GB
                  </td>
                </tr>
              ))}
              {filteredUsers.length === 0 && (
                <tr>
                  <td colSpan="5" className="px-6 py-4 text-center text-sm text-gray-500">
                    {searchQuery ? 'No users found matching your search.' : 'No user storage data available.'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default ViewStorage;
