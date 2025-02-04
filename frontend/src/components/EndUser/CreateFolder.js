import React, { useState } from 'react';
import { X } from 'lucide-react';

const CreateFolder = ({ isOpen, onClose, currentFolder, onFolderCreated }) => {
    const [folderName, setFolderName] = useState('');
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/api/folders', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    name: folderName,
                    parent_folder_id: currentFolder ? currentFolder.id : null
                })
            });

            const data = await response.json();
            
            if (data.status === 'success') {
                setFolderName('');
                onFolderCreated();
                onClose();
            } else {
                throw new Error(data.error || 'Failed to create folder');
            }
        } catch (error) {
            console.error('Error creating folder:', error);
            setError(error.message || 'Failed to create folder. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <div className="flex justify-between items-center mb-4">
                    <h2 className="text-xl font-semibold">Create New Folder</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-500 hover:text-gray-700"
                    >
                        <X size={20} />
                    </button>
                </div>

                <form onSubmit={handleSubmit}>
                    <div className="mb-4">
                        <label htmlFor="folderName" className="block text-sm font-medium text-gray-700 mb-1">
                            Folder Name
                        </label>
                        <input
                            type="text"
                            id="folderName"
                            value={folderName}
                            onChange={(e) => setFolderName(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            placeholder="Enter folder name"
                            required
                        />
                    </div>

                    {error && (
                        <div className="mb-4 text-red-500 text-sm">
                            {error}
                        </div>
                    )}

                    <div className="flex justify-end space-x-3">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-gray-600 hover:text-gray-800"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isLoading}
                            className={`px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 
                                ${isLoading ? 'opacity-50 cursor-not-allowed' : ''}`}
                        >
                            {isLoading ? 'Creating...' : 'Create Folder'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default CreateFolder;