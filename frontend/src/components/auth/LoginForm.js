import React, { useState, useEffect } from 'react';
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
  const [lockoutInfo, setLockoutInfo] = useState(null);
  const [remainingAttempts, setRemainingAttempts] = useState(null);

  useEffect(() => {
    let timer;
    if (lockoutInfo) {
      timer = setInterval(() => {
        const now = new Date();
        const lockUntil = new Date(lockoutInfo.locked_until);
        const remaining = Math.max(0, Math.ceil((lockUntil - now) / 1000 / 60));
        
        if (remaining <= 0) {
          setLockoutInfo(null);
        } else {
          setLockoutInfo(prev => ({...prev, remaining_minutes: remaining}));
        }
      }, 1000);
    }
    return () => clearInterval(timer);
  }, [lockoutInfo]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    setLockoutInfo(null);
    setRemainingAttempts(null);

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
      const responseData = err.response?.data;
      console.log('Full error response:', {
          status: err.response?.status,
          data: err.response?.data
      });
      
      if (err.response?.status === 429) {
          setLockoutInfo({
              locked_until: responseData.locked_until,
              remaining_minutes: responseData.remaining_minutes
          });
          setError(responseData.error);
      } else if (responseData?.status === 'failed' && responseData?.remaining_attempts) {
          setRemainingAttempts(responseData.remaining_attempts);
          setError(responseData.error);
      } else {
          setError(responseData?.error || 'Login failed');
      }
      setRequires2FA(false);
  } finally {
        setIsLoading(false);
    }
};

const getErrorDisplay = () => {
  if (!error) return null;

  if (requires2FA) {
    return (
      <div className="p-4 mb-4 rounded-lg bg-blue-100 text-blue-700">
        {error}
      </div>
    );
  }

  if (lockoutInfo) {
    return (
      <div className="p-4 mb-4 rounded-lg bg-orange-100 text-orange-700">
        <p>
          Account locked for {lockoutInfo.remaining_minutes} minute{lockoutInfo.remaining_minutes !== 1 ? 's' : ''}
        </p>
      </div>
    );
  }

  return (
    <div className={`p-4 mb-4 rounded-lg ${remainingAttempts ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}`}>
      {error}
    </div>
  );
};

  return (
    <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded shadow">
      <h2 className="text-2xl font-bold mb-4">Login</h2>
      {getErrorDisplay()}
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
                disabled={isLoading || lockoutInfo}
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
                disabled={isLoading || lockoutInfo}
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
          disabled={isLoading || lockoutInfo}
        >
          {isLoading ? 'Processing...' : 
           lockoutInfo ? `Account Locked (${lockoutInfo.remaining_minutes}m)` :
           requires2FA ? 'Verify Code' : 'Login'}
        </button>
      </form>
    </div>
  );
}

export default LoginForm;