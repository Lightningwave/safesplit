import React, { useState, useEffect } from 'react';
import { Folder, MoreVertical, ChevronLeft, File } from 'lucide-react';
import FileActions from './FileActions';
import SortingHeader from './SortingHeader';

const ViewFolder = ({ 
    currentFolder,
    onFolderClick,
    onFolderDelete,
    onBackClick,
    selectedSection,
    showActions = true,
    refreshTrigger = 0
}) => {
    const [folders, setFolders] = useState([]);
    const [files, setFiles] = useState([]);
    const [error, setError] = useState(null);
    const [loading, setLoading] = useState(true);
    const [sortType, setSortType] = useState('date');   // State for sorting type
    const [sortOrder, setSortOrder] = useState('desc'); // State for sorting order

    const fetchContents = async () => {
        try {
            setLoading(true);
            const token = localStorage.getItem('token');
            const endpoint = currentFolder
                ? `http://localhost:8080/api/folders/${currentFolder.id}`
                : 'http://localhost:8080/api/folders';

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

            // Handle different response structures
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

    // Filter files based on selectedSection
    const filteredFiles = files.filter(file => {
        if (selectedSection === 'Archives') {
            return file.is_archived;
        } else if (selectedSection === 'Shared Files') {
            return file.is_shared && !file.is_archived;
        }
        return !file.is_archived;
    });

    // Sorting Logic
    const sortFiles = (files) => {
        if (sortType === 'date') {  // Last Modified (date)
            return files.sort((a, b) => 
                sortOrder === 'desc' ? new Date(b.created_at) - new Date(a.created_at) : new Date(a.created_at) - new Date(b.created_at)
            );
        } else if (sortType === 'name') {   // Name (string)      
            return files.sort((a, b) => {
                const nameA = a.original_name || a.name;
                const nameB = b.original_name || b.name;
                return sortOrder === 'desc' ? nameB.localeCompare(nameA) : nameA.localeCompare(nameB);
            });
        } else if (sortType === 'size') {   // Size (raw size values)
            return files.sort((a, b) => {
                return sortOrder === 'desc' ? b.size - a.size : a.size - b.size;
            });
        } else if (sortType === 'location') {   // Location (string)
            return files.sort((a, b) => {
                const locationA = a.folder_name || '';
                const locationB = b.folder_name || '';
                return sortOrder === 'desc' ? locationB.localeCompare(locationA) : locationA.localeCompare(locationB);
            });
        }
    }
    // Apply sorting
    const sortedFiles = sortFiles(filteredFiles);

    // Change sorting type (name, size, location, date)
    const handleSortChange = (event) => {
        const newSortType = event.target.value;
        setSortType(newSortType);
        setSortOrder('desc'); // Default to descending when changing sorting type
    };

    // Change the sorting order (asc,desc)
    const toggleSortOrder = () => {
        setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center text-gray-600 py-8">
                <div className="animate-spin rounded-full h-6 w-6 border-2 border-gray-500 border-t-transparent mr-2" />
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
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                                {folders.map((folder) => (
                                    <div
                                        key={folder.id}
                                        className="group relative"
                                    >
                                        <button
                                            onClick={() => onFolderClick(folder)}
                                            className="w-full flex items-center p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                                        >
                                            <Folder className="text-gray-400 mr-3" size={24} />
                                            <span className="text-gray-700 font-medium truncate">
                                                {folder.name}
                                            </span>
                                        </button>
                                        {showActions && (
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    onFolderDelete(folder);
                                                }}
                                                className="absolute right-2 top-2 p-2 opacity-0 group-hover:opacity-100 transition-opacity text-gray-500 hover:text-gray-700"
                                            >
                                                <MoreVertical size={16} />
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
                            <div className="border rounded-lg">
                                {/* SORT BY DATE/NAME/SIZE - DROPDOWN SELECTION */}
                                <div className="flex justify-end items-center p-2 bg-gray-50">                
                                    <select 
                                        onChange={handleSortChange} 
                                        className="border rounded p-2 text-sm"
                                    >
                                        <option value="date">Sort by Date</option>
                                        <option value="name">Sort by Name</option>
                                        <option value="size">Sort by Size</option>
                                    </select>
                                </div>
                                <div className="grid grid-cols-12 gap-4 p-4 border-b bg-gray-50 text-sm font-medium">
                                    <div className="col-span-5">
                                        <SortingHeader
                                            label="Name"
                                            sortType="name"
                                            currentSortType={sortType}
                                            sortOrder={sortOrder}
                                            onSortChange={handleSortChange}
                                            onToggleSortOrder={toggleSortOrder}
                                        />
                                    </div>
                                    <div className="col-span-2">
                                        <SortingHeader
                                            label="Size"
                                            sortType="size"
                                            currentSortType={sortType}
                                            sortOrder={sortOrder}
                                            onSortChange={handleSortChange}
                                            onToggleSortOrder={toggleSortOrder}
                                        />
                                    </div>
                                    <div className="col-span-3">
                                        <SortingHeader
                                            label="Last Modified"
                                            sortType="date"
                                            currentSortType={sortType}
                                            sortOrder={sortOrder}
                                            onSortChange={handleSortChange}
                                            onToggleSortOrder={toggleSortOrder}
                                        />
                                    </div>
                                    <div className="col-span-2">Actions</div>
                                </div>
                                {filteredFiles.map((file) => (
                                    <div key={file.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50">
                                        <div className="col-span-5 flex items-center">
                                            <File size={20} className="text-gray-400 mr-3" />
                                            <span className="truncate">
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
                                                onRefresh={fetchContents}
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