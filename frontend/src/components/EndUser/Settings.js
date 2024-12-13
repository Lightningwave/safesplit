import React, { useState } from 'react';

const Settings = ({ user }) => {
    const [activeTab, setActiveTab] = useState('account');

    // Dummy storage data - in a real scenario, this could be fetched from an API
    const usedStorage = '2GB';
    const totalStorage = '10GB';

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

            {activeTab === 'account' && (
                <div>
                    <h2 className="text-xl font-semibold mb-4">Account Details</h2>
                    <p><strong>Name:</strong> {user.name}</p>
                    <p><strong>Email:</strong> {user.email}</p>
                    <p><strong>Joined:</strong> {user.joinedDate}</p>
                    {/* Add more user account detail fields as needed */}
                </div>
            )}

            {activeTab === 'password' && (
                <div>
                    <h2 className="text-xl font-semibold mb-4">Change Password</h2>
                    <form className="space-y-4 max-w-sm">
                        <div>
                            <label className="block mb-1 font-medium">Current Password</label>
                            <input type="password" className="w-full border rounded px-3 py-2" />
                        </div>
                        <div>
                            <label className="block mb-1 font-medium">New Password</label>
                            <input type="password" className="w-full border rounded px-3 py-2" />
                        </div>
                        <div>
                            <label className="block mb-1 font-medium">Confirm New Password</label>
                            <input type="password" className="w-full border rounded px-3 py-2" />
                        </div>
                        <button 
                            type="submit"
                            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                        >
                            Update Password
                        </button>
                    </form>
                </div>
            )}

            {activeTab === '2fa' && (
                <div>
                    <h2 className="text-xl font-semibold mb-4">Setup Two-Factor Authentication</h2>
                    {/* Instructions or QR code for 2FA setup */}
                    <p className="mb-4">Enable Two-Factor Authentication for increased account security.</p>
                    <button className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700">
                        Enable 2FA
                    </button>
                </div>
            )}

            {activeTab === 'storage' && (
                <div>
                    <h2 className="text-xl font-semibold mb-4">Storage Usage</h2>
                    <p>You have used {usedStorage} of {totalStorage}.</p>
                    <div className="w-full bg-gray-200 rounded h-4 mt-2 mb-4 relative">
                        <div
                            className="bg-blue-600 h-4 rounded"
                            style={{ width: `${(parseFloat(usedStorage) / parseFloat(totalStorage)) * 100}%` }}
                        ></div>
                    </div>
                    <p>Consider upgrading your plan for more storage.</p>
                    <button className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                        Upgrade Plan
                    </button>
                </div>
            )}
        </div>
    );
};

export default Settings;
