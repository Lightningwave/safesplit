import React, { useState, useEffect } from 'react';
import { File, MoreVertical, Loader } from 'lucide-react';

const ViewFile = ({ searchQuery, user }) => {
    const [files, setFiles] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [selectedFiles, setSelectedFiles] = useState(new Set());

    useEffect(() => {
        if (user) {
            fetchFiles();
        }
    }, [user]);

    const fetchFiles = async () => {
        try {
            const token = localStorage.getItem('token');
            console.log('Fetching files with token:', token);

            if (!token) {
                setError('Authentication required');
                return;
            }

            const response = await fetch('http://localhost:8080/api/files', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to fetch files');
            }

            const data = await response.json();
            setFiles(data.data?.files || []);
            setError('');
        } catch (error) {
            console.error('Error fetching files:', error);
            setError(error.message || 'Failed to load files');
        } finally {
            setLoading(false);
        }
    };

    const filteredFiles = searchQuery
        ? files.filter(file => 
            (file.original_name || file.name)
                .toLowerCase()
                .includes(searchQuery.toLowerCase())
          )
        : files;

    const formatFileSize = (bytes) => {
        if (bytes === 0) return '0 Bytes';
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
            setSelectedFiles(new Set(filteredFiles.map(file => file.id)));
        } else {
            setSelectedFiles(new Set());
        }
    };

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
                                <input 
                                    type="checkbox"
                                    onChange={handleBulkSelection}
                                    checked={selectedFiles.size === filteredFiles.length && filteredFiles.length > 0}
                                    className="rounded"
                                />
                                <span>Name</span>
                            </div>
                        </div>
                        <div className="col-span-2">Size</div>
                        <div className="col-span-2">Folder</div>
                        <div className="col-span-2">Last Modified</div>
                        <div className="col-span-1">Actions</div>
                    </div>

                    {filteredFiles.length === 0 ? (
                        <div className="text-gray-500 text-center p-8">
                            No files found
                        </div>
                    ) : (
                        filteredFiles.map(file => (
                            <div key={file.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50 text-sm">
                                <div className="col-span-5">
                                    <div className="flex items-center space-x-4">
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
                                </div>
                                <div className="col-span-2">{formatFileSize(file.size)}</div>
                                <div className="col-span-2">{file.folder || 'My Files'}</div>
                                <div className="col-span-2">{formatDate(file.created_at)}</div>
                                <div className="col-span-1">
                                    <button 
                                        className="p-1 hover:bg-gray-100 rounded transition-colors"
                                        aria-label="More options"
                                    >
                                        <MoreVertical size={16} />
                                    </button>
                                </div>
                            </div>
                        ))
                    )}
                </>
            )}
        </div>
    );
};

export default ViewFile;