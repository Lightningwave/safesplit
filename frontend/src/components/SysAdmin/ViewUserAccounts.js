import React, { useState, useEffect } from 'react';
import { AlertCircle, MoreVertical, Loader2 } from 'lucide-react';
import PropTypes from 'prop-types';
import ConfirmationAction from './ConfirmationAction';
import ViewUserAction from './ViewUserAction';
import UpdateUserAction from './UpdateUserAction';

const ViewUserAccounts = ({ selectedType }) => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [activeMenu, setActiveMenu] = useState(null);
  
  // Modal states
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [userToDelete, setUserToDelete] = useState(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [userToView, setUserToView] = useState(null);
  const [isUpdateModalOpen, setIsUpdateModalOpen] = useState(false);
  const [userToUpdate, setUserToUpdate] = useState(null);
  const [currentUserToUpdate, setCurrentUserToUpdate] = useState(null);

  const fetchUsers = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/system/users', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch users');
      const data = await response.json();
      setUsers(data.users);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleSelectAll = (event) => {
    if (event.target.checked) {
      setSelectedUsers(filteredUsers.map(user => user.id));
    } else {
      setSelectedUsers([]);
    }
  };

  const handleSelectUser = (userId) => {
    setSelectedUsers(prev => {
      if (prev.includes(userId)) {
        return prev.filter(id => id !== userId);
      }
      return [...prev, userId];
    });
  };

  const filteredUsers = users.filter(user => {
    if (selectedType === 'premium') {
      return user.subscription_status === 'premium';
    } else if (selectedType === 'normal') {
      return user.subscription_status === 'free';
    }
    return true;
  });

  const handleDeleteClick = (user) => {
    setUserToDelete(user);
    setIsDeleteModalOpen(true);
  };

  const handleViewClick = (user) => {
    setUserToView(user.id);
    setIsViewModalOpen(true);
  };

  const handleUpdateClick = (user) => {
    setUserToUpdate(user.id);
    setCurrentUserToUpdate(user);
    setIsUpdateModalOpen(true);
  };

  const handleUpdateSuccess = (updatedUser) => {
    setUsers(prevUsers => 
      prevUsers.map(user => 
        user.id === updatedUser.id ? { ...user, ...updatedUser } : user
      )
    );
  };

  const confirmDelete = async () => {
    if (!userToDelete) return;

    try {
      const response = await fetch(`http://localhost:8080/api/system/users/${userToDelete.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to delete user');
      }

      fetchUsers();
      setIsDeleteModalOpen(false);
      setUserToDelete(null);
    } catch (err) {
      setError(err.message);
      setIsDeleteModalOpen(false);
      setUserToDelete(null);
    }
  };

  const ActionMenu = ({ user }) => (
    <div className="relative">
      <button 
        onClick={() => setActiveMenu(activeMenu === user.id ? null : user.id)}
        className="p-1 hover:bg-gray-100 rounded transition-colors duration-200"
        aria-haspopup="true"
        aria-expanded={activeMenu === user.id}
      >
        <MoreVertical size={16} />
      </button>
      
      {activeMenu === user.id && (
        <div className="absolute right-0 mt-2 w-40 bg-white rounded-md shadow-lg z-50 border">
          <div className="py-1">
            <button 
              onClick={() => {
                handleViewClick(user);
                setActiveMenu(null);
              }}
              className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 w-full text-left"
            >
              View
            </button>
            <button 
              onClick={() => {
                handleUpdateClick(user);
                setActiveMenu(null);
              }}
              className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 w-full text-left"
            >
              Update
            </button>
            <button 
              onClick={() => {
                handleDeleteClick(user);
                setActiveMenu(null);
              }}
              className="block px-4 py-2 text-sm text-red-600 hover:bg-gray-100 w-full text-left"
            >
              Delete
            </button>
          </div>
        </div>
      )}
    </div>
  );

  ActionMenu.propTypes = {
    user: PropTypes.object.isRequired,
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="animate-spin mr-2" />
        <span>Loading users...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8 text-red-500">
        <AlertCircle className="mr-2" />
        <span>{error}</span>
      </div>
    );
  }

  return (
    <>
      <div className="space-y-8">
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold">Users</h2>
          </div>

          <div className="bg-white rounded-lg shadow">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">
                    <input
                      type="checkbox"
                      onChange={handleSelectAll}
                      className="rounded border-gray-300"
                      aria-label="Select all users"
                    />
                  </th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">UserID</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Name</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Account Type</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Last Viewed</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredUsers.map(user => (
                  <tr key={user.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <input
                        type="checkbox"
                        checked={selectedUsers.includes(user.id)}
                        onChange={() => handleSelectUser(user.id)}
                        className="rounded border-gray-300"
                        aria-label={`Select user ${user.username}`}
                      />
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">{user.id}</td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {user.username}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {user.subscription_status === 'premium' ? 'Premium' : 'Normal'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {user.last_login ? new Date(user.last_login).toLocaleDateString() : 'Never'}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <ActionMenu user={user} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {filteredUsers.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                No users found
              </div>
            )}
          </div>
        </div>
      </div>

      <ConfirmationAction 
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setUserToDelete(null);
        }}
        onConfirm={confirmDelete}
        message={`Are you sure you want to delete user ${userToDelete?.username}?`}
      />

      <ViewUserAction 
        isOpen={isViewModalOpen}
        onClose={() => {
          setIsViewModalOpen(false);
          setUserToView(null);
        }}
        userId={userToView}
      />

      <UpdateUserAction 
        isOpen={isUpdateModalOpen}
        onClose={() => {
          setIsUpdateModalOpen(false);
          setUserToUpdate(null);
          setCurrentUserToUpdate(null);
        }}
        userId={userToUpdate}
        currentUser={currentUserToUpdate}
        onUpdateSuccess={handleUpdateSuccess}
      />
    </>
  );
};

ViewUserAccounts.propTypes = {
  selectedType: PropTypes.oneOf(['premium', 'normal', 'all']).isRequired,
};

export default ViewUserAccounts;