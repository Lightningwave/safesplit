import React, { useState, useEffect } from 'react';

const TwoFactorAuthentication = () => {
  const [isEnabled, setIsEnabled] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  
  useEffect(() => {
    fetchTwoFactorStatus();
  }, []);

  const fetchTwoFactorStatus = async () => {
    try {
      const response = await fetch('/api/2fa/status', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      const data = await response.json();
      setIsEnabled(data.two_factor_enabled);
    } catch (err) {
      setError('Failed to fetch 2FA status');
    }
  };

  const handleToggle2FA = async () => {
    setLoading(true);
    setError('');
    setMessage('');

    try {
      const response = await fetch(`/api/2fa/${isEnabled ? 'disable' : 'enable'}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      setIsEnabled(!isEnabled);
      setMessage(isEnabled ? '2FA has been disabled' : '2FA has been enabled');
    } catch (err) {
      setError(err.message || 'Failed to update 2FA settings');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Two-Factor Authentication</h2>
      
      {error && (
        <div className="p-4 mb-4 bg-red-50 text-red-700 rounded-md">
          {error}
        </div>
      )}
      
      {message && (
        <div className="p-4 mb-4 bg-green-50 text-green-700 rounded-md">
          {message}
        </div>
      )}

      <div className="mb-6">
        <p className="mb-4">
          {isEnabled 
            ? "Two-factor authentication is currently enabled."
            : "Enable two-factor authentication to add an extra layer of security to your account."}
        </p>
        
        <button
          onClick={handleToggle2FA}
          disabled={loading}
          className={`px-4 py-2 rounded transition-colors ${
            isEnabled 
              ? 'bg-red-600 hover:bg-red-700 text-white'
              : 'bg-green-600 hover:bg-green-700 text-white'
          } disabled:opacity-50`}
        >
          {loading 
            ? 'Processing...' 
            : isEnabled 
              ? 'Disable 2FA'
              : 'Enable 2FA'}
        </button>
      </div>
    </div>
  );
};

export default TwoFactorAuthentication;