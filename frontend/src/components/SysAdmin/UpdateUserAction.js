import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { X } from 'lucide-react';

const UpdateUserAction = ({ isOpen, onClose, userId, currentUser, onUpdateSuccess }) => {
  const initialFormState = {
    username: '',
    email: '',
    account_type: 'free',
    read_access: false,
    write_access: false
  };

  const [formData, setFormData] = useState(initialFormState);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (isOpen && currentUser) {
      setFormData({
        username: currentUser.username || '',
        email: currentUser.email || '',
        account_type: currentUser.subscription_status || 'free',
        read_access: currentUser.read_access || false,
        write_access: currentUser.write_access || false
      });
    }
    return () => {
      setFormData(initialFormState);
      setError(null);
      setLoading(false);
    };
  }, [isOpen, currentUser]);

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/api/system/users/${userId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          username: formData.username,
          email: formData.email,
          account_type: formData.account_type,
          read_access: formData.read_access,
          write_access: formData.write_access
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to update user account');
      }

      onUpdateSuccess(data.user);
      onClose();
    } catch (err) {
      console.error('Error updating user:', err);
      setError(err.message || 'An error occurred while updating the user account');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-white rounded-lg shadow-lg w-11/12 md:w-2/3 lg:w-1/2 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center p-4 border-b">
          <h3 className="text-xl font-semibold">Update User Account</h3>
          <button onClick={onClose} className="hover:bg-gray-100 p-2 rounded-full" aria-label="Close modal">
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6">
          {loading && (
            <div className="flex items-center justify-center py-4">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            </div>
          )}

          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative mb-4">
              <p>{error}</p>
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-gray-700">Username</label>
              <input
                type="text"
                id="username"
                name="username"
                value={formData.username}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                required
                minLength={3}
                maxLength={255}
              />
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">Email</label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                required
              />
            </div>

            <div>
              <label htmlFor="account_type" className="block text-sm font-medium text-gray-700">Account Type</label>
              <select
                id="account_type"
                name="account_type"
                value={formData.account_type}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
              >
                <option value="free">Free</option>
                <option value="premium">Premium</option>
              </select>
            </div>

            <div className="space-y-2">
              <div className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  id="read_access"
                  name="read_access"
                  checked={formData.read_access}
                  onChange={handleInputChange}
                  className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label htmlFor="read_access" className="text-sm font-medium text-gray-700">Read Access</label>
              </div>

              <div className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  id="write_access"
                  name="write_access"
                  checked={formData.write_access}
                  onChange={handleInputChange}
                  className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label htmlFor="write_access" className="text-sm font-medium text-gray-700">Write Access</label>
              </div>
            </div>
          </div>

          <div className="flex justify-end space-x-3 mt-6 pt-4 border-t">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {loading ? 'Updating...' : 'Update Account'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

UpdateUserAction.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  userId: PropTypes.string,
  currentUser: PropTypes.object,
  onUpdateSuccess: PropTypes.func.isRequired,
};

UpdateUserAction.defaultProps = {
  userId: null,
  currentUser: null,
};

export default UpdateUserAction;