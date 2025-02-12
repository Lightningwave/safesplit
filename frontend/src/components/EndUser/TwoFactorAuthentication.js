import React, { useState, useEffect } from 'react';
import { Loader } from 'lucide-react';

const TwoFactorAuthentication = () => {
  const [isEnabled, setIsEnabled] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [showVerification, setShowVerification] = useState(false);
  const [verificationCode, setVerificationCode] = useState('');
  const [isVerifying, setIsVerifying] = useState(false);
  const [verificationMode, setVerificationMode] = useState(null);
  
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
      
      if (!response.ok) {
        throw new Error('Failed to fetch 2FA status');
      }
      
      const data = await response.json();
      setIsEnabled(data.two_factor_enabled);
    } catch (err) {
      setError('Failed to fetch 2FA status');
    }
  };

  const initiateEnable2FA = async () => {
    setError('');
    setMessage('');
    setLoading(true);
    
    try {
      const response = await fetch('/api/2fa/enable/initiate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      setVerificationMode('enable');
      setShowVerification(true);
      setMessage('Please check your email for the verification code');
    } catch (err) {
      setError(err.message || 'Failed to initiate 2FA enable');
    } finally {
      setLoading(false);
    }
  };

  const initiateDisable2FA = async () => {
    setError('');
    setMessage('');
    setLoading(true);
    
    try {
      const response = await fetch('/api/2fa/disable/initiate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      setVerificationMode('disable');
      setShowVerification(true);
      setMessage('Please check your email for the verification code');
    } catch (err) {
      setError(err.message || 'Failed to initiate 2FA disable');
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async () => {
    setError('');
    setMessage('');
    setIsVerifying(true);
    
    const endpoint = verificationMode === 'enable' 
      ? '/api/2fa/enable/verify'
      : '/api/2fa/disable/verify';

    try {
      const response = await fetch(endpoint, {
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

      setIsEnabled(verificationMode === 'enable');
      setShowVerification(false);
      setVerificationCode('');
      setVerificationMode(null);
      setMessage(`2FA has been ${verificationMode === 'enable' ? 'enabled' : 'disabled'}`);
    } catch (err) {
      setError(err.message || 'Invalid verification code');
    } finally {
      setIsVerifying(false);
    }
  };

  const handleCancel = () => {
    setShowVerification(false);
    setVerificationCode('');
    setVerificationMode(null);
    setError('');
    setMessage('');
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
                onClick={handleVerify}
                disabled={isVerifying || !verificationCode}
                className={`px-4 py-2 text-white rounded disabled:opacity-50 flex items-center space-x-2 ${
                  verificationMode === 'enable' 
                    ? 'bg-green-600 hover:bg-green-700'
                    : 'bg-red-600 hover:bg-red-700'
                }`}
              >
                {isVerifying && <Loader size={16} className="animate-spin" />}
                <span>
                  Verify and {verificationMode === 'enable' ? 'Enable' : 'Disable'} 2FA
                </span>
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
            onClick={isEnabled ? initiateDisable2FA : initiateEnable2FA}
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