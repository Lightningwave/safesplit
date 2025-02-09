import React, { useState } from 'react';
import { Archive, Loader } from 'lucide-react';

const UnarchiveFileAction = ({ file, selectedFiles = [], onRefresh }) => {
    const [isUnarchiving, setIsUnarchiving] = useState(false);

    const handleUnarchive = async () => {
        const files = selectedFiles.length > 0 ? selectedFiles : [file];
        const confirmMessage = `Are you sure you want to unarchive ${files.length > 1 ? `these ${files.length} files` : 'this file'}?`;
        
        if (!window.confirm(confirmMessage)) return;
        
        setIsUnarchiving(true);

        try {
            const token = localStorage.getItem('token');
            
            if (files.length === 1) {
                const response = await fetch(`http://localhost:8080/api/files/${files[0].id}/unarchive`, {
                    method: 'PUT',
                    headers: { 
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    }
                });
                
                if (!response.ok) throw new Error('Failed to unarchive file');
            } else {
                const response = await fetch('http://localhost:8080/api/files/mass-unarchive', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        file_ids: files.map(f => f.id)
                    })
                });
                
                if (!response.ok) throw new Error('Failed to unarchive files');
                
                const result = await response.json();
                const failedUnarchives = result.data?.unarchive_status?.filter(
                    status => status.status === 'error'
                );
                
                if (failedUnarchives?.length > 0) {
                    throw new Error(`Failed to unarchive ${failedUnarchives.length} files`);
                }
            }

            onRefresh?.();
        } catch (error) {
            console.error('Unarchive error:', error);
            alert(error.message || 'Failed to unarchive file(s)');
        } finally {
            setIsUnarchiving(false);
        }
    };

    return (
        <button
            onClick={handleUnarchive}
            disabled={isUnarchiving}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2 disabled:opacity-50"
        >
            {isUnarchiving ? (
                <Loader size={16} className="animate-spin" />
            ) : (
                <Archive size={16} />
            )}
            <span>Unarchive{selectedFiles.length > 0 ? ` (${selectedFiles.length})` : ''}</span>
        </button>
    );
};

export default UnarchiveFileAction;