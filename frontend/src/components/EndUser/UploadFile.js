import React, { useState, useEffect, useCallback } from 'react';
import { Upload, X, Info, FolderIcon, Lock } from 'lucide-react';

const UploadFile = ({ isOpen, onClose, onUpload, currentFolder }) => {
    const [selectedFile, setSelectedFile] = useState(null);
    const [uploading, setUploading] = useState(false);
    const [error, setError] = useState('');
    const [shares, setShares] = useState(3);
    const [threshold, setThreshold] = useState(2);
    const [showTooltip, setShowTooltip] = useState(false);
    const [encryptionTypes, setEncryptionTypes] = useState([]);
    const [selectedEncryption, setSelectedEncryption] = useState('standard');
    const [isPremium, setIsPremium] = useState(false);

    useEffect(() => {
        if (isOpen) {
            setSelectedFile(null);
            setError('');
            setShares(3);
            setThreshold(2);
            fetchEncryptionOptions();
        }
    }, [isOpen]);

    const fetchEncryptionOptions = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/files/encryption/options', {
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
        const file = event.target.files[0];
        setSelectedFile(file);
        setError('');
    }, []);

    const handleDrop = useCallback((event) => {
        event.preventDefault();
        const file = event.dataTransfer.files[0]; // Take only the first file
        setSelectedFile(file);
        setError('');
    }, []);

    const handleDragOver = useCallback((event) => {
        event.preventDefault();
    }, []);

    const handleUpload = async () => {
        if (!selectedFile) {
            setError('Please select a file to upload');
            return;
        }

        if (shares < threshold) {
            setError('Number of shares must be greater than or equal to threshold');
            return;
        }

        if (threshold < 2) {
            setError('Threshold must be at least 2');
            return;
        }

        if (shares > 10) {
            setError('Number of shares cannot exceed 10');
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
            formData.append('file', selectedFile);
            formData.append('shares', shares);
            formData.append('threshold', threshold);
            formData.append('encryption_type', selectedEncryption);

            if (currentFolder?.id) {
                formData.append('folder_id', currentFolder.id);
            }

            console.log('Starting file upload:', {
                fileName: selectedFile.name,
                fileSize: selectedFile.size,
                fileType: selectedFile.type,
                shares,
                threshold,
                encryptionType: selectedEncryption,
                folderId: currentFolder?.id || 'root'
            });

            const response = await fetch('http://localhost:8080/api/files/upload', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
                body: formData,
                credentials: 'include',
            });

            console.log('Upload response status:', response.status);

            const responseText = await response.text();
            console.log('Raw response:', responseText);

            if (!response.ok) {
                try {
                    const errorData = JSON.parse(responseText);
                    throw new Error(errorData.error || 'Upload failed');
                } catch (parseError) {
                    throw new Error(responseText || 'Upload failed');
                }
            }

            let result;
            try {
                result = JSON.parse(responseText);
            } catch (parseError) {
                console.error('Failed to parse response:', responseText);
                throw new Error('Invalid response format from server');
            }

            onUpload(result.data.file);
            onClose();
            
        } catch (err) {
            console.error('Upload error details:', {
                message: err.message,
                stack: err.stack
            });
            
            if (err.message.includes('Unauthorized') || err.message.includes('invalid token')) {
                localStorage.removeItem('token');
                setError('Please log in again to upload files');
            } else {
                setError(err.message || 'Failed to upload file. Please try again.');
            }
        } finally {
            setUploading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <div className="flex justify-between items-center mb-4">
                    <div>
                        <h2 className="text-xl font-semibold">Upload File</h2>
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

                <div className="mb-6">
                    <div 
                        className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center"
                        onDrop={handleDrop}
                        onDragOver={handleDragOver}
                    >
                        <input
                            type="file"
                            onChange={handleFileSelect}
                            className="hidden"
                            id="fileInput"
                        />
                        <label
                            htmlFor="fileInput"
                            className="cursor-pointer flex flex-col items-center"
                        >
                            <Upload size={48} className="text-gray-400 mb-4" />
                            <p className="text-gray-600 mb-2">
                                {selectedFile ? selectedFile.name : 'Click to select a file'}
                            </p>
                            <p className="text-sm text-gray-500">
                                {selectedFile
                                    ? `Size: ${(selectedFile.size / 1024 / 1024).toFixed(2)} MB`
                                    : 'or drag and drop here'}
                            </p>
                        </label>
                    </div>
                </div>

                {/* Encryption Options */}
                <div className="mb-6">
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
                    {encryptionTypes.map((type) => 
                        type.type === selectedEncryption && (
                            <p key={type.type} className="mt-1 text-sm text-gray-500">
                                {type.description}
                            </p>
                        )
                    )}
                </div>

                <div className="mb-6 space-y-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center">
                            <label htmlFor="shares" className="text-sm font-medium text-gray-700 mr-2">
                                Number of Shares
                            </label>
                            <div className="relative">
                                <Info 
                                    size={16} 
                                    className="text-gray-400 cursor-help"
                                    onMouseEnter={() => setShowTooltip(true)}
                                    onMouseLeave={() => setShowTooltip(false)}
                                />
                                {showTooltip && (
                                    <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 p-2 bg-gray-800 text-white text-xs rounded whitespace-nowrap">
                                        Total number of key shares to create
                                    </div>
                                )}
                            </div>
                        </div>
                        <input
                            type="number"
                            id="shares"
                            min="2"
                            max="10"
                            value={shares}
                            onChange={(e) => setShares(parseInt(e.target.value))}
                            className="w-20 px-2 py-1 border rounded-md"
                        />
                    </div>

                    <div className="flex items-center justify-between">
                        <div className="flex items-center">
                            <label htmlFor="threshold" className="text-sm font-medium text-gray-700 mr-2">
                                Threshold
                            </label>
                            <div className="relative">
                                <Info 
                                    size={16} 
                                    className="text-gray-400 cursor-help"
                                    onMouseEnter={() => setShowTooltip(true)}
                                    onMouseLeave={() => setShowTooltip(false)}
                                />
                                {showTooltip && (
                                    <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 p-2 bg-gray-800 text-white text-xs rounded whitespace-nowrap">
                                        Minimum shares required to decrypt
                                    </div>
                                )}
                            </div>
                        </div>
                        <input
                            type="number"
                            id="threshold"
                            min="2"
                            max={shares}
                            value={threshold}
                            onChange={(e) => setThreshold(parseInt(e.target.value))}
                            className="w-20 px-2 py-1 border rounded-md"
                        />
                    </div>
                </div>

                {error && (
                    <div className="mb-4 p-3 bg-red-50 text-red-600 rounded-md text-sm text-center">
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
                        disabled={uploading || !selectedFile}
                        className={`px-4 py-2 bg-gray-600 text-white rounded-md 
                            ${uploading || !selectedFile ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-700'}`}
                    >
                        {uploading ? 'Uploading...' : 'Upload'}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default UploadFile;