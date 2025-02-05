import React, { useState } from 'react';
import { Share2, Copy, X } from 'lucide-react';

const ShareFileAction = ({ file, user }) => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [shareLink, setShareLink] = useState('');
    const [password, setPassword] = useState('');
    const [expiresAt, setExpiresAt] = useState('');
    const [maxDownloads, setMaxDownloads] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const [copySuccess, setCopySuccess] = useState(false);

    const isPremium = user?.role === 'premium_user';

    const handleShare = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');
    
        try {
            const token = localStorage.getItem('token');
            if (!token) {
                setError('Please log in to share files.');
                return;
            }
                
            // Format the date if it exists
            const formattedExpiresAt = expiresAt 
                ? new Date(expiresAt).toISOString()
                : null;
    
            console.log('Debug expiry:', {
                originalExpiresAt: expiresAt,
                formattedExpiresAt: formattedExpiresAt
            });
    
            const shareData = {
                file_id: file.id,
                password: password,
                ...(isPremium && {
                    expires_at: formattedExpiresAt,  
                    max_downloads: maxDownloads ? parseInt(maxDownloads) : null
                })
            };
    
            console.log('Sending share data:', shareData);
    
            const endpoint = isPremium
                ? `/api/premium/shares/files/${file.id}`
                : `/api/files/${file.id}/share`;
    
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify(shareData),
            });
    
            if (!response.ok) {
                const errorData = await response.json();
                console.error('Share creation failed:', {
                    status: response.status,
                    error: errorData
                });
                throw new Error(errorData.error || 'Failed to create share link');
            }

            const { data } = await response.json();
            const baseUrl = window.location.origin + (isPremium 
                ? `/premium/share/`
                : `/share/`);
            setShareLink(baseUrl + data.share_link);
        } catch (error) {
            setError(error.message);
        } finally {
            setIsLoading(false);
        }
    };

    const copyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(shareLink);
            setCopySuccess(true);
            setTimeout(() => setCopySuccess(false), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
            setError('Failed to copy to clipboard');
        }
    };

    const closeModal = () => {
        setIsModalOpen(false);
        setShareLink('');
        setPassword('');
        setExpiresAt('');
        setMaxDownloads('');
        setError('');
        setCopySuccess(false);
    };

    return (
        <>
            <button
                onClick={() => setIsModalOpen(true)}
                className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2"
            >
                <Share2 size={16} />
                <span>{isPremium ? 'Advanced Share' : 'Share'}</span>
            </button>

            {isModalOpen && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                    <div className="bg-white rounded-lg p-6 max-w-md w-full">
                        <div className="flex justify-between items-center mb-4">
                            <div>
                                <h2 className="text-xl font-semibold">Share {file.original_name}</h2>
                                {isPremium && (
                                    <span className="text-sm text-blue-600">Premium Share</span>
                                )}
                            </div>
                            <button 
                                onClick={closeModal}
                                className="p-1 hover:bg-gray-100 rounded"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        {error && (
                            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                                {error}
                            </div>
                        )}

                        {!shareLink ? (
                            <form onSubmit={handleShare} className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">
                                        Set Password
                                    </label>
                                    <input
                                        type="password"
                                        placeholder="Enter password (min 6 characters)"
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        required
                                        minLength={6}
                                        className="w-full px-3 py-2 border rounded-md"
                                    />
                                </div>

                                {isPremium && (
                                    <>
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                                Expiration Date (Optional)
                                            </label>
                                            <input
                                                type="datetime-local"
                                                value={expiresAt}
                                                onChange={(e) => setExpiresAt(e.target.value)}
                                                className="w-full px-3 py-2 border rounded-md"
                                            />
                                        </div>

                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                                Maximum Downloads (Optional)
                                            </label>
                                            <input
                                                type="number"
                                                placeholder="Unlimited if not set"
                                                value={maxDownloads}
                                                onChange={(e) => setMaxDownloads(e.target.value)}
                                                min="1"
                                                className="w-full px-3 py-2 border rounded-md"
                                            />
                                        </div>
                                    </>
                                )}

                                <button
                                    type="submit"
                                    disabled={isLoading}
                                    className="w-full bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 disabled:bg-blue-300 disabled:cursor-not-allowed"
                                >
                                    {isLoading ? 'Generating Link...' : 'Generate Share Link'}
                                </button>
                            </form>
                        ) : (
                            <div className="space-y-4">
                                <div className="flex items-center space-x-2">
                                    <input
                                        type="text"
                                        value={shareLink}
                                        readOnly
                                        className="flex-1 px-3 py-2 border rounded-md bg-gray-50"
                                    />
                                    <button
                                        onClick={copyToClipboard}
                                        className={`p-2 rounded-md transition-colors ${
                                            copySuccess 
                                                ? 'bg-green-100 text-green-600' 
                                                : 'bg-gray-100 hover:bg-gray-200'
                                        }`}
                                    >
                                        <Copy size={20} />
                                    </button>
                                </div>
                                {copySuccess && (
                                    <p className="text-sm text-green-600">
                                        Link copied to clipboard!
                                    </p>
                                )}
                                <div className="space-y-2">
                                    <p className="text-sm text-gray-600">Password: {password}</p>
                                    {isPremium && expiresAt && (
                                        <p className="text-sm text-gray-600">
                                            Expires: {new Date(expiresAt).toLocaleString()}
                                        </p>
                                    )}
                                    {isPremium && maxDownloads && (
                                        <p className="text-sm text-gray-600">
                                            Max Downloads: {maxDownloads}
                                        </p>
                                    )}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </>
    );
};

export default ShareFileAction;