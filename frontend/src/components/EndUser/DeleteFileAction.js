import React, { useState } from 'react';
import { Trash2, Loader } from 'lucide-react';

const DeleteFileAction = ({ file, selectedFiles = [], onRefresh, onClearSelection }) => {
    const [isDeleting, setIsDeleting] = useState(false);

    const handleDelete = async () => {
        const files = selectedFiles.length > 0 ? selectedFiles : [file];
        const confirmMessage = `Are you sure you want to delete ${files.length > 1 ? 'these files' : 'this file'}?`;
        
        if (!window.confirm(confirmMessage)) return;

        setIsDeleting(true);

        try {
            const token = localStorage.getItem('token');
            
            if (files.length === 1) {
                const response = await fetch(`http://localhost:8080/api/files/${files[0].id}`, {
                    method: 'DELETE',
                    headers: { 'Authorization': `Bearer ${token}` },
                });
                
                if (!response.ok) throw new Error('Failed to delete file');
            } else {
                const response = await fetch('http://localhost:8080/api/files/mass-delete', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        file_ids: files.map(f => f.id)
                    })
                });
                
                if (!response.ok) throw new Error('Failed to delete files');
            }

            onClearSelection?.();
            onRefresh?.();
        } catch (error) {
            console.error('Delete error:', error);
            alert(error.message || 'Failed to delete file(s)');
        } finally {
            setIsDeleting(false);
        }
    };

    return (
        <button
            onClick={handleDelete}
            disabled={isDeleting}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2 text-red-600 disabled:opacity-50"
        >
            {isDeleting ? (
                <Loader size={16} className="animate-spin" />
            ) : (
                <Trash2 size={16} />
            )}
            <span>Delete {selectedFiles.length > 0 ? `(${selectedFiles.length})` : ''}</span>
        </button>
    );
};

export default DeleteFileAction;