import React, { useState } from 'react';
import LoginForm from './components/auth/LoginForm';
import RegisterForm from './components/auth/RegisterForm';
import { logout } from './services/authService';

function App() {
    const [user, setUser] = useState(() => {
        const savedUser = localStorage.getItem('user');
        return savedUser ? JSON.parse(savedUser) : null;
    });
    const [isLogin, setIsLogin] = useState(true);

    const handleLogin = (user) => {
        setUser(user);
    };

    const handleLogout = () => {
        logout();
        setUser(null);
    };

    const handleRegisterSuccess = () => {
        setIsLogin(true);
    };

    return (
        <div className="min-h-screen bg-gray-100 py-6">
            {user ? (
                <div className="max-w-md mx-auto p-6 bg-white rounded shadow">
                    <h2 className="text-2xl font-bold mb-4">Welcome, {user.username}!</h2>
                    <button
                        onClick={handleLogout}
                        className="w-full bg-red-500 text-white p-2 rounded hover:bg-red-600"
                    >
                        Logout
                    </button>
                </div>
            ) : (
                <div>
                    {isLogin ? (
                        <>
                            <LoginForm onLogin={handleLogin} />
                            <div className="text-center mt-4">
                                <button
                                    onClick={() => setIsLogin(false)}
                                    className="text-blue-500 hover:text-blue-700"
                                >
                                    Need an account? Register
                                </button>
                            </div>
                        </>
                    ) : (
                        <>
                            <RegisterForm onRegisterSuccess={handleRegisterSuccess} />
                            <div className="text-center mt-4">
                                <button
                                    onClick={() => setIsLogin(true)}
                                    className="text-blue-500 hover:text-blue-700"
                                >
                                    Already have an account? Login
                                </button>
                            </div>
                        </>
                    )}
                </div>
            )}
        </div>
    );
}

export default App;