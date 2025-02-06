import React, { useState } from 'react';
import { Share2, Copy, X, Info } from 'lucide-react';

const ShareFileAction = ({ file, user }) => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [shareLink, setShareLink] = useState('');
    const [password, setPassword] = useState('');
    const [expiresAt, setExpiresAt] = useState('');
    const [maxDownloads, setMaxDownloads] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const [copySuccess, setCopySuccess] = useState(false);
    const [shares, setShares] = useState(2);                        // state for no. of shares
    const [threshold, setThreshold] = useState(2);                  // state for threshold
    const [emailInput, setEmailInput] = useState('');               // state for email input
    const [emails, setEmails] = useState([]);                       // state for an array for email
    const [showTooltip, setShowTooltip] = useState(false);          // state for tooltip visibility
    const [shareGenerated, setShareGenerated] = useState(false);    // state for share link generation
    const [downloadMessage, setDownloadMessage] = useState('');     // state for download message

    const isPremium = user?.role === 'premium_user';

    const handleAddEmail = (e) => {
        if (e.key === 'Enter' && e.target.value) {
            const newEmail = e.target.value.trim();             // to remove whitespace
            const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;  // regex for email
    
            // to validate the email format
            if (!emailPattern.test(newEmail)) {
                setError(`Please enter a valid email address: ${newEmail}`);
                return; // to exit if email is invalid
            }
    
            // to check for duplicates
            if (!emails.includes(newEmail)) {
                setEmails([...emails, newEmail]);
                setEmailInput('');      // to clear the input
                setError('');           // to clear previous error
            } else {
                setError(`This email is already in the list: ${newEmail}`);
            }
        }
    };

    const handleDeleteEmail = (emailToDelete) => {
        setEmails(emails.filter(email => email !== emailToDelete));
    };


    const handleShare = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');

        // to validate no. of shares & threshold
        if (shares < threshold) {
            setError('Number of shares must be greater than or equal to threshold');
            setIsLoading(false)
            return;
        }

        if (threshold < 2) {
            setError('Threshold must be at least 2');
            setIsLoading(false)
            return;
        }

        if (shares > 10) {
            setError('Number of shares cannot exceed 10');
            setIsLoading(false)
            return;
        }

        // to validate no. of email (min 1)
        if (emails.length === 0) {
            setError('At least one email must be provided.');
            setIsLoading(false);
            return;
        }

        try {
            const token = localStorage.getItem('token');
            if (!token) {
                setError('Please log in to share files.');
                return;
            }

            const shareData = {
                file_id: file.id,
                password: password,
                shares: shares,         // to include no. of shares
                threshold: threshold,   // to include threshold
                email: emails,          // to include emails
                ...(isPremium && {
                    expires_at: expiresAt || null,
                    max_downloads: maxDownloads ? parseInt(maxDownloads) : null
                })
            };

            const endpoint = isPremium
                ? `http://localhost:8080/api/premium/shares/files/${file.id}`
                : `http://localhost:8080/api/files/${file.id}/share`;

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
                throw new Error(errorData.error || 'Failed to create share link');
            }

            const { data } = await response.json();
            const baseUrl = isPremium 
                ? `http://localhost:3000/premium/share/`
                : `http://localhost:3000/share/`;
            setShareLink(baseUrl + data.share_link);
            setShareGenerated(true);        // to set true, after share successful
            setEmailInput('');              // to clear the email input, after share successful
            console.log(emails);            // to debug
            setDownloadMessage('Split successful!'); // to set download message
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
        setShares(2);               // to reset no. of shares
        setThreshold(2);            // to reset threshold
        setEmailInput('');          // to reset email input
        setEmails([]);              // to reset array of email
        setError('');
        setCopySuccess(false);
        setShareGenerated(false);   // to reset share link generation state
        setDownloadMessage('');     // to reset download message
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

                        <div className="mb-6 space-y-4">
                            <div className="flex items-center justify-between">
                                {/* NO. OF SHARES INPUT SECTION */}
                                <div className="flex items-center">
                                    <label htmlFor="shares" className="text-sm font-medium text-gray-700 mr-2">
                                        Number of Shares
                                    </label>
                                    <div className="relative">
                                        <Info 
                                            size={16} 
                                            className="text-gray-400 cursor-help"
                                            onMouseEnter={() => setShowTooltip('shares')}
                                            onMouseLeave={() => setShowTooltip(false)}
                                        />
                                        {showTooltip === 'shares' && (
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
                                    className={`w-20 px-2 py-1 border rounded-md ${shareGenerated ? 'bg-gray-200 text-gray-500 cursor-not-allowed' : ''} ml-2`}
                                    disabled={shareGenerated}   // to disable, if share link is generated
                                />
                            </div>
                            
                            {/* THRESHOLD INPUT SECTION */}
                            <div className="flex items-center justify-between">
                                <div className="flex items-center">
                                    <label htmlFor="threshold" className="text-sm font-medium text-gray-700 mr-2">
                                        Threshold
                                    </label>
                                    <div className="relative">
                                        <Info 
                                            size={16} 
                                            className="text-gray-400 cursor-help"
                                            onMouseEnter={() => setShowTooltip('threshold')}
                                            onMouseLeave={() => setShowTooltip(false)}
                                        />
                                        {showTooltip === 'threshold' && (
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
                                    className={`w-20 px-2 py-1 border rounded-md ${shareGenerated ? 'bg-gray-200 text-gray-500 cursor-not-allowed' : ''} ml-2`}
                                    disabled={shareGenerated}   // to disable, if share link is generated
                                />
                            </div>
                            
                            {/* EMAIL INPUT SECTION */}
                            <div className="flex items-center justify-between">
                                <div className="flex items-center">
                                    <label htmlFor="email" className="text-sm font-medium text-gray-700 mr-2">
                                        Email
                                    </label>
                                    <div className="relative">
                                            <Info 
                                                size={16} 
                                                className="text-gray-400 cursor-help"
                                                onMouseEnter={() => setShowTooltip('email')}
                                                onMouseLeave={() => setShowTooltip(false)}
                                            />
                                            {showTooltip === 'email' && (
                                                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 p-2 bg-gray-800 text-white text-xs rounded whitespace-nowrap">
                                                    Emails to share with (comma-separated)
                                                </div>
                                            )}
                                    </div>
                                </div>
                                <input
                                    type="email"
                                    id="email"
                                    value={emailInput}
                                    onChange={(e) => setEmailInput(e.target.value)}
                                    onKeyDown={handleAddEmail}
                                    className={`w-full px-2 py-1 border rounded-md ${shareGenerated ? 'bg-gray-200 text-gray-500 cursor-not-allowed' : ''} ml-2`}
                                    placeholder="Enter email(s) and press Enter"
                                    disabled={shareGenerated}   // to disable, if share link is generated
                                    />
                            </div>

                            {/* EMAIL TAGS */}
                            <div className="flex flex-wrap gap-2 mt-2">
                                {emails.map((email, index) => (
                                    <div key={index} className="flex items-center bg-gray-200 rounded-md px-2 py-1">
                                        <span>{email}</span>
                                        <button
                                            onClick={() => handleDeleteEmail(email)}
                                            className={`ml-2 ${shareGenerated ? 'text-gray-500 cursor-not-allowed' : 'text-red-600 hover:text-red-800'}`}
                                            disabled={shareGenerated}   // to disable, if share link is generated
                                        >
                                            <X size={16} />
                                        </button>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* DOWNLOAD MESSAGE */}
                        {downloadMessage && (
                            <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4">
                                {downloadMessage} <strong>Please download the fragment</strong>. This is a one-time action.
                            </div>
                        )}
                        
                        {/* SHARE LINK SECTION */}
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

                                {/* PREMIUM USER SECTION - additional features */}
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
                                    {isLoading ? 'Splitting...' : 'Split'}
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