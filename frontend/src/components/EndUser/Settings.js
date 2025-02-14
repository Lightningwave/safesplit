import React, { useState, useEffect } from 'react';
import { AlertCircle } from 'lucide-react';
import PasswordReset from './PasswordReset';
import ViewStorage from './ViewStorage';
import TwoFactorSettings from './TwoFactorAuthentication';

const Settings = ({ user: initialUser, onUserUpdate }) => {
    const [activeTab, setActiveTab] = useState('account');
    const [currentUser, setCurrentUser] = useState(initialUser?.data?.user || {});
    const [billingProfile, setBillingProfile] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [showCancelModal, setShowCancelModal] = useState(false);
    const [cancelInfo, setCancelInfo] = useState(null);
    const [isEditing, setIsEditing] = useState(false);
    const [formData, setFormData] = useState({
        billing_name: '',
        billing_email: '',
        billing_address: '',
        country_code: '',
        default_payment_method: 'credit_card',
        billing_cycle: 'monthly',
        currency: 'USD'
    });

    useEffect(() => {
        fetchUserData();
    }, []);

    useEffect(() => {
        if (billingProfile) {
            setFormData({
                billing_name: billingProfile.billing_name || '',
                billing_email: billingProfile.billing_email || '',
                billing_address: billingProfile.billing_address || '',
                country_code: billingProfile.country_code || '',
                default_payment_method: billingProfile.default_payment_method || 'credit_card',
                billing_cycle: billingProfile.billing_cycle || 'monthly',
                currency: billingProfile.currency || 'USD'
            });
        }
    }, [billingProfile]);

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
                if (data.data.billing_profile) {
                    setBillingProfile(data.data.billing_profile);
                }
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

    const handleUpdateBilling = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/premium/billing/details', {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });

            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Failed to update billing details');
            }

            setBillingProfile(data.data);
            setIsEditing(false);
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

    const renderAccountDetails = () => {
        return (
            <div className="space-y-3">
                <p><strong>Username:</strong> {currentUser.username}</p>
                <p><strong>Email:</strong> {currentUser.email}</p>
                <p><strong>Subscription Plan:</strong> {currentUser.subscription_status}</p>
                
                {billingProfile && (
                    <>
                        <p><strong>Billing Cycle:</strong> {billingProfile.billing_cycle}</p>
                        <p>
                            <strong>Next Billing Date:</strong> {' '}
                            {billingProfile.next_billing_date 
                                ? new Date(billingProfile.next_billing_date).toLocaleDateString()
                                : 'N/A'
                            }
                        </p>
                    </>
                )}

                {currentUser.subscription_status === 'premium' && 
                 billingProfile?.billing_status === 'active' && (
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
                            Your subscription is cancelled and will end on{' '}
                            {new Date(billingProfile.next_billing_date).toLocaleDateString()}
                        </p>
                    </div>
                )}
            </div>
        );
    };

    const renderBillingDetails = () => {
        if (!isEditing) {
            return (
                <div className="space-y-4">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-xl font-semibold">Billing Details</h2>
                        <button
                            onClick={() => setIsEditing(true)}
                            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                        >
                            Edit Details
                        </button>
                    </div>
                    
                    {billingProfile ? (
                        <div className="space-y-3">
                            <p><strong>Billing Name:</strong> {billingProfile.billing_name}</p>
                            <p><strong>Billing Email:</strong> {billingProfile.billing_email}</p>
                            <p><strong>Billing Address:</strong> {billingProfile.billing_address}</p>
                            <p><strong>Country:</strong> {billingProfile.country_code}</p>
                            <p><strong>Payment Method:</strong> {billingProfile.default_payment_method}</p>
                            <p><strong>Billing Cycle:</strong> {billingProfile.billing_cycle}</p>
                            <p><strong>Currency:</strong> {billingProfile.currency}</p>
                        </div>
                    ) : (
                        <p className="text-gray-500">No billing profile found. Click Edit to set up billing details.</p>
                    )}
                </div>
            );
        }

        return (
            <form onSubmit={handleUpdateBilling} className="space-y-4">
                <h2 className="text-xl font-semibold mb-4">Edit Billing Details</h2>
                
                <div>
                    <label className="block text-sm font-medium text-gray-700">Billing Name</label>
                    <input
                        type="text"
                        value={formData.billing_name}
                        onChange={(e) => setFormData(prev => ({ ...prev, billing_name: e.target.value }))}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                        required
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">Billing Email</label>
                    <input
                        type="email"
                        value={formData.billing_email}
                        onChange={(e) => setFormData(prev => ({ ...prev, billing_email: e.target.value }))}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                        required
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">Billing Address</label>
                    <textarea
                        value={formData.billing_address}
                        onChange={(e) => setFormData(prev => ({ ...prev, billing_address: e.target.value }))}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                        required
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">Country Code</label>
                    <input
                        type="text"
                        value={formData.country_code}
                        onChange={(e) => setFormData(prev => ({ ...prev, country_code: e.target.value }))}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                        maxLength="2"
                        required
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">Payment Method</label>
                    <select
                        value={formData.default_payment_method}
                        onChange={(e) => setFormData(prev => ({ ...prev, default_payment_method: e.target.value }))}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                    >
                        <option value="credit_card">Credit Card</option>
                        <option value="bank_account">Bank Account</option>
                        <option value="paypal">PayPal</option>
                    </select>
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">Billing Cycle</label>
                    <select
                        value={formData.billing_cycle}
                        onChange={(e) => setFormData(prev => ({ ...prev, billing_cycle: e.target.value }))}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                    >
                        <option value="monthly">Monthly</option>
                        <option value="yearly">Yearly</option>
                    </select>
                </div>

                <div className="flex space-x-4 pt-4">
                    <button
                        type="submit"
                        disabled={loading}
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                    >
                        {loading ? 'Saving...' : 'Save Changes'}
                    </button>
                    <button
                        type="button"
                        onClick={() => setIsEditing(false)}
                        className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-50"
                    >
                        Cancel
                    </button>
                </div>
            </form>
        );
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
                    onClick={() => setActiveTab('billing')}
                    className={`pb-2 border-b-2 ${
                        activeTab === 'billing' ? 'border-blue-500 text-blue-500' : 'border-transparent text-gray-600 hover:text-gray-800'
                    }`}
                >
                    Billing Details
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
                        {renderAccountDetails()}
                    </div>
                )}
    
                {activeTab === 'billing' && renderBillingDetails()}
    
                {activeTab === 'password' && <PasswordReset />}
                
                {activeTab === '2fa' && <TwoFactorSettings />}
    
                {activeTab === 'storage' && <ViewStorage user={currentUser} />}
    
                {showCancelModal && <CancelModal />}
            </div>
        );
    };
    
    export default Settings;