import React, { useState } from 'react';
import PasswordReset from './PasswordReset';
import ViewStorage from './ViewStorage';  

const Settings = ({ user }) => {
    const [activeTab, setActiveTab] = useState('account');

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
                    <p><strong>Username:</strong> {user.username}</p>
                    <p><strong>Email:</strong> {user.email}</p>
                </div>
            )}

            {activeTab === 'password' && <PasswordReset />}

            {activeTab === '2fa' && (
                <div>
                    <h2 className="text-xl font-semibold mb-4">Setup Two-Factor Authentication</h2>
                    <p className="mb-4">Enable Two-Factor Authentication for increased account security.</p>
                    <button className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700">
                        Enable 2FA
                    </button>
                </div>
            )}

            {activeTab === 'storage' && <ViewStorage />}
        </div>
    );
};

export default Settings;