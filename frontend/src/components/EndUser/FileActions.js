import React, { useState } from 'react';
import { Download, Trash2, Share2, Archive, MoreVertical } from 'lucide-react';
import DownloadFileAction from './DownloadFileAction';
import DeleteFileAction from './DeleteFileAction';
import ShareFileAction from './ShareFileAction';
import ArchiveFileAction from './ArchiveFileAction';

const FileActions = ({ file }) => {
    const [showActions, setShowActions] = useState(false);

    return (
        <div className="relative">
            <button 
                onClick={() => setShowActions(!showActions)}
                className="p-1 hover:bg-gray-100 rounded transition-colors"
                aria-label="Show actions"
            >
                <MoreVertical size={16} />
            </button>

            {showActions && (
                <div className="absolute right-0 mt-2 py-2 w-48 bg-white rounded-md shadow-xl z-20 border">
                    <DownloadFileAction file={file} />
                    <ShareFileAction file={file} />
                    <ArchiveFileAction file={file} />
                    <DeleteFileAction file={file} />
                </div>
            )}
        </div>
    );
};

export default FileActions;