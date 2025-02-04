import React from 'react';
import { Trash2 } from 'lucide-react';

const DeleteFileAction = ({ file, selectedFiles = [] }) => {
    const handleDelete = async () => {
        const files = selectedFiles.length > 0 ? selectedFiles : [file];
        const confirmMessage = `Are you sure you want to delete ${files.length > 1 ? 'these files' : 'this file'}?`;
        
        if (!window.confirm(confirmMessage)) return;

        try {
            const token = localStorage.getItem('token');
            
            if (files.length === 1) {
                await fetch(`/api/files/${files[0].id}`, {
                    method: 'DELETE',
                    headers: { 'Authorization': `Bearer ${token}` },
                });
            } else {
                await fetch('/api/files/mass-delete', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        file_ids: files.map(f => f.id)
                    })
                });
            }

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
            <span>Delete {selectedFiles.length > 0 ? `(${selectedFiles.length})` : ''}</span>
        </button>
    );
};

export default DeleteFileAction;