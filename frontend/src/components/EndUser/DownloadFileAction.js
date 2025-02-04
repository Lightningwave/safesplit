import React from 'react';
import { Download } from 'lucide-react';

const DownloadFileAction = ({ file, selectedFiles = [] }) => {
    const handleDownload = async () => {
        try {
            const token = localStorage.getItem('token');
            const filesToDownload = selectedFiles.length > 0 ? selectedFiles : [file];

            // If multiple files selected, get download status first
            if (selectedFiles.length > 0) {
                const statusResponse = await fetch('http://localhost:8080/api/files/mass-download', {
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

                // Download each available file
                for (const fileStatus of availableFiles) {
                    try {
                        const response = await fetch(`http://localhost:8080/api/files/mass-download/${fileStatus.file_id}`, {
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
            } else {
                // Single file download
                const response = await fetch(`http://localhost:8080/api/files/${file.id}/download`, {
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
        } catch (error) {
            console.error('Download error:', error);
        }
    };

    return (
        <button
            onClick={handleDownload}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2"
        >
            <Download size={16} />
            <span>
                {selectedFiles.length > 0 
                    ? `Download ${selectedFiles.length} Files` 
                    : 'Download'}
            </span>
        </button>
    );
};

export default DownloadFileAction;