import React, { useState, useEffect } from 'react';
import { File, Loader } from 'lucide-react';
import FileActions from './FileActions';

const ViewFile = ({ 
    searchQuery, 
    user, 
    selectedSection, 
    currentFolder, 
    showRecentsOnly = false,
    maxItems = null,
    refreshTrigger = 0
}) => {
    const [files, setFiles] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [selectedFiles, setSelectedFiles] = useState(new Set());

    const fetchFiles = async () => {
        try {
            const token = localStorage.getItem('token');
            if (!token) {
                setError('Authentication required');
                return;
            }

            let url = 'http://localhost:8080/api/files';
            if (currentFolder) {
                url += `?folder_id=${currentFolder.id}`;
            }

            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to fetch files');
            }

            const data = await response.json();
            let filesList = data.data?.files || [];

            // Sort by last modified date
            filesList = filesList.sort((a, b) => 
                new Date(b.created_at) - new Date(a.created_at)
            );

            // If maxItems is set, limit the number of files
            if (maxItems && maxItems > 0) {
                filesList = filesList.slice(0, maxItems);
            }

            setFiles(filesList);
            setError('');
        } catch (error) {
            console.error('Error fetching files:', error);
            setError(error.message || 'Failed to load files');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (user) {
            fetchFiles();
            setSelectedFiles(new Set());
        }
    }, [user, currentFolder, selectedSection, refreshTrigger]);

    const handleFileAction = async (action, file) => {
        switch (action) {
            case 'delete':
            case 'archive':
            case 'share':
                await fetchFiles();
                break;
            default:
                break;
        }
    };

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

    // Update the section filtering logic
    let filteredFiles = files;

    // Apply search filter
    if (searchQuery) {
        filteredFiles = filteredFiles.filter(file => 
            (file.original_name || file.name)
                .toLowerCase()
                .includes(searchQuery.toLowerCase())
        );
    }

    // Apply section filters
    if (currentFolder) {
        // When viewing a folder, show all non-deleted files regardless of archive status
        filteredFiles = filteredFiles.filter(file => !file.is_deleted);
    } else {
        // Only apply archive/shared filters when not in a folder view
        if (selectedSection === 'Archives') {
            filteredFiles = filteredFiles.filter(file => file.is_archived === true);
        } else if (selectedSection === 'Shared Files') {
            filteredFiles = filteredFiles.filter(file => 
                file.is_shared === true && file.is_archived === false
            );
        } else if (selectedSection !== 'Dashboard') {
            filteredFiles = filteredFiles.filter(file => file.is_archived === false);
        }
    }
    

    const showCheckboxes = !showRecentsOnly;
    const showActions = !showRecentsOnly;

    if (!user) {
        return (
            <div className="text-red-500 text-center p-8">
                Please log in to view files
            </div>
        );
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center p-8">
                <Loader className="animate-spin mr-2" />
                <span>Loading files...</span>
            </div>
        );
    }

    return (
        <div className="border rounded-lg shadow-sm">
            {error ? (
                <div className="text-red-500 text-center p-8">
                    {error}
                </div>
            ) : (
                <>
                    <div className="grid grid-cols-12 gap-4 p-4 border-b bg-gray-50 text-sm font-medium">
                        <div className="col-span-5">
                            <div className="flex items-center space-x-4">
                                {showCheckboxes && (
                                    <input 
                                        type="checkbox"
                                        onChange={handleBulkSelection}
                                        checked={filteredFiles.length > 0 && selectedFiles.size === filteredFiles.length}
                                        className="rounded"
                                    />
                                )}
                                <span>Name</span>
                            </div>
                        </div>
                        <div className="col-span-2">Size</div>
                        <div className="col-span-2">Location</div>
                        <div className="col-span-2">Last Modified</div>
                        {showActions && <div className="col-span-1">Actions</div>}
                    </div>

                    {filteredFiles.length === 0 ? (
                        <div className="text-gray-500 text-center p-8">
                            {selectedSection === 'Shared Files' 
                                ? "No shared files found"
                                : `No files found in ${currentFolder ? `folder "${currentFolder.name}"` : 'this location'}`
                            }
                        </div>
                    ) : (
                        filteredFiles.map(file => (
                            <div key={file.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50 text-sm">
                                <div className="col-span-5">
                                    <div className="flex items-center space-x-4">
                                        {showCheckboxes && (
                                            <input 
                                                type="checkbox"
                                                checked={selectedFiles.has(file.id)}
                                                onChange={() => handleFileSelection(file.id)}
                                                className="rounded"
                                            />
                                        )}
                                        <div className="flex items-center space-x-2">
                                            <File size={20} className="text-gray-400" />
                                        </div>
                                        <span className="truncate" title={file.original_name || file.name}>
                                            {file.original_name || file.name}
                                        </span>
                                    </div>
                                </div>
                                <div className="col-span-2">{formatFileSize(file.size)}</div>
                                <div className="col-span-2">{file.folder_name}</div>
                                <div className="col-span-2">{formatDate(file.created_at)}</div>
                                {showActions && (
                                    <div className="col-span-1">
                                        <FileActions 
                                            file={file} 
                                            user={user} 
                                            onRefresh={fetchFiles}
                                            onAction={handleFileAction}
                                        />
                                    </div>
                                )}
                            </div>
                        ))
                    )}
                </>
            )}
        </div>
    );
};

export default ViewFile;