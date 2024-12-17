import React, { useState } from 'react';
import { 
  Users, 
  ChevronDown, 
  ChevronRight, 
  HardDrive,
  Settings
} from 'lucide-react';
import ViewUserAccounts from './SysAdmin/ViewUserAccounts';
import ViewStorage from './SysAdmin/ViewStorage';
import ViewDeletedAccounts from './SysAdmin/ViewDeletedAccounts';

const SysAdminDashboard = ({ user, onLogout }) => {
  const [isAccountsOpen, setIsAccountsOpen] = useState(false); // Initially closed
  const [selectedSection, setSelectedSection] = useState('Dashboard');
  const [selectedUserType, setSelectedUserType] = useState('all');
  const [recentViewedUsers, setRecentViewedUsers] = useState([]); // For Recent Viewed Users

  const handleUserTypeSelect = (type) => {
    setSelectedUserType(type);
    setSelectedSection('Account Management'); // Navigate to Account Management section
    setIsAccountsOpen(false); // Close the dropdown after selection
  };

  const handleViewUser = (userId) => {

    const user = getUserById(userId); 
    if (user && !recentViewedUsers.find(u => u.id === user.id)) {
      setRecentViewedUsers(prev => [user, ...prev].slice(0, 5)); // Keep last 5
    }
  };

  const getUserById = (id) => {

    return null;
  };

  const renderDashboardContent = () => (
    <div className="space-y-8">
      {/* Accounts Section */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Accounts</h2>
        <div className="grid grid-cols-2 gap-4">
          <div 
            onClick={() => handleUserTypeSelect('premium')}
            className={`p-6 border-2 border-dashed rounded-lg cursor-pointer transition-colors duration-200 hover:bg-gray-50
              ${selectedUserType === 'premium' ? 'border-blue-500' : 'border-gray-300'}`}
          >
            <div className="aspect-square flex items-center justify-center">
              <Users size={40} className="text-gray-400" />
            </div>
            <p className="text-center mt-2 text-sm">Premium End users</p>
          </div>

          <div 
            onClick={() => handleUserTypeSelect('normal')}
            className={`p-6 border-2 border-dashed rounded-lg cursor-pointer transition-colors duration-200 hover:bg-gray-50
              ${selectedUserType === 'normal' ? 'border-blue-500' : 'border-gray-300'}`}
          >
            <div className="aspect-square flex items-center justify-center">
              <Users size={40} className="text-gray-400" />
            </div>
            <p className="text-center mt-2 text-sm">End users</p>
          </div>
        </div>
      </div>

    

      {/* Recent Viewed Users Section */}
      <div>
        <h2 className="text-lg font-semibold">Recent Viewed Users</h2>
        <div className="bg-white rounded-lg shadow p-4">
          {recentViewedUsers.length > 0 ? (
            <ul>
              {recentViewedUsers.map(user => (
                <li key={user.id} className="flex justify-between items-center py-2">
                  <span>{user.username}</span>
                  <button 
                    onClick={() => handleViewUser(user.id)}
                    className="text-blue-500 hover:underline text-sm"
                  >
                    View
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-gray-600">No recent users viewed.</p>
          )}
        </div>
      </div>
    </div>
  );

  const renderAccountManagementContent = () => (
    <div className="space-y-8">
      <ViewUserAccounts selectedType={selectedUserType} />
    </div>
  );

  const renderMainContent = () => {
    switch (selectedSection) {
      case 'Dashboard':
        return renderDashboardContent();
      case 'Account Management':
        return renderAccountManagementContent();
      case 'View Storage':
        return <ViewStorage />;
      case 'Deleted Accounts':
        return <ViewDeletedAccounts />;
      case 'Settings':
        return <Settings />; // Assuming you have a Settings component
      default:
        return renderDashboardContent();
    }
  };

  return (
    <div className="flex h-screen bg-gray-100">
      <div className="w-64 bg-gray-900 text-white flex flex-col">
        <div className="p-6 border-b border-gray-800 flex justify-center items-center">
          <img src="/safesplit-logo_white.png" alt="Logo" className="h-24 w-auto" />
        </div>
        
        <nav className="flex-1 p-4">
          <ul className="space-y-2">
            <li>
              <button 
                onClick={() => {
                  setSelectedSection('Dashboard');
                  setSelectedUserType('all');
                }}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-800
                  ${selectedSection === 'Dashboard' ? 'bg-gray-800' : ''}`}
              >
                Dashboard
              </button>
            </li>
            
            <li>
              <div>
                <button 
                  onClick={() => setIsAccountsOpen(!isAccountsOpen)}
                  className={`w-full flex items-center justify-between px-4 py-2 rounded transition-colors duration-200
                    hover:bg-gray-800 ${selectedSection === 'Account Management' ? 'bg-gray-800' : ''}`}
                  aria-haspopup="true"
                  aria-expanded={isAccountsOpen}
                >
                  <span>Account Management</span>
                  {isAccountsOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                </button>

                {/* Dropdown Sub-items */}
                {isAccountsOpen && (
                  <ul className="mt-2 space-y-1 pl-4">
                    <li>
                      <button 
                        onClick={() => handleUserTypeSelect('premium')}
                        className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-700
                          ${selectedUserType === 'premium' ? 'bg-gray-700' : ''}`}
                      >
                        Premium End user
                      </button>
                    </li>
                    <li>
                      <button 
                        onClick={() => handleUserTypeSelect('normal')}
                        className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-700
                          ${selectedUserType === 'normal' ? 'bg-gray-700' : ''}`}
                      >
                        End user
                      </button>
                    </li>
                  </ul>
                )}
              </div>
            </li>

            <li>
              <button 
                onClick={() => setSelectedSection('Deleted Accounts')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-800
                  ${selectedSection === 'Deleted Accounts' ? 'bg-gray-800' : ''}`}
              >
                Deleted Accounts
              </button>
            </li>

            <li>
              <button 
                onClick={() => setSelectedSection('View Storage')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-800
                  ${selectedSection === 'View Storage' ? 'bg-gray-800' : ''}`}
              >
                View Storage
              </button>
            </li>

            <li>
              <button 
                onClick={() => setSelectedSection('Settings')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-800
                  ${selectedSection === 'Settings' ? 'bg-gray-800' : ''}`}
              >
                Settings
              </button>
            </li>
          </ul>
        </nav>

        <div className="border-t border-gray-800 p-4">
          <button className="block w-full text-left px-4 py-2 hover:bg-gray-800 rounded transition-colors duration-200">
            Get Help
          </button>
          <button 
            onClick={onLogout}
            className="block w-full text-left px-4 py-2 hover:bg-gray-800 rounded transition-colors duration-200"
          >
            Logout
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        <div className="p-8">
          <div className="mb-8">
            <h1 className="text-2xl font-semibold">Admin Dashboard</h1>
            <p className="text-gray-600 mt-1">
              {selectedSection === 'Dashboard' 
                ? 'Monitor and manage system operations'
                : `Manage ${selectedSection.toLowerCase()}`}
            </p>
          </div>

          {renderMainContent()}
        </div>
      </div>
    </div>
  );
};

export default SysAdminDashboard;
