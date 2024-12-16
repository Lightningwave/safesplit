import React, { useState } from 'react';
import { X, AlertTriangle } from 'lucide-react';

const DeleteFolder = ({ isOpen, onClose, folder, onFolderDeleted }) => {
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    const handleDelete = async () => {
        if (!folder) return;
        
        setError('');
        setIsLoading(true);

        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`http://localhost:8080/api/folders/${folder.id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });

            const data = await response.json();

            if (response.ok && data.status === 'success') {
                console.log('Folder deleted:', data.data.folder_path);
                onFolderDeleted();
                onClose();
            } else {
                throw new Error(data.message || 'Failed to delete folder');
            }
        } catch (error) {
            console.error('Error deleting folder:', error);
            setError(error.message || 'Failed to delete folder. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    if (!isOpen || !folder) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <div className="flex justify-between items-center mb-4">
                    <h2 className="text-xl font-semibold">Delete Folder</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-500 hover:text-gray-700"
                        disabled={isLoading}
                    >
                        <X size={20} />
                    </button>
                </div>

                <div className="mb-6">
                    <div className="flex items-center justify-center text-yellow-500 mb-4">
                        <AlertTriangle size={48} />
                    </div>
                    <p className="text-center text-gray-600">
                        Are you sure you want to delete the folder "{folder.name}"?
                        <br />
                        This action cannot be undone and all contents will be deleted.
                    </p>
                </div>

                {error && (
                    <div className="mb-4 text-red-500 text-sm text-center">
                        {error}
                    </div>
                )}

                <div className="flex justify-end space-x-3">
                    <button
                        type="button"
                        onClick={onClose}
                        className="px-4 py-2 text-gray-600 hover:text-gray-800"
                        disabled={isLoading}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleDelete}
                        disabled={isLoading}
                        className={`px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 
                            ${isLoading ? 'opacity-50 cursor-not-allowed' : ''}`}
                    >
                        {isLoading ? 'Deleting...' : 'Delete Folder'}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default DeleteFolder;