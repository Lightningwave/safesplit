import React, { useState, useEffect } from 'react';
import { AlertCircle, MoreVertical, Loader2 } from 'lucide-react';
import PropTypes from 'prop-types';
import ViewUserAction from './ViewUserAction';
import UpdateUserAction from './UpdateUserAction';
import DeleteUserAction from './DeleteUserAction';

const ViewUserAccounts = ({ selectedType, onUserViewed }) => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [activeMenu, setActiveMenu] = useState(null);
  
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [userToDelete, setUserToDelete] = useState(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [userToView, setUserToView] = useState(null);
  const [isUpdateModalOpen, setIsUpdateModalOpen] = useState(false);
  const [userToUpdate, setUserToUpdate] = useState(null);
  const [currentUserToUpdate, setCurrentUserToUpdate] = useState(null);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:8080/api/system/users', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to fetch users');
      }
      
      const data = await response.json();
      const usersList = Array.isArray(data.users) ? data.users : 
                       Array.isArray(data.data) ? data.data :
                       Array.isArray(data) ? data : [];
      
      const validUsers = usersList.filter(user => 
        user && 
        typeof user === 'object' && 
        user.id !== undefined &&
        user.username !== undefined &&
        user.subscription_status !== undefined
      );
      
      setUsers(validUsers);
    } catch (err) {
      console.error('Error fetching users:', err);
      setError(err.message);
      setUsers([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => setError(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [error]);

  const filteredUsers = users.filter(user => {
    if (!user || !user.id || !user.subscription_status) return false;
    
    if (selectedType === 'premium') {
      return user.subscription_status === 'premium';
    } else if (selectedType === 'normal') {
      return user.subscription_status === 'free';
    }
    return true;
  });

  const handleSelectAll = (event) => {
    if (event.target.checked && filteredUsers.length > 0) {
      const validUserIds = filteredUsers
        .filter(user => user && user.id)
        .map(user => user.id);
      setSelectedUsers(validUserIds);
    } else {
      setSelectedUsers([]);
    }
  };

  const handleSelectUser = (userId) => {
    if (!userId) return;

    setSelectedUsers(prev => {
      if (prev.includes(userId)) {
        return prev.filter(id => id !== userId);
      }
      return [...prev, userId];
    });
  };

  const handleDeleteClick = (user) => {
    if (!user || !user.id) return;
    setUserToDelete(user);
    setIsDeleteModalOpen(true);
  };

  const handleDeleteSuccess = (deletedUserId) => {
    if (!deletedUserId) return;
    setUsers(prevUsers => prevUsers.filter(user => user.id !== deletedUserId));
    setSelectedUsers(prevSelected => 
      prevSelected.filter(id => id !== deletedUserId)
    );
  };

  const handleViewClick = (user) => {
    if (!user || !user.id) return;
    setUserToView(user.id);
    setIsViewModalOpen(true);
    onUserViewed(user);
  };

  const handleUpdateClick = (user) => {
    if (!user || !user.id) return;
    setUserToUpdate(user.id);
    setCurrentUserToUpdate(user);
    setIsUpdateModalOpen(true);
  };

  const handleUpdateSuccess = (updatedUser) => {
    if (!updatedUser || !updatedUser.id) return;
    
    setUsers(prevUsers => 
      prevUsers.map(user => 
        user.id === updatedUser.id ? { ...user, ...updatedUser } : user
      )
    );
    setIsUpdateModalOpen(false);
    setUserToUpdate(null);
    setCurrentUserToUpdate(null);
  };

  const ActionMenu = ({ user }) => {
    if (!user || !user.id) return null;

    return (
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
          <div 
            className="absolute right-0 mt-2 w-40 bg-white rounded-md shadow-lg z-50 border"
            onBlur={() => setActiveMenu(null)}
          >
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
  };

  ActionMenu.propTypes = {
    user: PropTypes.shape({
      id: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
      username: PropTypes.string.isRequired,
      subscription_status: PropTypes.string.isRequired,
    }).isRequired,
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="animate-spin mr-2" />
        <span>Loading users...</span>
      </div>
    );
  }

  const headerText = selectedType === 'all' ? 'All Users' :
                    selectedType === 'premium' ? 'Premium End Users' :
                    selectedType === 'normal' ? 'End Users' : 'Users';

  return (
    <>
      <div className="space-y-8">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-md flex items-center">
            <AlertCircle className="mr-2" size={16} />
            {error}
          </div>
        )}

        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold">{headerText}</h2>
          </div>

          <div className="bg-white rounded-lg shadow">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">
                    <input
                      type="checkbox"
                      onChange={handleSelectAll}
                      checked={selectedUsers.length === filteredUsers.length && filteredUsers.length > 0}
                      className="rounded border-gray-300"
                      aria-label="Select all users"
                    />
                  </th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">UserID</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Name</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Account Type</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Last Updated</th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredUsers.map(user => {
                  if (!user || !user.id) return null;
                  
                  return (
                    <tr key={user.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4">
                        <input
                          type="checkbox"
                          checked={selectedUsers.includes(user.id)}
                          onChange={() => handleSelectUser(user.id)}
                          className="rounded border-gray-300"
                          aria-label={`Select user ${user.username || ''}`}
                        />
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-900">{user.id}</td>
                      <td className="px-6 py-4 text-sm text-gray-900">
                        {user.username || 'N/A'}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-900">
                        {user.subscription_status === 'premium' ? 'Premium' : 'Normal'}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-900">
                        {user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'Never'}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <ActionMenu user={user} />
                      </td>
                    </tr>
                  );
                })}
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

      <DeleteUserAction 
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setUserToDelete(null);
        }}
        onDeleteSuccess={handleDeleteSuccess}
        userToDelete={userToDelete}
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
  onUserViewed: PropTypes.func.isRequired,  
};

export default ViewUserAccounts;