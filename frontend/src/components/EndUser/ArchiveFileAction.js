import React, { useState } from 'react';
import { Archive, Loader } from 'lucide-react';

const ArchiveFileAction = ({ file, selectedFiles = [], onRefresh }) => {
    const [isArchiving, setIsArchiving] = useState(false);

    const handleArchive = async () => {
        const files = selectedFiles.length > 0 ? selectedFiles : [file];
        const confirmMessage = `Are you sure you want to archive ${files.length > 1 ? `these ${files.length} files` : 'this file'}?`;
        
        if (!window.confirm(confirmMessage)) return;
        
        setIsArchiving(true);

        try {
            const token = localStorage.getItem('token');
            
            if (files.length === 1) {
                const response = await fetch(`http://localhost:8080/api/files/${files[0].id}/archive`, {
                    method: 'PUT',
                    headers: { 
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    }
                });
                
                if (!response.ok) throw new Error('Failed to archive file');
            } else {
                const response = await fetch('http://localhost:8080/api/files/mass-archive', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        file_ids: files.map(f => f.id)
                    })
                });
                
                if (!response.ok) throw new Error('Failed to archive files');
                
                const result = await response.json();
                const failedArchives = result.data?.archive_status?.filter(
                    status => status.status === 'error'
                );
                
                if (failedArchives?.length > 0) {
                    throw new Error(`Failed to archive ${failedArchives.length} files`);
                }
            }

            onRefresh?.();
        } catch (error) {
            console.error('Archive error:', error);
            alert(error.message || 'Failed to archive file(s)');
        } finally {
            setIsArchiving(false);
        }
    };

    return (
        <button
            onClick={handleArchive}
            disabled={isArchiving}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2 disabled:opacity-50"
        >
            {isArchiving ? (
                <Loader size={16} className="animate-spin" />
            ) : (
                <Archive size={16} />
            )}
            <span>Archive{selectedFiles.length > 0 ? ` (${selectedFiles.length})` : ''}</span>
        </button>
    );
};

export default ArchiveFileAction;