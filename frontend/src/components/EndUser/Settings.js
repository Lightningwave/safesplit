import React, { useState, useEffect } from 'react';
import { AlertCircle } from 'lucide-react';
import PasswordReset from './PasswordReset';
import ViewStorage from './ViewStorage';
import TwoFactorSettings from './TwoFactorAuthentication';

const Settings = ({ user: initialUser, onUserUpdate }) => {
    const [activeTab, setActiveTab] = useState('account');
    const [currentUser, setCurrentUser] = useState(initialUser?.data?.user || {});
    const [billingProfile, setBillingProfile] = useState(initialUser?.data?.billing_profile || {});
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [showCancelModal, setShowCancelModal] = useState(false);
    const [cancelInfo, setCancelInfo] = useState(null);

    useEffect(() => {
        fetchUserData();
    }, []);

    const fetchUserData = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/me', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                setCurrentUser(data.data.user);
                setBillingProfile(data.data.billing_profile);
                if (onUserUpdate) onUserUpdate(data);
            }
        } catch (error) {
            console.error('Failed to fetch user data:', error);
        }
    };

    const handleCancelSubscription = async () => {
        setLoading(true);
        setError('');
        
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/payment/cancel', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Failed to cancel subscription');
            }

            setCancelInfo(data);
            setShowCancelModal(true);
            await fetchUserData();
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const CancelModal = () => (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white p-6 rounded-lg max-w-md w-full m-4">
                <h3 className="text-xl font-semibold mb-4">Subscription Cancelled</h3>
                <p className="mb-4">Your subscription has been cancelled successfully.</p>
                <p className="mb-2">Your premium features will remain active for:</p>
                <p className="text-xl font-bold mb-4">{cancelInfo?.remaining_days} days</p>
                <p className="text-sm text-gray-600 mb-4">
                    Until: {cancelInfo?.end_date ? new Date(cancelInfo.end_date).toLocaleDateString() : 'N/A'}
                </p>
                <div className="mt-4 bg-yellow-50 border border-yellow-200 rounded p-4 mb-4">
                    <p className="text-yellow-800">{cancelInfo?.downgrade_info}</p>
                </div>
                <button
                    onClick={() => setShowCancelModal(false)}
                    className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                    Close
                </button>
            </div>
        </div>
    );

    const formatBytes = (bytes) => {
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        if (bytes === 0) return '0 Bytes';
        const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
        return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i];
    };

    return (
        <div className="bg-white p-6 rounded shadow-md">
            <div className="flex space-x-4 mb-6 border-b pb-2">
                <button
                    onClick={() => setActiveTab('account')}
                    className={`pb-2 border-b-2 ${
                        activeTab === 'account' ? 'border-blue-500 text-blue-500' : 'border-transparent text-gray-600 hover:text-gray-800'
                    }`}
                >
                    Account Details
                </button>
                <button
                    onClick={() => setActiveTab('password')}
                    className={`pb-2 border-b-2 ${
                        activeTab === 'password' ? 'border-blue-500 text-blue-500' : 'border-transparent text-gray-600 hover:text-gray-800'
                    }`}
                >
                    Change Password
                </button>
                <button
                    onClick={() => setActiveTab('2fa')}
                    className={`pb-2 border-b-2 ${
                        activeTab === '2fa' ? 'border-blue-500 text-blue-500' : 'border-transparent text-gray-600 hover:text-gray-800'
                    }`}
                >
                    Setup 2FA
                </button>
                <button
                    onClick={() => setActiveTab('storage')}
                    className={`pb-2 border-b-2 ${
                        activeTab === 'storage' ? 'border-blue-500 text-blue-500' : 'border-transparent text-gray-600 hover:text-gray-800'
                    }`}
                >
                    Storage
                </button>
            </div>

            {error && (
                <div className="flex items-center space-x-2 p-4 mb-4 bg-red-50 text-red-700 rounded-md">
                    <AlertCircle size={20} />
                    <span>{error}</span>
                </div>
            )}

            {activeTab === 'account' && (
                <div>
                    <h2 className="text-xl font-semibold mb-4">Account Details</h2>
                    <div className="space-y-3">
                        <p><strong>Username:</strong> {currentUser.username}</p>
                        <p><strong>Email:</strong> {currentUser.email}</p>
                        <p><strong>Subscription Plan:</strong> {currentUser.subscription_status}</p>
                        {billingProfile && (
                            <>
                                <p><strong>Billing Cycle:</strong> {billingProfile.billing_cycle}</p>
                                <p><strong>Next Billing Date:</strong> {new Date(billingProfile.next_billing_date).toLocaleDateString()}</p>
                            </>
                        )}
                        {currentUser.subscription_status === 'premium' && billingProfile?.billing_status === 'active' && (
                            <div className="mt-6">
                                <button
                                    onClick={handleCancelSubscription}
                                    disabled={loading}
                                    className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50 transition-colors"
                                >
                                    {loading ? 'Processing...' : 'Cancel Subscription'}
                                </button>
                            </div>
                        )}
                        {billingProfile?.billing_status === 'cancelled' && (
                            <div className="mt-6 p-4 bg-gray-50 rounded-md">
                                <p className="text-gray-700">
                                    Your subscription is cancelled and will end on {new Date(billingProfile.next_billing_date).toLocaleDateString()}
                                </p>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {activeTab === 'password' && <PasswordReset />}
            
            {activeTab === '2fa' && <TwoFactorSettings />}

            {activeTab === 'storage' && <ViewStorage user={currentUser} />}

            {showCancelModal && <CancelModal />}
        </div>
    );
};

export default Settings;