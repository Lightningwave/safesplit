import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginSuperAdmin } from '../../services/authService';
import { AlertCircle } from 'lucide-react';

function SuperAdminLogin({ onLogin }) {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        email: '',
        password: '',
        twoFactorCode: ''
    });
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [requires2FA, setRequires2FA] = useState(false);
    const [userId, setUserId] = useState(null);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');

        try {
            const response = await loginSuperAdmin(
                formData.email, 
                formData.password,
                formData.twoFactorCode
            );
            
            // Handle 2FA requirement
            if (response.requires_2fa) {
                setRequires2FA(true);
                setUserId(response.user_id);
                setError('Please check your email for 2FA code');
                setIsLoading(false);
                return;
            }

            // Normal login success
            onLogin(response.user);
            navigate('/super-dashboard');
        } catch (err) {
            setError(err.message || 'Access denied');
        } finally {
            setIsLoading(false);
        }
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    return (
        <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded shadow">
            <h2 className="text-2xl font-bold mb-4">Super Admin Login</h2>
            {error && (
                <div className={`border px-4 py-3 rounded-md flex items-center mb-4 ${
                    requires2FA ? 'bg-blue-50 border-blue-200 text-blue-600' : 'bg-red-50 border-red-200 text-red-600'
                }`}>
                    <AlertCircle className="mr-2" size={16} />
                    {error}
                </div>
            )}
            <form onSubmit={handleSubmit}>
                <div className="mb-4">
                    <input
                        type="email"
                        name="email"
                        placeholder="Super Admin Email"
                        value={formData.email}
                        onChange={handleChange}
                        className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
                        disabled={isLoading || requires2FA}
                        required
                    />
                </div>
                <div className="mb-4">
                    <input
                        type="password"
                        name="password"
                        placeholder="Password"
                        value={formData.password}
                        onChange={handleChange}
                        className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
                        disabled={isLoading || requires2FA}
                        required
                    />
                </div>
                {requires2FA && (
                    <div className="mb-4">
                        <input
                            type="text"
                            name="twoFactorCode"
                            placeholder="Enter 2FA Code"
                            value={formData.twoFactorCode}
                            onChange={handleChange}
                            className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
                            disabled={isLoading}
                            required
                        />
                    </div>
                )}
                <button 
                    type="submit"
                    className="w-full bg-purple-600 text-white p-2 rounded hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 disabled:opacity-50"
                    disabled={isLoading}
                >
                    {isLoading ? 'Authenticating...' : requires2FA ? 'Verify 2FA Code' : 'Login as Super Admin'}
                </button>
            </form>
        </div>
    );
}

export default SuperAdminLogin;