import React, { useState, useEffect } from 'react';
import { File, Loader } from 'lucide-react';
import FileActions from './FileActions';

const SortHeader = ({ label, field, currentSort, onSort }) => {
    const isActive = currentSort.field === field;
    
    return (
        <button
            onClick={() => onSort(field)}
            className={`flex items-center space-x-1 hover:text-gray-900 ${
                isActive ? 'text-gray-900' : 'text-gray-600'
            }`}
        >
            <span>{label}</span>
            <span className="text-gray-400">
                {isActive && (currentSort.direction === 'asc' ? '↑' : '↓')}
            </span>
        </button>
    );
};

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
    const [isSelectionMode, setIsSelectionMode] = useState(false);
    const [sortField, setSortField] = useState('created_at');
    const [sortDirection, setSortDirection] = useState('desc');

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
            setIsSelectionMode(false);
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

    const getSelectedFiles = () => {
        return filteredFiles.filter(file => selectedFiles.has(file.id));
    };

    const handleSort = (field) => {
        setSortDirection(current => 
            sortField === field 
                ? current === 'asc' ? 'desc' : 'asc' 
                : 'desc'
        );
        setSortField(field);
    };

    const sortFiles = (files, field, direction) => {
        return [...files].sort((a, b) => {
            let aVal, bVal;
            switch (field) {
                case 'name':
                    aVal = (a.original_name || a.name).toLowerCase();
                    bVal = (b.original_name || b.name).toLowerCase();
                    break;
                case 'size':
                    aVal = a.size;
                    bVal = b.size;
                    break;
                case 'folder':
                    aVal = (a.folder_name || '').toLowerCase();
                    bVal = (b.folder_name || '').toLowerCase();
                    break;
                case 'created_at':
                    aVal = new Date(a.created_at);
                    bVal = new Date(b.created_at);
                    break;
                default:
                    return 0;
            }
            
            if (direction === 'asc') {
                return aVal > bVal ? 1 : -1;
            }
            return aVal < bVal ? 1 : -1;
        });
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
        filteredFiles = filteredFiles.filter(file => !file.is_deleted);
    } else {
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

    // Apply sorting
    filteredFiles = sortFiles(filteredFiles, sortField, sortDirection);

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
                    {selectedFiles.size > 0 && (
                        <div className="bg-gray-100 p-4 flex justify-between items-center border-b">
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
                                <SortHeader 
                                    label="Name"
                                    field="name"
                                    currentSort={{ field: sortField, direction: sortDirection }}
                                    onSort={handleSort}
                                />
                            </div>
                        </div>
                        <div className="col-span-2">
                            <SortHeader 
                                label="Size"
                                field="size"
                                currentSort={{ field: sortField, direction: sortDirection }}
                                onSort={handleSort}
                            />
                        </div>
                        <div className="col-span-2">
                            <SortHeader 
                                label="Location"
                                field="folder"
                                currentSort={{ field: sortField, direction: sortDirection }}
                                onSort={handleSort}
                            />
                        </div>
                        <div className="col-span-2">
                            <SortHeader 
                                label="Last Modified"
                                field="created_at"
                                currentSort={{ field: sortField, direction: sortDirection }}
                                onSort={handleSort}
                            />
                        </div>
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
                                <div className="col-span-2">{file.folder_name || '-'}</div>
                                <div className="col-span-2">{formatDate(file.created_at)}</div>
                                {showActions && (
                                    <div className="col-span-1">
                                        <FileActions 
                                            file={file} 
                                            user={user} 
                                            onRefresh={fetchFiles}
                                            onAction={handleFileAction}
                                            isSelectable={isSelectionMode}
                                            selected={selectedFiles.has(file.id)}
                                            onSelect={() => handleFileSelection(file.id)}
                                            selectedFiles={getSelectedFiles()}
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