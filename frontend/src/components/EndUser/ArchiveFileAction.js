import React from 'react';
import { Archive } from 'lucide-react';

const ArchiveFileAction = ({ file }) => {
    const handleArchive = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`http://localhost:8080/api/files/${file.id}/archive`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) throw new Error('Archive failed');
            
            // Refresh file list after successful archive
            window.location.reload();
        } catch (error) {
            console.error('Archive error:', error);
        }
    };

    return (
        <button
            onClick={handleArchive}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2"
        >
            <Archive size={16} />
            <span>Archive</span>
        </button>
    );
};

export default ArchiveFileAction;