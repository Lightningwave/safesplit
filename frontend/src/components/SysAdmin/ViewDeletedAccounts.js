import React, { useState, useEffect } from 'react';
import { Trash2, RefreshCcw, Loader2, AlertCircle, Search } from 'lucide-react';

const ViewDeletedAccounts = () => {
  const [deletedUsers, setDeletedUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [restoreLoading, setRestoreLoading] = useState(null);
  const [successMessage, setSuccessMessage] = useState(null);

  const fetchDeletedUsers = async () => {
    try {
      const response = await fetch('/api/system/users/deleted', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to fetch deleted accounts');
      const data = await response.json();
      setDeletedUsers(data.users);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleRestore = async (userId) => {
    setRestoreLoading(userId);
    setError(null);
    try {
      // Updated endpoint to match the backend route
      const response = await fetch(`/api/system/users/deleted/${userId}/restore`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to restore account');
      }

      setSuccessMessage('Account restored successfully');
      await fetchDeletedUsers();
      
      // Clear success message after 3 seconds
      setTimeout(() => {
        setSuccessMessage(null);
      }, 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setRestoreLoading(null);
    }
  };

  useEffect(() => {
    fetchDeletedUsers();
  }, []);

  const filteredUsers = deletedUsers.filter(user =>
    user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="animate-spin mr-2" />
        <span>Loading deleted accounts...</span>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      <div className="p-6 border-b border-gray-200">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold">Deleted Accounts</h2>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={18} />
            <input
              type="text"
              placeholder="Search deleted accounts..."
              className="pl-10 pr-4 py-2 border rounded-md"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      </div>

      {error && (
        <div className="p-4 m-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center text-red-600">
            <AlertCircle className="mr-2" size={20} />
            <span>{error}</span>
          </div>
        </div>
      )}

      {successMessage && (
        <div className="p-4 m-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center text-green-600">
            <RefreshCcw className="mr-2" size={20} />
            <span>{successMessage}</span>
          </div>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Username</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Email</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Deleted On</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Account Type</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredUsers.map((user) => (
              <tr key={user.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 text-sm text-gray-900">{user.username}</td>
                <td className="px-6 py-4 text-sm text-gray-900">{user.email}</td>
                <td className="px-6 py-4 text-sm text-gray-900">
                  {new Date(user.updated_at).toLocaleDateString()}
                </td>
                <td className="px-6 py-4 text-sm">
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                    user.subscription_status === 'premium' 
                      ? 'bg-purple-100 text-purple-800' 
                      : 'bg-gray-100 text-gray-800'
                  }`}>
                    {user.subscription_status}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm">
                  <button
                    onClick={() => handleRestore(user.id)}
                    disabled={restoreLoading === user.id}
                    className="flex items-center text-blue-600 hover:text-blue-800 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {restoreLoading === user.id ? (
                      <Loader2 className="animate-spin mr-1" size={16} />
                    ) : (
                      <RefreshCcw className="mr-1" size={16} />
                    )}
                    Restore Account
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredUsers.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            {searchQuery ? 'No deleted accounts found matching your search' : 'No deleted accounts found'}
          </div>
        )}
      </div>
    </div>
  );
};

export default ViewDeletedAccounts;