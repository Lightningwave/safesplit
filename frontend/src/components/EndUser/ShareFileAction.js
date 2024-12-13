import React from 'react';
import { Share2 } from 'lucide-react';

const ShareFileAction = ({ file }) => {
    const handleShare = async () => {
        try {
            const token = localStorage.getItem('token');
            // Implement share functionality here
            console.log('Sharing file:', file);
        } catch (error) {
            console.error('Share error:', error);
        }
    };

    return (
        <button
            onClick={handleShare}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2"
        >
            <Share2 size={16} />
            <span>Share</span>
        </button>
    );
};

export default ShareFileAction;