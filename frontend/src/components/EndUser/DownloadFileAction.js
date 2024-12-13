import React from 'react';
import { Download } from 'lucide-react';

const DownloadFileAction = ({ file }) => {
    const handleDownload = async () => {
        try {
            const token = localStorage.getItem('token');
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
            <span>Download</span>
        </button>
    );
};

export default DownloadFileAction;