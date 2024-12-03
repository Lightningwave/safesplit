import React from 'react';
import { Link, useLocation } from 'react-router-dom';

function NavigationBar({ user, onLogout }) {
  const location = useLocation();

  const scrollToSection = (sectionId) => {
    if (location.pathname !== '/') {
      window.location.href = `/#${sectionId}`;
      return;
    }
    
    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  };

  return (
    <header className="px-4 lg:px-6 h-14 flex items-center border-b">
      <div className="container mx-auto max-w-[1200px] flex items-center justify-between">
        <Link to="/" className="flex items-center">
          <img 
            src="/safesplit-logo.png" 
            alt="Safesplit Logo" 
            className="h-10 w-auto"
          />
          <span className="ml-2 font-bold text-xl">Safesplit</span>
        </Link>
        
        <nav className="flex items-center space-x-6">
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
          <Link to="/about" className="text-sm text-gray-600 hover:text-gray-900">
            About Us
          </Link>
          
          <div className="flex items-center space-x-4">
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
          </div>
        </nav>
      </div>
    </header>
  );
}

export default NavigationBar;