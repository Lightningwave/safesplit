import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginSuperAdmin } from '../../services/authService';
import { AlertCircle } from 'lucide-react';

function SuperAdminLogin({ onLogin }) {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        email: '',
        password: ''
    });
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');

        try {
            const data = await loginSuperAdmin(formData.email, formData.password);
            if (data.user.role !== 'super_admin') {
                throw new Error('Unauthorized access');
            }
            onLogin(data.user);
            navigate('/super-dashboard');
        } catch (err) {
            setError(err.message || 'Access denied');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded shadow">
            <h2 className="text-2xl font-bold mb-4">Super Admin Login</h2>
            {error && (
                <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-md flex items-center mb-4">
                    <AlertCircle className="mr-2" size={16} />
                    {error}
                </div>
            )}
            <form onSubmit={handleSubmit}>
                <div className="mb-4">
                    <input
                        type="email"
                        placeholder="Super Admin Email"
                        value={formData.email}
                        onChange={(e) => setFormData({...formData, email: e.target.value})}
                        className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
                        disabled={isLoading}
                    />
                </div>
                <div className="mb-4">
                    <input
                        type="password"
                        placeholder="Password"
                        value={formData.password}
                        onChange={(e) => setFormData({...formData, password: e.target.value})}
                        className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
                        disabled={isLoading}
                    />
                </div>
                <button 
                    type="submit"
                    className="w-full bg-purple-600 text-white p-2 rounded hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 disabled:opacity-50"
                    disabled={isLoading}
                >
                    {isLoading ? 'Authenticating...' : 'Login as Super Admin'}
                </button>
            </form>
        </div>
    );
}

export default SuperAdminLogin;