import React, { useState, useRef, useEffect } from 'react';
import { Download, Trash2, Share2, Archive, MoreVertical, Check } from 'lucide-react';
import DownloadFileAction from './DownloadFileAction';
import DeleteFileAction from './DeleteFileAction';
import ShareFileAction from './ShareFileAction';
import ArchiveFileAction from './ArchiveFileAction';
import UnarchiveFileAction from './UnarchiveFileAction';
import ReportFileAction from './ReportFileAction';

const FileActions = ({ 
    file, 
    user, 
    onRefresh, 
    onAction, 
    isSelectable = false, 
    selected = false, 
    onSelect, 
    selectedFiles = [],
    onClearSelection
}) => {
    const [showActions, setShowActions] = useState(false);
    const actionsRef = useRef(null);

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (actionsRef.current && !actionsRef.current.contains(event.target)) {
                setShowActions(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

    const handleClick = (e) => {
        if (isSelectable) {
            e.stopPropagation();
            onSelect && onSelect(file);
        } else {
            setShowActions(!showActions);
        }
    };

    const handleClose = () => {
        setShowActions(false);
    };

    const allFilesArchived = selectedFiles.length > 0 
        ? selectedFiles.every(f => f.is_archived)
        : file.is_archived;

    return (
        <div className="relative" ref={actionsRef}>
            <button 
                onClick={handleClick}
                className={`p-1 hover:bg-gray-100 rounded transition-colors ${selected ? 'bg-gray-100' : ''}`}
                aria-label={isSelectable ? "Select file" : "Show actions"}
            >
                {isSelectable ? (
                    <div className={`w-4 h-4 border rounded ${selected ? 'bg-blue-500 border-blue-500' : 'border-gray-400'}`}>
                        {selected && <Check size={16} className="text-white" />}
                    </div>
                ) : (
                    <MoreVertical size={16} />
                )}
            </button>

            {showActions && !isSelectable && (
                <div className="absolute right-0 mt-2 py-2 w-48 bg-white rounded-md shadow-xl z-20 border">
                    <DownloadFileAction 
                        file={file} 
                        selectedFiles={selectedFiles.length > 1 ? selectedFiles : []}
                        onClearSelection={onClearSelection}
                        onClose={handleClose}
                    />
                    <ShareFileAction 
                        file={file} 
                        user={user} 
                        onClose={handleClose}
                    />
                    {allFilesArchived ? (
                        <UnarchiveFileAction 
                            file={file} 
                            selectedFiles={selectedFiles.length > 1 ? selectedFiles : []}
                            onRefresh={onRefresh}
                            onClearSelection={onClearSelection}
                            onClose={handleClose}
                        />
                    ) : (
                        <ArchiveFileAction 
                            file={file} 
                            selectedFiles={selectedFiles.length > 1 ? selectedFiles : []}
                            onRefresh={onRefresh}
                            onClearSelection={onClearSelection}
                            onClose={handleClose}
                        />
                    )}
                    <DeleteFileAction 
                        file={file} 
                        selectedFiles={selectedFiles.length > 1 ? selectedFiles : []}
                        onRefresh={onRefresh}
                        onClearSelection={onClearSelection}
                        onClose={handleClose}
                    />
                    <ReportFileAction 
                        file={file}
                        onClose={handleClose}
                    />
                </div>
            )}
        </div>
    );
};

export default FileActions;