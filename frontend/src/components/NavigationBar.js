import React, { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Menu, X } from 'lucide-react';

function NavigationBar({ user, onLogout }) {
    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const location = useLocation();
    const navigate = useNavigate();

    const scrollToSection = (sectionId) => {
        setIsMenuOpen(false);
        if (location.pathname !== '/') {
            window.location.href = `/#${sectionId}`;
            return;
        }
        const element = document.getElementById(sectionId);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth' });
        }
    };

    const handleLogout = () => {
        onLogout();
        navigate('/', { replace: true });
    };

    return (
        <header className="px-4 lg:px-6 h-14 flex items-center border-b fixed w-full bg-white top-0 z-50">
            <div className="container mx-auto max-w-[1200px] flex items-center justify-between w-full">
                <Link to="/" className="flex items-center">
                    <img 
                        src="/safesplit-logo_nobg.png" 
                        alt="Safesplit Logo" 
                        className="h-8 w-auto"
                    />
                    <span className="ml-2 font-bold text-lg">Safesplit</span>
                </Link>

                {/* Desktop Navigation */}
                <nav className="hidden md:flex items-center space-x-6">
                    {!user && (
                        <>
                            <button 
                                onClick={() => scrollToSection('how-it-works')} 
                                className="text-sm text-gray-600 hover:text-gray-900"
                            >
                                How It Works
                            </button>
                            <button 
                                onClick={() => scrollToSection('pricing')} 
                                className="text-sm text-gray-600 hover:text-gray-900"
                            >
                                Pricing
                            </button>
                        </>
                    )}
                    
                    <Link to="/about" className="text-sm text-gray-600 hover:text-gray-900">
                        About Us
                    </Link>
                    
                    {user ? (
                        <div className="flex items-center space-x-4">
                            <span className="text-sm text-gray-600">
                                Welcome, {user.username}
                            </span>
                            <button
                                onClick={handleLogout}
                                className="text-sm px-4 py-2 bg-black text-white rounded-md hover:bg-gray-800"
                            >
                                Logout
                            </button>
                        </div>
                    ) : (
                        <>
                            <Link 
                                to="/login" 
                                className="text-sm font-medium hover:text-gray-900"
                            >
                                Login
                            </Link>
                            <Link 
                                to="/register"
                                className="text-sm px-4 py-2 bg-black text-white rounded-md hover:bg-gray-800"
                            >
                                Register
                            </Link>
                        </>
                    )}
                </nav>

                {/* Mobile Navigation */}
{isMenuOpen && (
    <div className="absolute top-14 left-0 right-0 bg-white border-b md:hidden">
        <div className="flex flex-col p-4 space-y-4">
            <div className="flex flex-col items-center space-y-4"> {/* Center align container */}
                {!user && (
                    <>
                        <button 
                            onClick={() => scrollToSection('how-it-works')} 
                            className="text-sm text-gray-600 hover:text-gray-900"
                        >
                            How It Works
                        </button>
                        <button 
                            onClick={() => scrollToSection('pricing')} 
                            className="text-sm text-gray-600 hover:text-gray-900"
                        >
                            Pricing
                        </button>
                    </>
                )}
                
                <Link 
                    to="/about" 
                    className="text-sm text-gray-600 hover:text-gray-900"
                    onClick={() => setIsMenuOpen(false)}
                >
                    About Us
                </Link>
                
                {user ? (
                    <>
                        <span className="text-sm text-gray-600">
                            Welcome, {user.username}
                        </span>
                        <button
                            onClick={() => {
                                handleLogout();
                                setIsMenuOpen(false);
                            }}
                            className="text-sm px-4 py-2 bg-black text-white rounded-md hover:bg-gray-800"
                        >
                            Logout
                        </button>
                    </>
                ) : (
                    <>
                        <Link 
                            to="/login" 
                            className="text-sm font-medium hover:text-gray-900"
                            onClick={() => setIsMenuOpen(false)}
                        >
                            Login
                        </Link>
                        <Link 
                            to="/register"
                            className="text-sm px-4 py-2 bg-black text-white rounded-md hover:bg-gray-800"
                            onClick={() => setIsMenuOpen(false)}
                        >
                            Register
                        </Link>
                    </>
                )}
            </div>
        </div>
    </div>
)}
            </div>
        </header>
    );
}

export default NavigationBar;