import React, { useState, useEffect } from 'react';
import { Download, Lock, KeyRound } from 'lucide-react';
import { useLocation, useParams } from 'react-router-dom';

const SharedFileAccess = () => {
    const [password, setPassword] = useState('');
    const [verificationCode, setVerificationCode] = useState('');
    const [showVerification, setShowVerification] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const [fileInfo, setFileInfo] = useState(null);
    const [showPassword, setShowPassword] = useState(false);

    const location = useLocation();
    const { shareId } = useParams();
    const isPremiumShare = location.pathname.includes('/premium/share/');
    const isProtectedShare = location.pathname.includes('/protected-share/');

    useEffect(() => {
        fetchFileInfo();
    }, []);

    const fetchFileInfo = async () => {
        try {
            const endpoint = `/api/${isPremiumShare ? 'premium/shares' : 'files/share'}/${shareId}`;
            const response = await fetch(endpoint);
            if (response.ok) {
                const data = await response.json();
                setFileInfo(data.data);
            }
        } catch (error) {
            console.error('Error fetching file info:', error);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');
    
        try {
            const endpoint = `/api/${isPremiumShare ? 'premium/shares' : 'files/share'}/${shareId}`;
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ password }),
            });
    
            const contentType = response.headers.get('Content-Type');
    
            // If response is JSON
            if (contentType && contentType.includes('application/json')) {
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.error || 'Failed to access file');
                }
    
                if (data.message?.includes('2FA code sent')) {
                    setShowVerification(true);
                    return;
                }
            } 
            // If response is a file
            else if (response.ok) {
                await handleDownload(response);
            } 
            // If error response
            else {
                throw new Error('Failed to access file');
            }
    
        } catch (error) {
            setError(error.message);
        } finally {
            setIsLoading(false);
        }
    };
    
    const handleVerification = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');
    
        try {
            const endpoint = `/api/${isPremiumShare ? 'premium/shares' : 'files/share'}/${shareId}/verify`;
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    code: verificationCode,
                    password
                }),
            });
    
            const contentType = response.headers.get('Content-Type');
    
            // If response is JSON (error case)
            if (contentType && contentType.includes('application/json')) {
                const data = await response.json();
                if (!response.ok) {
                    throw new Error(data.error || 'Verification failed');
                }
            }
            // If response is a file
            else if (response.ok) {
                await handleDownload(response);
            }
            // If error response
            else {
                throw new Error('Verification failed');
            }
    
        } catch (error) {
            setError(error.message);
        } finally {
            setIsLoading(false);
        }
    };
    
    const handleDownload = async (response) => {
        const blob = await response.blob();
        const disposition = response.headers.get('Content-Disposition');
        let filename = 'download';
        
        if (disposition) {
            const utf8FilenameMatch = disposition.match(/filename\*=UTF-8''([^;]+)/i);
            if (utf8FilenameMatch) {
                filename = decodeURIComponent(utf8FilenameMatch[1]);
            } else {
                const asciiFilenameMatch = disposition.match(/filename="([^"]+)"/i);
                if (asciiFilenameMatch) {
                    filename = asciiFilenameMatch[1];
                }
            }
        }
    
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
    };

    return (
        <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8 bg-white p-8 rounded-xl shadow-lg">
                <div className="text-center">
                    <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100 mb-4">
                        {showVerification ? (
                            <KeyRound className="h-6 w-6 text-blue-600" />
                        ) : (
                            <Lock className="h-6 w-6 text-blue-600" />
                        )}
                    </div>
                    <h2 className="text-3xl font-bold text-gray-900">
                        {showVerification ? 'Verify Access' : 'Access Shared File'}
                    </h2>
                    {fileInfo && (
                        <div className="mt-2 text-sm text-gray-600">
                            <p className="font-medium">{fileInfo.file_name}</p>
                            <p>Size: {(fileInfo.file_size / 1024 / 1024).toFixed(2)} MB</p>
                            {isPremiumShare && fileInfo.expires_at && (
                                <p>Expires: {new Date(fileInfo.expires_at).toLocaleString()}</p>
                            )}
                            {isPremiumShare && fileInfo.max_downloads && (
                                <p>Downloads: {fileInfo.download_count} / {fileInfo.max_downloads}</p>
                            )}
                        </div>
                    )}
                </div>

                {error && (
                    <div className="bg-red-50 border-l-4 border-red-400 p-4 rounded">
                        <div className="flex">
                            <div className="flex-shrink-0">
                                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                                </svg>
                            </div>
                            <div className="ml-3">
                                <p className="text-sm text-red-700">{error}</p>
                            </div>
                        </div>
                    </div>
                )}

                {!showVerification ? (
                    <form onSubmit={handleSubmit} className="mt-8 space-y-6">
                        <div className="rounded-md shadow-sm space-y-4">
                            <div>
                                <label htmlFor="password" className="sr-only">Password</label>
                                <div className="relative">
                                    <input
                                        id="password"
                                        name="password"
                                        type={showPassword ? "text" : "password"}
                                        required
                                        className="appearance-none rounded-md relative block w-full px-3 py-3 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        placeholder="Enter share password"
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                    />
                                    <button
                                        type="button"
                                        className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                                        onClick={() => setShowPassword(!showPassword)}
                                    >
                                        {showPassword ? (
                                            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                            </svg>
                                        ) : (
                                            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                                            </svg>
                                        )}
                                    </button>
                                </div>
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={isLoading}
                            className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-blue-300 disabled:cursor-not-allowed transition-colors duration-200"
                        >
                            {isLoading ? (
                                <div className="flex items-center">
                                    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                    </svg>
                                    Processing...
                                </div>
                            ) : (
                                <div className="flex items-center">
                                    <Download className="mr-2 h-5 w-5" />
                                    Access File
                                </div>
                            )}
                        </button>
                    </form>
                ) : (
                    <form onSubmit={handleVerification} className="mt-8 space-y-6">
                        <div>
                            <label htmlFor="code" className="sr-only">Verification Code</label>
                            <input
                                id="code"
                                name="code"
                                type="text"
                                required
                                className="appearance-none rounded-md relative block w-full px-3 py-3 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                placeholder="Enter verification code"
                                value={verificationCode}
                                onChange={(e) => setVerificationCode(e.target.value)}
                            />
                        </div>

                        <button
                            type="submit"
                            disabled={isLoading}
                            className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-blue-300 disabled:cursor-not-allowed"
                        >
                            {isLoading ? (
                                <div className="flex items-center">
                                    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                    </svg>
                                    Verifying...
                                </div>
                            ) : (
                                'Verify and Download'
                            )}
                        </button>
                    </form>
                )}
            </div>
        </div>
    );
};

export default SharedFileAccess;