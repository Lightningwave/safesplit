import React, { useState, useEffect } from 'react';
import { 
  Users, 
  ChevronDown, 
  ChevronRight,
  HardDrive,
} from 'lucide-react';
import ViewUserAccounts from './SysAdmin/ViewUserAccounts';
import ViewStorage from './SysAdmin/ViewStorage';
import ViewDeletedAccounts from './SysAdmin/ViewDeletedAccounts';
import ViewFeedback from './SysAdmin/ViewFeedback';
import ViewReport from './SysAdmin/ViewReport';
import ViewBillingRecords from './SysAdmin/ViewBillingRecords';

const SysAdminDashboard = ({ user, onLogout }) => {
  const [isAccountsOpen, setIsAccountsOpen] = useState(false);
  const [selectedSection, setSelectedSection] = useState('Dashboard');
  const [selectedUserType, setSelectedUserType] = useState('all');
  const [recentUpdatedUsers, setRecentUpdatedUsers] = useState([]);

  // Fetch recently updated users
  const fetchRecentUsers = async () => {
    try {
      const response = await fetch('/api/system/users', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch recent users');
      }

      const data = await response.json();
      const usersList = Array.isArray(data.users) ? data.users : 
                       Array.isArray(data.data) ? data.data :
                       Array.isArray(data) ? data : [];

      // Sort users by updated_at timestamp
      const sortedUsers = usersList
        .filter(user => user && user.updated_at)
        .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
        .slice(0, 5);

      setRecentUpdatedUsers(sortedUsers);
    } catch (error) {
      console.error('Error fetching recent users:', error);
    }
  };

  // Fetch recent users on component mount
  useEffect(() => {
    fetchRecentUsers();
  }, []);

  const handleUserTypeSelect = (type) => {
    setSelectedUserType(type);
    setSelectedSection('Account Management');
    setIsAccountsOpen(false);
  };

  const handleUserViewed = (viewedUser) => {
    if (!viewedUser) return;
    
    setRecentUpdatedUsers(prev => {
      // Remove the user if already in the list
      const filteredUsers = prev.filter(u => u.id !== viewedUser.id);
      // Add the user at the beginning and maintain sorting
      const updatedUsers = [viewedUser, ...filteredUsers]
        .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
        .slice(0, 5);
      return updatedUsers;
    });
  };

  // Refresh recent users when returning to dashboard
  useEffect(() => {
    if (selectedSection === 'Dashboard') {
      fetchRecentUsers();
    }
  }, [selectedSection]);

  const renderDashboardContent = () => (
    <div className="space-y-8">
      {/* Accounts Section */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Account Types</h2>
        <div className="grid grid-cols-2 gap-4">
          <div 
            onClick={() => handleUserTypeSelect('premium')}
            className={`p-6 border-2 border-dashed rounded-lg cursor-pointer transition-colors duration-200 hover:bg-gray-50
              ${selectedUserType === 'premium' ? 'border-blue-500' : 'border-gray-300'}`}
          >
            <div className="aspect-square flex items-center justify-center">
              <Users size={40} className="text-gray-400" />
            </div>
            <p className="text-center mt-2 text-sm">Premium End Users</p>
          </div>

          <div 
            onClick={() => handleUserTypeSelect('normal')}
            className={`p-6 border-2 border-dashed rounded-lg cursor-pointer transition-colors duration-200 hover:bg-gray-50
              ${selectedUserType === 'normal' ? 'border-blue-500' : 'border-gray-300'}`}
          >
            <div className="aspect-square flex items-center justify-center">
              <Users size={40} className="text-gray-400" />
            </div>
            <p className="text-center mt-2 text-sm">End Users</p>
          </div>
        </div>
      </div>

      {/* Recent Updated Users Section */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Recent Updated Users</h2>
        <div className="bg-white rounded-lg shadow">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">UserID</th>
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Name</th>
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Account Type</th>
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Last Updated</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {recentUpdatedUsers.length > 0 ? (
                recentUpdatedUsers.map(user => (
                  <tr key={user.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm text-gray-900">{user.id}</td>
                    <td className="px-6 py-4 text-sm text-gray-900">{user.username}</td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {user.subscription_status === 'premium' ? 'Premium' : 'Normal'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'Never'}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="4" className="px-6 py-8 text-center text-gray-500">
                    No recent updated users.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );

  const renderMainContent = () => {
    switch (selectedSection) {
      case 'Dashboard':
        return renderDashboardContent();
      case 'Account Management':
        return <ViewUserAccounts 
          selectedType={selectedUserType}
          onUserViewed={handleUserViewed} 
        />;
      case 'View Storage':
        return <ViewStorage />;
      case 'Deleted Accounts':
        return <ViewDeletedAccounts />;
      case 'Feedback':
        return <ViewFeedback feedbackType="feedback" />;
      case 'Reports':
        return <ViewReport feedbackType="suspicious_activity" />;
      case 'Billing Records':
        return <ViewBillingRecords />;
      default:
        return renderDashboardContent();
    }
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <div className="w-64 bg-gray-900 text-white flex flex-col">
        <div className="p-6 border-b border-gray-800 flex justify-center items-center">
          <img src="/safesplit-logo_white.png" alt="Logo" className="h-24 w-auto" />
        </div>
        
        {/* Navigation */}
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
                >
                  <span>Account Management</span>
                  {isAccountsOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                </button>

                {isAccountsOpen && (
                  <ul className="mt-2 space-y-1 pl-4">
                    <li>
                      <button 
                        onClick={() => handleUserTypeSelect('all')}
                        className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-700
                          ${selectedUserType === 'all' ? 'bg-gray-700' : ''}`}
                      >
                        All Users
                      </button>
                    </li>
                    <li>
                      <button 
                        onClick={() => handleUserTypeSelect('premium')}
                        className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-700
                          ${selectedUserType === 'premium' ? 'bg-gray-700' : ''}`}
                      >
                        Premium End Users
                      </button>
                    </li>
                    <li>
                      <button 
                        onClick={() => handleUserTypeSelect('normal')}
                        className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-700
                          ${selectedUserType === 'normal' ? 'bg-gray-700' : ''}`}
                      >
                        End Users
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
                onClick={() => setSelectedSection('Feedback')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-800
                  ${selectedSection === 'Feedback' ? 'bg-gray-800' : ''}`}
              >
                User Feedback
              </button>
            </li>

            <li>
              <button 
                onClick={() => setSelectedSection('Reports')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-800
                  ${selectedSection === 'Reports' ? 'bg-gray-800' : ''}`}
              >
                Suspicious Reports
              </button>
            </li>

            <li>
              <button 
                onClick={() => setSelectedSection('Billing Records')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-800
                  ${selectedSection === 'Billing Records' ? 'bg-gray-800' : ''}`}
              >
                Billing Records
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
          </ul>
        </nav>

        <div className="border-t border-gray-800 p-4">
          <button 
            onClick={onLogout}
            className="block w-full text-left px-4 py-2 hover:bg-gray-800 rounded transition-colors duration-200"
          >
            Logout
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto">
        <div className="p-8">
          <div className="mb-8">
            <h1 className="text-2xl font-semibold">
              {selectedSection === 'Dashboard' ? 'Admin Dashboard' : 
               selectedSection === 'Feedback' ? 'User Feedback' :
               selectedSection === 'Reports' ? 'Suspicious Activity Reports' :
               `${selectedSection}`}
            </h1>
            <p className="text-gray-600 mt-1">
              {selectedSection === 'Dashboard' ? 'Monitor and manage system operations' :
               selectedSection === 'Feedback' ? 'View and manage user feedback and suggestions' :
               selectedSection === 'Reports' ? 'Review and investigate reported activities' :
               `Manage ${selectedSection.toLowerCase()}`}
            </p>
          </div>

          {renderMainContent()}
        </div>
      </div>
    </div>
  );
};

export default SysAdminDashboard;