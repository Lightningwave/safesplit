import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { register } from '../../services/authService';

function RegisterForm() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: ''
  });
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const validatePassword = (password) => {
    const hasUpper = /[A-Z]/.test(password);
    const hasLower = /[a-z]/.test(password);
    const hasNumber = /[0-9]/.test(password);
    const isLongEnough = password.length >= 8;

    return {
      isValid: hasUpper && hasLower && hasNumber && isLongEnough,
      requirements: {
        hasUpper,
        hasLower,
        hasNumber,
        isLongEnough
      }
    };
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    if (!formData.username || !formData.email || !formData.password) {
      setError('Please fill in all fields');
      setIsLoading(false);
      return;
    }

    if (formData.username.length < 3) {
      setError('Username must be at least 3 characters long');
      setIsLoading(false);
      return;
    }

    const passwordValidation = validatePassword(formData.password);
    if (!passwordValidation.isValid) {
      setError('Password does not meet requirements');
      setIsLoading(false);
      return;
    }

    try {
      await register(formData.username, formData.email, formData.password);
      navigate('/login');
    } catch (err) {
      console.error('Registration error:', err);
      setError(err.message || 'Registration failed');
    } finally {
      setIsLoading(false);
    }
  };

  const passwordValidation = validatePassword(formData.password);

  return (
    <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded shadow">
      <h2 className="text-2xl font-bold mb-4">Register</h2>
      {error && <div className="text-red-500 mb-4 p-2 bg-red-50 rounded">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">Username</label>
          <input
            type="text"
            placeholder="At least 3 characters"
            value={formData.username}
            onChange={(e) => setFormData({...formData, username: e.target.value})}
            className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={isLoading}
            minLength={3}
          />
        </div>
        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">Email</label>
          <input
            type="email"
            placeholder="your@email.com"
            value={formData.email}
            onChange={(e) => setFormData({...formData, email: e.target.value})}
            className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={isLoading}
          />
        </div>
        <div className="mb-6">
          <label className="block text-gray-700 text-sm font-bold mb-2">Password</label>
          <input
            type="password"
            placeholder="Password"
            value={formData.password}
            onChange={(e) => setFormData({...formData, password: e.target.value})}
            className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={isLoading}
          />
          <div className="mt-2 text-sm">
            <p className="font-semibold mb-1">Password requirements:</p>
            <ul className="space-y-1 text-gray-600">
              <li className={passwordValidation.requirements.isLongEnough ? "text-green-600" : ""}>
                ✓ At least 8 characters
              </li>
              <li className={passwordValidation.requirements.hasUpper ? "text-green-600" : ""}>
                ✓ One uppercase letter
              </li>
              <li className={passwordValidation.requirements.hasLower ? "text-green-600" : ""}>
                ✓ One lowercase letter
              </li>
              <li className={passwordValidation.requirements.hasNumber ? "text-green-600" : ""}>
                ✓ One number
              </li>
            </ul>
          </div>
        </div>
        <button 
          type="submit"
          className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          disabled={isLoading || !passwordValidation.isValid}
        >
          {isLoading ? 'Registering...' : 'Register'}
        </button>
      </form>
      <div className="mt-4 text-center text-gray-600">
        Already have an account?{' '}
        <button
          onClick={() => navigate('/login')}
          className="text-blue-500 hover:underline"
          disabled={isLoading}
        >
          Login here
        </button>
      </div>
    </div>
  );
}

export default RegisterForm;