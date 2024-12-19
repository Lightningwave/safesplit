import React, { useState, useEffect } from 'react';
import { HardDrive, AlertCircle } from 'lucide-react';

const ViewStorage = () => {
    const [storageInfo, setStorageInfo] = useState(null);
    const [error, setError] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        fetchStorageInfo();
    }, []);

    const fetchStorageInfo = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/storage/info', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch storage information');
            }

            const data = await response.json();
            if (data.status === 'success') {
                setStorageInfo(data.data);
            } else {
                throw new Error(data.error || 'Failed to load storage information');
            }
        } catch (err) {
            setError(err.message);
        } finally {
            setIsLoading(false);
        }
    };

    const formatStorageSize = (bytes) => {
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        let size = bytes;
        let unitIndex = 0;

        while (size >= 1024 && unitIndex < units.length - 1) {
            size /= 1024;
            unitIndex++;
        }

        return `${size.toFixed(2)} ${units[unitIndex]}`;
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center p-6">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex items-center space-x-2 p-4 bg-red-50 text-red-700 rounded-md">
                <AlertCircle size={20} />
                <span>{error}</span>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex items-start justify-between">
                <div>
                    <h2 className="text-xl font-semibold mb-2">Storage Usage</h2>
                    <p className="text-gray-600">
                        {storageInfo?.is_premium ? 'Premium Storage' : 'Free Storage'}
                    </p>
                </div>
                <HardDrive size={24} className="text-gray-400" />
            </div>

            <div className="space-y-4">
                <div>
                    <div className="flex justify-between text-sm mb-1">
                        <span>{formatStorageSize(storageInfo?.storage_used || 0)} used</span>
                        <span>{formatStorageSize(storageInfo?.storage_quota || 0)} total</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2.5">
                        <div
                            className={`h-2.5 rounded-full ${
                                (storageInfo?.percentage_used || 0) > 90 
                                    ? 'bg-red-600' 
                                    : (storageInfo?.percentage_used || 0) > 70 
                                        ? 'bg-yellow-600' 
                                        : 'bg-blue-600'
                            }`}
                            style={{ width: `${Math.min(storageInfo?.percentage_used || 0, 100)}%` }}
                        ></div>
                    </div>
                </div>

                <div className="bg-gray-50 p-4 rounded-lg space-y-2">
                    <div className="flex justify-between">
                        <span className="text-gray-600">Total Files</span>
                        <span className="font-medium">{storageInfo?.total_files || 0}</span>
                    </div>
                    <div className="flex justify-between">
                        <span className="text-gray-600">Storage Used</span>
                        <span className="font-medium">{formatStorageSize(storageInfo?.storage_used || 0)}</span>
                    </div>
                    <div className="flex justify-between">
                        <span className="text-gray-600">Storage Available</span>
                        <span className="font-medium">
                            {formatStorageSize((storageInfo?.storage_quota || 0) - (storageInfo?.storage_used || 0))}
                        </span>
                    </div>
                </div>

                {!storageInfo?.is_premium && (
                    <div className="bg-blue-50 p-4 rounded-lg">
                        <p className="text-blue-700 text-sm">
                            Upgrade to Premium for 50GB storage and additional features!
                        </p>
                        <button className="mt-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors w-full">
                            Upgrade to Premium
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default ViewStorage;