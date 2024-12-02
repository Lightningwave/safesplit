import React, { useState } from 'react';
import { login } from '../../services/authService';

export default function LoginForm({ onLogin }) {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        
        if (!email || !password) {
            setError('Please fill in all fields');
            return;
        }

        try {
            const data = await login(email, password);
            onLogin(data.user);
        } catch (err) {
            setError(err.message || 'Invalid credentials');
            console.error('Login error:', err);
        }
    };

    return (
        <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded shadow">
            <h2 className="text-2xl font-bold mb-4">Login</h2>
            {error && <div className="text-red-500 mb-4">{error}</div>}
            <form onSubmit={handleSubmit}>
                <input
                    type="email"
                    placeholder="Email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full p-2 mb-4 border rounded"
                />
                <input
                    type="password"
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full p-2 mb-4 border rounded"
                />
                <button 
                    type="submit"
                    className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600"
                >
                    Login
                </button>
            </form>
        </div>
    );
}