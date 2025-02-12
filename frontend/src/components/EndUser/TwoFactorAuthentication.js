import React, { useState, useEffect } from 'react';
import { Loader } from 'lucide-react';

const TwoFactorAuthentication = () => {
  const [isEnabled, setIsEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [showVerification, setShowVerification] = useState(false);
  const [verificationCode, setVerificationCode] = useState('');
  const [isVerifying, setIsVerifying] = useState(false);
  
  useEffect(() => {
    fetchTwoFactorStatus();
  }, []);

  const fetchTwoFactorStatus = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/2fa/status', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch 2FA status');
      }
      
      const data = await response.json();
      setIsEnabled(data.two_factor_enabled);
    } catch (err) {
      setError('Failed to fetch 2FA status');
    } finally {
      setLoading(false);
    }
  };

  const handleEnable2FA = async () => {
    setError('');
    setMessage('');
    setLoading(true);
    
    try {
      const response = await fetch('http://localhost:8080/api/2fa/enable', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      setIsEnabled(true);
      setMessage('2FA has been enabled');
    } catch (err) {
      setError(err.message || 'Failed to enable 2FA');
    } finally {
      setLoading(false);
    }
  };

  const initiateDisable2FA = async () => {
    setError('');
    setMessage('');
    setLoading(true);
    
    try {
      const response = await fetch('http://localhost:8080/api/2fa/disable/initiate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      setShowVerification(true);
      setMessage('Please check your email for the verification code');
    } catch (err) {
      setError(err.message || 'Failed to initiate 2FA disable');
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyAndDisable = async () => {
    setError('');
    setMessage('');
    setIsVerifying(true);
    
    try {
      const response = await fetch('http://localhost:8080/api/2fa/disable/verify', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ code: verificationCode })
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      setIsEnabled(false);
      setShowVerification(false);
      setVerificationCode('');
      setMessage('2FA has been disabled');
    } catch (err) {
      setError(err.message || 'Invalid verification code');
    } finally {
      setIsVerifying(false);
    }
  };

  const handleCancel = () => {
    setShowVerification(false);
    setVerificationCode('');
    setError('');
    setMessage('');
  };

  if (loading && !isEnabled && !error) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader className="animate-spin mr-2" />
        <span>Loading 2FA status...</span>
      </div>
    );
  }

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
        
        {showVerification ? (
          <div className="space-y-4">
            <div>
              <label htmlFor="verificationCode" className="block text-sm font-medium text-gray-700 mb-1">
                Verification Code
              </label>
              <input
                id="verificationCode"
                type="text"
                value={verificationCode}
                onChange={(e) => setVerificationCode(e.target.value)}
                placeholder="Enter verification code"
                className="w-full p-2 border rounded focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                maxLength={6}
              />
            </div>
            <div className="flex space-x-4">
              <button
                onClick={handleVerifyAndDisable}
                disabled={isVerifying || !verificationCode}
                className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50 flex items-center space-x-2"
              >
                {isVerifying && <Loader size={16} className="animate-spin" />}
                <span>Verify and Disable 2FA</span>
              </button>
              <button
                onClick={handleCancel}
                className="px-4 py-2 border rounded hover:bg-gray-50"
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <button
            onClick={isEnabled ? initiateDisable2FA : handleEnable2FA}
            disabled={loading}
            className={`px-4 py-2 rounded transition-colors flex items-center space-x-2 ${
              isEnabled 
                ? 'bg-red-600 hover:bg-red-700 text-white'
                : 'bg-green-600 hover:bg-green-700 text-white'
            } disabled:opacity-50`}
          >
            {loading && <Loader size={16} className="animate-spin" />}
            <span>
              {loading 
                ? isEnabled ? 'Disabling 2FA...' : 'Enabling 2FA...'
                : isEnabled ? 'Disable 2FA' : 'Enable 2FA'}
            </span>
          </button>
        )}
      </div>
    </div>
  );
};

export default TwoFactorAuthentication;