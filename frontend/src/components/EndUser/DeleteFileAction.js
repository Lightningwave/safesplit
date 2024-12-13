import React from 'react';
import { Trash2 } from 'lucide-react';

const DeleteFileAction = ({ file }) => {
    const handleDelete = async () => {
        if (!window.confirm('Are you sure you want to delete this file?')) return;

        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`http://localhost:8080/api/files/${file.id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (!response.ok) throw new Error('Delete failed');
            
            // Refresh file list after successful deletion
            window.location.reload();
        } catch (error) {
            console.error('Delete error:', error);
        }
    };

    return (
        <button
            onClick={handleDelete}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2 text-red-600"
        >
            <Trash2 size={16} />
            <span>Delete</span>
        </button>
    );
};

export default DeleteFileAction;