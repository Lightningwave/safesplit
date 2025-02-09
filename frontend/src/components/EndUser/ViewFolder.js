import React, { useState, useEffect } from 'react';
import { Folder, MoreVertical, ChevronLeft, File, Loader } from 'lucide-react';
import FileActions from './FileActions';

const ViewFolder = ({ 
    currentFolder,
    onFolderClick,
    onFolderDelete,
    onBackClick,
    selectedSection,
    user,
    showActions = true,
    refreshTrigger = 0
}) => {
    const [folders, setFolders] = useState([]);
    const [files, setFiles] = useState([]);
    const [error, setError] = useState(null);
    const [loading, setLoading] = useState(true);
    const [selectedFiles, setSelectedFiles] = useState(new Set());
    const [isSelectionMode, setIsSelectionMode] = useState(false);

    const fetchContents = async () => {
        try {
            setLoading(true);
            const token = localStorage.getItem('token');
            const endpoint = currentFolder
                ? `/api/folders/${currentFolder.id}`
                : '/api/folders';

            const response = await fetch(endpoint, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            if (currentFolder) {
                setFolders(data.data.folder.sub_folders || []);
                setFiles(data.data.folder.files || []);
            } else {
                setFolders(data.data.folders || []);
                setFiles(data.data.files || []);
            }

            setError(null);
        } catch (err) {
            console.error('Fetch error:', err);
            setError('Failed to load folders. Please try again later.');
            setFolders([]);
            setFiles([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchContents();
        setSelectedFiles(new Set());
        setIsSelectionMode(false);
    }, [currentFolder, selectedSection, refreshTrigger]);

    const formatFileSize = (bytes) => {
        if (!bytes) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    };

    const handleFileSelection = (fileId) => {
        setSelectedFiles(prev => {
            const newSelected = new Set(prev);
            if (newSelected.has(fileId)) {
                newSelected.delete(fileId);
            } else {
                newSelected.add(fileId);
            }
            return newSelected;
        });
    };

    const handleBulkSelection = (event) => {
        if (event.target.checked) {
            const visibleFileIds = filteredFiles.map(file => file.id);
            setSelectedFiles(new Set(visibleFileIds));
        } else {
            setSelectedFiles(new Set());
        }
    };

    const getSelectedFiles = () => {
        return filteredFiles.filter(file => selectedFiles.has(file.id));
    };

    // Filter files based on selectedSection
    const filteredFiles = files.filter(file => {
        if (selectedSection === 'Archives') {
            return file.is_archived;
        } else if (selectedSection === 'Shared Files') {
            return file.is_shared && !file.is_archived;
        }
        return !file.is_archived;
    });

    if (loading) {
        return (
            <div className="flex items-center justify-center text-gray-600 py-8">
                <Loader className="animate-spin mr-2" />
                <span>Loading contents...</span>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-4 mb-4 text-red-700 bg-red-100 rounded-md flex items-center justify-between">
                <div>{error}</div>
                <button 
                    onClick={fetchContents}
                    className="px-3 py-1 bg-red-700 text-white rounded-md text-sm hover:bg-red-800"
                >
                    Retry
                </button>
            </div>
        );
    }

    return (
        <div className="space-y-8">
            {currentFolder && (
                <div className="mb-4">
                    <button
                        onClick={onBackClick}
                        className="flex items-center text-gray-600 hover:text-gray-900"
                    >
                        <ChevronLeft size={20} className="mr-1" />
                        Back
                    </button>
                </div>
            )}

            {selectedFiles.size > 0 && (
                <div className="bg-gray-100 p-4 flex justify-between items-center rounded-lg">
                    <span className="text-sm text-gray-600">
                        {selectedFiles.size} files selected
                    </span>
                    <button
                        onClick={() => setSelectedFiles(new Set())}
                        className="px-4 py-2 text-gray-600 hover:text-gray-800"
                    >
                        Clear Selection
                    </button>
                </div>
            )}

            {folders.length === 0 && filteredFiles.length === 0 ? (
                <div className="p-4 text-gray-500 bg-gray-50 rounded-md text-center">
                    {currentFolder 
                        ? `No contents found in "${currentFolder.name}"`
                        : 'No contents found. Create a folder or upload files to get started.'}
                </div>
            ) : (
                <>
                    {folders.length > 0 && (
                        <div>
                            <h3 className="text-lg font-medium mb-4">Folders</h3>
                            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
                            {folders.map((folder) => (
                                    <div
                                        key={folder.id}
                                        className="group relative bg-white rounded-lg border border-gray-200 hover:border-gray-300 hover:shadow-md transition-all duration-200"
                                    >
                                        <button
                                            onClick={() => onFolderClick(folder)}
                                            className="w-full h-full flex flex-col items-center p-4"
                                        >
                                            <div className="w-full aspect-square flex items-center justify-center">
                                                <div className="bg-gray-50 rounded-lg p-4 group-hover:bg-gray-100 transition-colors">
                                                    <Folder className="w-12 h-12 text-gray-600" />
                                                </div>
                                            </div>
                                            <div className="w-full mt-2 text-center">
                                                <span className="text-sm font-medium text-gray-700 line-clamp-2">
                                                    {folder.name}
                                                </span>
                                            </div>
                                        </button>
                                        {showActions && (
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    onFolderDelete(folder);
                                                }}
                                                className="absolute top-2 right-2 p-1.5 bg-white rounded-full shadow 
                                                        opacity-0 group-hover:opacity-100 transition-opacity 
                                                        hover:bg-gray-50"
                                                title="Delete folder"
                                            >
                                                <MoreVertical size={14} className="text-gray-500" />
                                            </button>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {filteredFiles.length > 0 && (
                        <div>
                            <h3 className="text-lg font-medium mb-4">Files</h3>
                            <div className="border rounded-lg shadow-sm">
                                <div className="grid grid-cols-12 gap-4 p-4 border-b bg-gray-50 text-sm font-medium">
                                    <div className="col-span-5 flex items-center space-x-4">
                                        <input 
                                            type="checkbox"
                                            onChange={handleBulkSelection}
                                            checked={filteredFiles.length > 0 && selectedFiles.size === filteredFiles.length}
                                            className="rounded"
                                        />
                                        <span>Name</span>
                                    </div>
                                    <div className="col-span-2">Size</div>
                                    <div className="col-span-3">Last Modified</div>
                                    <div className="col-span-2">Actions</div>
                                </div>
                                {filteredFiles.map((file) => (
                                    <div key={file.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50">
                                        <div className="col-span-5 flex items-center space-x-4">
                                            <input 
                                                type="checkbox"
                                                checked={selectedFiles.has(file.id)}
                                                onChange={() => handleFileSelection(file.id)}
                                                className="rounded"
                                            />
                                            <File size={20} className="text-gray-400" />
                                            <span className="truncate" title={file.original_name || file.name}>
                                                {file.original_name || file.name}
                                            </span>
                                        </div>
                                        <div className="col-span-2">
                                            {formatFileSize(file.size)}
                                        </div>
                                        <div className="col-span-3">
                                            {formatDate(file.created_at)}
                                        </div>
                                        <div className="col-span-2">
                                            <FileActions 
                                                file={file}
                                                user={user}
                                                onRefresh={fetchContents}
                                                isSelectable={isSelectionMode}
                                                selected={selectedFiles.has(file.id)}
                                                onSelect={() => handleFileSelection(file.id)}
                                                selectedFiles={getSelectedFiles()}
                                            />
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </>
            )}
        </div>
    );
};

export default ViewFolder;