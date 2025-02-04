import React, { useState, useCallback, useEffect } from 'react';
import { Upload, X, Info, FolderIcon, Lock, FileIcon, CheckCircle, AlertCircle } from 'lucide-react';

const MassUploadFile = ({ isOpen, onClose, onUpload, currentFolder }) => {
    const [selectedFiles, setSelectedFiles] = useState([]);
    const [uploading, setUploading] = useState(false);
    const [error, setError] = useState('');
    const [shares, setShares] = useState(3);
    const [threshold, setThreshold] = useState(2);
    const [showTooltip, setShowTooltip] = useState(false);
    const [encryptionTypes, setEncryptionTypes] = useState([]);
    const [selectedEncryption, setSelectedEncryption] = useState('standard');
    const [isPremium, setIsPremium] = useState(false);
    const [uploadProgress, setUploadProgress] = useState({});

    useEffect(() => {
        if (isOpen) {
            setSelectedFiles([]);
            setError('');
            setShares(3);
            setThreshold(2);
            setUploadProgress({});
            fetchEncryptionOptions();
        }
    }, [isOpen]);

    const fetchEncryptionOptions = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/api/files/encryption/options', {
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
                credentials: 'include',
            });

            if (!response.ok) throw new Error('Failed to fetch encryption options');

            const data = await response.json();
            setEncryptionTypes(data.data.available_encryption);
            setIsPremium(data.data.is_premium);
            setSelectedEncryption(data.data.default);
        } catch (err) {
            console.error('Failed to fetch encryption options:', err);
            setError('Failed to load encryption options');
        }
    };

    const handleFileSelect = useCallback((event) => {
        const files = Array.from(event.target.files);
        setSelectedFiles(prev => [...prev, ...files]);
        setError('');
    }, []);

    const handleDrop = useCallback((event) => {
        event.preventDefault();
        const files = Array.from(event.dataTransfer.files);
        setSelectedFiles(prev => [...prev, ...files]);
        setError('');
    }, []);

    const removeFile = (index) => {
        setSelectedFiles(prev => prev.filter((_, i) => i !== index));
    };

    const handleUpload = async () => {
        if (selectedFiles.length === 0) {
            setError('Please select files to upload');
            return;
        }

        if (shares < threshold) {
            setError('Number of shares must be greater than or equal to threshold');
            return;
        }

        setUploading(true);
        setError('');

        try {
            const token = localStorage.getItem('token');
            if (!token) {
                setError('Please log in to upload files');
                return;
            }

            const formData = new FormData();
            selectedFiles.forEach(file => {
                formData.append('files', file);
            });
            formData.append('shares', shares);
            formData.append('threshold', threshold);
            formData.append('encryption_type', selectedEncryption);

            if (currentFolder?.id) {
                formData.append('folder_id', currentFolder.id);
            }

            const response = await fetch('/api/files/mass-upload', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
                body: formData,
                credentials: 'include',
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.error || 'Upload failed');
            }

            onUpload(result.results);
            onClose();
            
        } catch (err) {
            console.error('Upload error:', err);
            setError(err.message || 'Failed to upload files');
        } finally {
            setUploading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl">
                <div className="flex justify-between items-center mb-4">
                    <div>
                        <h2 className="text-xl font-semibold">Mass Upload Files</h2>
                        {currentFolder && (
                            <p className="text-sm text-gray-500 mt-1 flex items-center">
                                <FolderIcon size={16} className="mr-1" />
                                To: {currentFolder.name}
                            </p>
                        )}
                    </div>
                    <button 
                        onClick={onClose}
                        className="text-gray-500 hover:text-gray-700"
                    >
                        <X size={24} />
                    </button>
                </div>

                {/* Drop Zone */}
                <div 
                    className="mb-6 border-2 border-dashed border-gray-300 rounded-lg p-8 text-center"
                    onDrop={handleDrop}
                    onDragOver={(e) => e.preventDefault()}
                >
                    <input
                        type="file"
                        onChange={handleFileSelect}
                        className="hidden"
                        id="fileInput"
                        multiple
                    />
                    <label
                        htmlFor="fileInput"
                        className="cursor-pointer flex flex-col items-center"
                    >
                        <Upload size={48} className="text-gray-400 mb-4" />
                        <p className="text-gray-600 mb-2">
                            Click to select files or drag and drop
                        </p>
                        <p className="text-sm text-gray-500">
                            You can select multiple files
                        </p>
                    </label>
                </div>

                {/* Selected Files List */}
                {selectedFiles.length > 0 && (
                    <div className="mb-6 max-h-40 overflow-y-auto">
                        <h3 className="text-sm font-medium text-gray-700 mb-2">
                            Selected Files ({selectedFiles.length})
                        </h3>
                        <div className="space-y-2">
                            {selectedFiles.map((file, index) => (
                                <div 
                                    key={index}
                                    className="flex items-center justify-between p-2 bg-gray-50 rounded"
                                >
                                    <div className="flex items-center">
                                        <FileIcon size={16} className="mr-2 text-gray-500" />
                                        <span className="text-sm text-gray-600">{file.name}</span>
                                        <span className="ml-2 text-xs text-gray-400">
                                            ({(file.size / 1024 / 1024).toFixed(2)} MB)
                                        </span>
                                    </div>
                                    <button
                                        onClick={() => removeFile(index)}
                                        className="text-red-500 hover:text-red-700"
                                    >
                                        <X size={16} />
                                    </button>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* Encryption and Share Settings */}
                <div className="grid grid-cols-2 gap-4 mb-6">
                    <div>
                        <div className="flex items-center mb-2">
                            <Lock size={16} className="mr-2 text-gray-600" />
                            <label className="text-sm font-medium text-gray-700">
                                Encryption Method
                            </label>
                        </div>
                        <select
                            value={selectedEncryption}
                            onChange={(e) => setSelectedEncryption(e.target.value)}
                            className="w-full p-2 border rounded-md bg-white"
                        >
                            {encryptionTypes.map((type) => (
                                <option 
                                    key={type.type} 
                                    value={type.type}
                                    disabled={!isPremium && type.type !== 'standard'}
                                >
                                    {type.name} {!isPremium && type.type !== 'standard' && '(Premium)'}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="space-y-4">
                        <div>
                            <label className="text-sm font-medium text-gray-700 mb-1 block">
                                Number of Shares
                            </label>
                            <input
                                type="number"
                                min="2"
                                max="10"
                                value={shares}
                                onChange={(e) => setShares(parseInt(e.target.value))}
                                className="w-full p-2 border rounded-md"
                            />
                        </div>
                        <div>
                            <label className="text-sm font-medium text-gray-700 mb-1 block">
                                Threshold
                            </label>
                            <input
                                type="number"
                                min="2"
                                max={shares}
                                value={threshold}
                                onChange={(e) => setThreshold(parseInt(e.target.value))}
                                className="w-full p-2 border rounded-md"
                            />
                        </div>
                    </div>
                </div>

                {error && (
                    <div className="mb-4 p-3 bg-red-50 text-red-600 rounded-md text-sm">
                        {error}
                    </div>
                )}

                <div className="flex justify-end space-x-4">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-gray-600 hover:text-gray-800"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleUpload}
                        disabled={uploading || selectedFiles.length === 0}
                        className={`px-4 py-2 bg-gray-600 text-white rounded-md 
                            ${uploading || selectedFiles.length === 0 ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-700'}`}
                    >
                        {uploading ? 'Uploading...' : `Upload ${selectedFiles.length} Files`}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default MassUploadFile;