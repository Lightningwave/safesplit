import React, { useState, useEffect } from 'react';
import { Shield, Loader2, AlertCircle, Search, Save } from 'lucide-react';

const UpdateUserPermissions = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [saving, setSaving] = useState(null);
  const [successMessage, setSuccessMessage] = useState('');

  const fetchUsers = async () => {
    try {
      const response = await fetch('/api/system/users', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
  
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to fetch users');
      }
      
      const data = await response.json();
      // Ensure we're accessing the correct data structure
      const usersList = Array.isArray(data.data) ? data.data : 
                       Array.isArray(data) ? data : [];
      setUsers(usersList);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };
  
  const updatePermissions = async (userId, updates) => {
    if (!userId) {
      console.error('No user ID provided');
      return;
    }
  
    setSaving(userId);
    setError(null);
    
    try {
      const currentUser = users.find(u => u.id === userId);
      if (!currentUser) {
        throw new Error('User not found');
      }
  
      const response = await fetch(`/api/system/users/${userId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          read_access: updates.readAccess,
          write_access: updates.writeAccess,
          username: currentUser.username,
          email: currentUser.email,
          account_type: currentUser.subscription_status
        })
      });
  
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to update permissions');
      }
      
      const data = await response.json();
      
      // Handle different response data structures
      const updatedUser = data.data?.user || data.user || data;
      
      setUsers(prevUsers => 
        prevUsers.map(user => 
          user.id === userId 
            ? { ...user, ...updatedUser }
            : user
        )
      );
  
      setSuccessMessage('Permissions updated successfully');
      setTimeout(() => setSuccessMessage(''), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(null);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const filteredUsers = users.filter(user =>
    user.username?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.email?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="animate-spin mr-2" />
        <span>Loading user permissions...</span>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header section */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-lg font-semibold">User Permissions</h2>
            <p className="text-sm text-gray-600 mt-1">Manage user access controls</p>
          </div>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={18} />
            <input
              type="text"
              placeholder="Search users..."
              className="pl-10 pr-4 py-2 border rounded-md"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      </div>

      {/* Error and Success Messages */}
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
            <Shield className="mr-2" size={20} />
            <span>{successMessage}</span>
          </div>
        </div>
      )}

      {/* Users Table */}
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Username</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Email</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Account Type</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Read Access</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Write Access</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredUsers.map((user) => (
              <tr key={user.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 text-sm text-gray-900">{user.username}</td>
                <td className="px-6 py-4 text-sm text-gray-900">{user.email}</td>
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
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={user.read_access}
                      onChange={(e) => updatePermissions(user.id, {
                        readAccess: e.target.checked,
                        writeAccess: user.write_access
                      })}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      disabled={saving === user.id}
                    />
                  </label>
                </td>
                <td className="px-6 py-4 text-sm">
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={user.write_access}
                      onChange={(e) => updatePermissions(user.id, {
                        readAccess: user.read_access,
                        writeAccess: e.target.checked
                      })}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      disabled={saving === user.id}
                    />
                  </label>
                </td>
                <td className="px-6 py-4 text-sm">
                  {saving === user.id && (
                    <Loader2 className="animate-spin" size={16} />
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredUsers.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            No users found matching your search
          </div>
        )}
      </div>
    </div>
  );
};

export default UpdateUserPermissions;