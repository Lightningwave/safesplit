import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login, getDashboardByRole } from '../../services/authService';

function LoginForm({ onLogin }) {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    twoFactorCode: ''
  });
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [requires2FA, setRequires2FA] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await login(formData.email, formData.password, formData.twoFactorCode);
      
      if (response.requires_2fa) {
        setRequires2FA(true);
        setError('Please check your email for the 2FA code');
      } else {
        onLogin(response.user);
        const dashboardRoute = getDashboardByRole(response.user.role);
        navigate(dashboardRoute);
      }
    } catch (err) {
      setError(err.message || 'Login failed');
      setRequires2FA(false);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded shadow">
      <h2 className="text-2xl font-bold mb-4">Login</h2>
      {error && (
        <div className={`p-4 mb-4 rounded-lg ${requires2FA ? 'bg-blue-100 text-blue-700' : 'bg-red-100 text-red-700'}`}>
          {error}
        </div>
      )}
      <form onSubmit={handleSubmit} className="space-y-4">
        {!requires2FA ? (
          <>
            <div>
              <input
                type="email"
                placeholder="Email"
                value={formData.email}
                onChange={(e) => setFormData({...formData, email: e.target.value})}
                className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                disabled={isLoading}
                required
              />
            </div>
            <div>
              <input
                type="password"
                placeholder="Password"
                value={formData.password}
                onChange={(e) => setFormData({...formData, password: e.target.value})}
                className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                disabled={isLoading}
                required
              />
            </div>
          </>
        ) : (
          <div>
            <input
              type="text"
              placeholder="Enter 2FA code from email"
              value={formData.twoFactorCode}
              onChange={(e) => setFormData({...formData, twoFactorCode: e.target.value})}
              className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isLoading}
              required
              autoFocus
            />
          </div>
        )}
        <button 
          type="submit"
          className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
          disabled={isLoading}
        >
          {isLoading ? 'Processing...' : requires2FA ? 'Verify Code' : 'Login'}
        </button>
      </form>
    </div>
  );
}

export default LoginForm;