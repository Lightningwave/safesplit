import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { LogOut, Home, User, Settings, CreditCard, Activity } from 'lucide-react';

function SideNavigation({ user, onLogout }) {
  const navigate = useNavigate();

  const handleLogout = (e) => {
    e.preventDefault();
    if (onLogout) {
      onLogout();
    }
    navigate('/');
  };

  const getNavigationItems = () => {
    const items = [
      { icon: <Home size={20} />, label: 'Dashboard', path: '/dashboard' }
    ];

    switch (user.role) {
      case 'end_user':
        items.push(
          { icon: <Activity size={20} />, label: 'Transactions', path: '/transactions' },
          { icon: <User size={20} />, label: 'Profile', path: '/profile' }
        );
        break;
      case 'premium_user':
        items.push(
          { icon: <Activity size={20} />, label: 'Transactions', path: '/transactions' },
          { icon: <CreditCard size={20} />, label: 'Premium Features', path: '/premium' },
          { icon: <User size={20} />, label: 'Profile', path: '/profile' }
        );
        break;
      case 'sys_admin':
        items.push(
          { icon: <Activity size={20} />, label: 'System Stats', path: '/admin/stats' },
          { icon: <Settings size={20} />, label: 'Settings', path: '/admin/settings' }
        );
        break;
      case 'super_admin':
        items.push(
          { icon: <Activity size={20} />, label: 'Global Stats', path: '/super/stats' },
          { icon: <Settings size={20} />, label: 'Admin Management', path: '/super/manage' }
        );
        break;
      default:
        break;
    }

    return items;
  };

  return (
    <div className="h-screen w-64 bg-gray-900 text-white fixed left-0 top-0 flex flex-col">
      {/* Logo Section */}
      <div className="p-4 border-b border-gray-800">
        <Link to="/" className="flex items-center">
          <span className="font-bold text-lg">Safesplit</span>
        </Link>
      </div>

      {/* User Info Section */}
      <div className="p-4 border-b border-gray-800">
        <div className="flex items-center">
          <div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center">
            {user.username.charAt(0).toUpperCase()}
          </div>
          <div className="ml-3">
            <p className="font-medium">{user.username}</p>
            <p className="text-sm text-gray-400">{user.role.replace('_', ' ')}</p>
          </div>
        </div>
      </div>

      {/* Navigation Items */}
      <nav className="flex-1 p-4">
        <ul className="space-y-2">
          {getNavigationItems().map((item, index) => (
            <li key={index}>
              <Link
                to={item.path}
                className="flex items-center space-x-3 px-4 py-2.5 rounded-lg hover:bg-gray-800 transition-colors"
              >
                {item.icon}
                <span>{item.label}</span>
              </Link>
            </li>
          ))}
        </ul>
      </nav>

      {/* Logout Section */}
      <div className="p-4 border-t border-gray-800">
        <button
          onClick={handleLogout}
          className="flex items-center space-x-3 px-4 py-2.5 w-full rounded-lg hover:bg-gray-800 transition-colors text-red-400 hover:text-red-300"
        >
          <LogOut size={20} />
          <span>Logout</span>
        </button>
      </div>
    </div>
  );
}

export default SideNavigation;