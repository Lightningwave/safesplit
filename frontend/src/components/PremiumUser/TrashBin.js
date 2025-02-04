import React, { useState, useEffect } from 'react';
import { RefreshCw, AlertCircle } from 'lucide-react';

const TrashBin = ({ user }) => {
    const [deletedFiles, setDeletedFiles] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState(null);
    const [recoveryStatus, setRecoveryStatus] = useState({});

    useEffect(() => {
        fetchDeletedFiles();
    }, []);

    const fetchDeletedFiles = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/api/premium/recovery/files', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch deleted files');
            }

            const result = await response.json();
            if (result.status === 'success') {
                setDeletedFiles(result.data || []);
            } else {
                throw new Error(result.error || 'Failed to load deleted files');
            }
        } catch (error) {
            setError('Failed to load deleted files. Please try again.');
            console.error('Error:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleRecoverFile = async (fileId) => {
        setRecoveryStatus(prev => ({ ...prev, [fileId]: 'recovering' }));
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`/api/premium/recovery/files/${fileId}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to recover file');
            }

            const result = await response.json();
            if (result.status === 'success') {
                setRecoveryStatus(prev => ({ ...prev, [fileId]: 'success' }));
                // Remove the file from the list after successful recovery
                setDeletedFiles(prev => prev.filter(file => file.id !== fileId));
            } else {
                throw new Error(result.error || 'Failed to recover file');
            }
        } catch (error) {
            setRecoveryStatus(prev => ({ ...prev, [fileId]: 'error' }));
            setError('Failed to recover file. Please try again.');
            console.error('Error:', error);
        }
    };

    const formatFileSize = (bytes) => {
        if (bytes < 1024) return bytes + ' B';
        else if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
        else if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB';
        else return (bytes / 1073741824).toFixed(1) + ' GB';
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleString();
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[200px]">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center mb-6">
                <div className="flex items-center space-x-2">
                    <h2 className="text-xl font-semibold">Deleted Files</h2>
                    <span className="text-sm text-gray-500">
                        ({deletedFiles.length} {deletedFiles.length === 1 ? 'item' : 'items'})
                    </span>
                </div>
                <button
                    onClick={fetchDeletedFiles}
                    className="p-2 hover:bg-gray-100 rounded-full transition-colors"
                    title="Refresh"
                >
                    <RefreshCw size={20} className="text-gray-500" />
                </button>
            </div>

            {error && (
                <div className="flex items-center space-x-2 p-4 bg-red-50 text-red-700 rounded-md">
                    <AlertCircle size={20} />
                    <span>{error}</span>
                </div>
            )}

            {deletedFiles.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                    <p>No deleted files found</p>
                </div>
            ) : (
                <div className="bg-white rounded-lg shadow overflow-hidden">
                    <table className="min-w-full divide-y divide-gray-200">
                        <thead className="bg-gray-50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Size</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Deleted At</th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                            {deletedFiles.map((file) => (
                                <tr key={file.id} className="hover:bg-gray-50">
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="text-sm font-medium text-gray-900">{file.name}</div>
                                        <div className="text-sm text-gray-500">{file.mime_type}</div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                        {formatFileSize(file.size)}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                        {formatDate(file.deleted_at)}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        <button
                                            onClick={() => handleRecoverFile(file.id)}
                                            disabled={recoveryStatus[file.id] === 'recovering'}
                                            className={`px-4 py-2 rounded-md text-white transition-colors ${
                                                recoveryStatus[file.id] === 'recovering'
                                                    ? 'bg-gray-400'
                                                    : 'bg-gray-600 hover:bg-gray-700'
                                            }`}
                                        >
                                            {recoveryStatus[file.id] === 'recovering' ? (
                                                <span className="flex items-center space-x-1">
                                                    <RefreshCw size={16} className="animate-spin" />
                                                    <span>Recovering...</span>
                                                </span>
                                            ) : (
                                                'Recover'
                                            )}
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};

export default TrashBin;