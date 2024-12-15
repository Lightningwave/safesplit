import React from 'react';
import PropTypes from 'prop-types';
import { AlertCircle, Loader2 } from 'lucide-react';

const DeleteUserAction = ({ 
  isOpen, 
  onClose, 
  onDeleteSuccess, 
  userToDelete 
}) => {
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState(null);

  const handleDelete = async () => {
    if (!userToDelete) return;

    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch(`http://localhost:8080/api/system/users/${userToDelete.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete user');
      }

      onDeleteSuccess(userToDelete.id);
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Clear error when modal closes
  React.useEffect(() => {
    if (!isOpen) {
      setError(null);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg p-6 max-w-md w-full mx-4">
        <h2 className="text-lg font-semibold mb-4">Delete User</h2>

        {error && (
          <div className="mb-4 bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-md flex items-center">
            <AlertCircle className="mr-2" size={16} />
            {error}
          </div>
        )}

        <p className="text-gray-600 mb-6">
          Are you sure you want to delete user <span className="font-semibold">{userToDelete?.username}</span>?
          This action cannot be undone.
        </p>

        <div className="flex justify-end space-x-3">
          <button
            onClick={onClose}
            disabled={loading}
            className="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md"
          >
            Cancel
          </button>
          <button
            onClick={handleDelete}
            disabled={loading}
            className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md flex items-center"
          >
            {loading && <Loader2 className="animate-spin mr-2" size={16} />}
            Delete
          </button>
        </div>
      </div>
    </div>
  );
};

DeleteUserAction.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onDeleteSuccess: PropTypes.func.isRequired,
  userToDelete: PropTypes.shape({
    id: PropTypes.number.isRequired,
    username: PropTypes.string.isRequired,
  }),
};

export default DeleteUserAction;