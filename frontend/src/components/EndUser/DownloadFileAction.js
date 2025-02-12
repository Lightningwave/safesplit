import React, { useState } from 'react';
import { Download, Loader } from 'lucide-react';

const DownloadFileAction = ({ file, selectedFiles = [], onClearSelection, onClose }) => {
    const [isDownloading, setIsDownloading] = useState(false);

    const handleDownload = async () => {
        setIsDownloading(true);
        try {
            const token = localStorage.getItem('token');
            const filesToDownload = selectedFiles.length > 0 ? selectedFiles : [file];

            if (selectedFiles.length > 0) {
                const statusResponse = await fetch('/api/files/mass-download', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        file_ids: filesToDownload.map(f => f.id)
                    })
                });

                if (!statusResponse.ok) throw new Error('Failed to get download status');

                const statusResult = await statusResponse.json();
                const availableFiles = statusResult.data.download_status.filter(
                    result => result.status === 'success'
                );

                for (const fileStatus of availableFiles) {
                    try {
                        const response = await fetch(`/api/files/mass-download/${fileStatus.file_id}`, {
                            headers: {
                                'Authorization': `Bearer ${token}`,
                            },
                        });

                        if (!response.ok) continue;

                        const blob = await response.blob();
                        const url = window.URL.createObjectURL(blob);
                        const a = document.createElement('a');
                        a.href = url;
                        a.download = fileStatus.file_name;
                        document.body.appendChild(a);
                        a.click();
                        window.URL.revokeObjectURL(url);
                        document.body.removeChild(a);
                    } catch (error) {
                        console.error(`Error downloading file ${fileStatus.file_id}:`, error);
                    }
                }

                onClearSelection?.();
            } else {
                const response = await fetch(`/api/files/${file.id}/download`, {
                    headers: {
                        'Authorization': `Bearer ${token}`,
                    },
                });

                if (!response.ok) throw new Error('Download failed');

                const blob = await response.blob();
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = file.original_name || file.name;
                document.body.appendChild(a);
                a.click();
                window.URL.revokeObjectURL(url);
                document.body.removeChild(a);
            }
            
            onClose?.();
        } catch (error) {
            console.error('Download error:', error);
            alert(error.message || 'Failed to download file(s)');
        } finally {
            setIsDownloading(false);
        }
    };

    return (
        <button
            onClick={handleDownload}
            disabled={isDownloading}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2 disabled:opacity-50"
        >
            {isDownloading ? (
                <Loader size={16} className="animate-spin" />
            ) : (
                <Download size={16} />
            )}
            <span>
                {selectedFiles.length > 0 
                    ? `Download ${selectedFiles.length} Files` 
                    : 'Download'}
            </span>
        </button>
    );
};

export default DownloadFileAction;