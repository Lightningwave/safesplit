
import React, { useState } from 'react';
import { Users, ChevronDown, ChevronRight, MoreVertical, Settings } from 'lucide-react';

const SysAdminDashboard = ({ user, onLogout }) => {
  const [isAccountsOpen, setIsAccountsOpen] = useState(true);
  const [selectedSection, setSelectedSection] = useState('Premium EndUsers');

  const recentUsers = [
    { 
      id: 1, 
      name: 'Hong Na', 
      additionalInfo: 'P',
      accountType: 'Premium', 
      lastViewed: 'October 4, 2024',
      selected: true 
    },
    { 
      id: 2, 
      name: 'Xiao Juan', 
      additionalInfo: 'P',
      accountType: 'Premium', 
      lastViewed: 'October 4, 2024',
      selected: false 
    },
    { 
      id: 3, 
      name: 'Ryan', 
      additionalInfo: '',
      accountType: 'Normal', 
      lastViewed: 'October 4, 2024',
      selected: false 
    },
    { 
      id: 4, 
      name: 'Wei Pin', 
      additionalInfo: '',
      accountType: 'Normal', 
      lastViewed: 'October 4, 2024',
      selected: false 
    }
  ];

  return (
    <div className="flex h-screen bg-white">
      {/* Sidebar */}
      <div className="w-64 bg-gray-700 text-white">
        <div className="p-6 border-b border-gray-600 flex justify-center items-center">
          <img src="/safesplit-logo_nobg.png" alt="Logo" className="h-24 w-auto" />
        </div>
        
        {/* Navigation */}
        <nav className="p-4">
          <ul className="space-y-2">
            <li>
              <a href="#" className="block px-4 py-2 rounded hover:bg-gray-600">
                Dashboard
              </a>
            </li>
            
            {/* Account Management Dropdown */}
            <li>
              <button 
                onClick={() => setIsAccountsOpen(!isAccountsOpen)}
                className="w-full flex items-center justify-between px-4 py-2 rounded hover:bg-gray-600 bg-gray-600"
              >
                <span>Account Management</span>
                {isAccountsOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
              </button>
              
              {isAccountsOpen && (
                <ul className="ml-4 mt-2 space-y-1">
                  <li>
                    <a 
                      href="#" 
                      onClick={() => setSelectedSection('Premium EndUsers')}
                      className={`block px-4 py-2 rounded ${selectedSection === 'Premium EndUsers' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      Premium EndUsers
                    </a>
                  </li>
                  <li>
                    <a 
                      href="#" 
                      onClick={() => setSelectedSection('End Users')}
                      className={`block px-4 py-2 rounded ${selectedSection === 'End Users' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      End Users
                    </a>
                  </li>
                </ul>
              )}
            </li>

            <li>
              <a href="#" className="block px-4 py-2 rounded hover:bg-gray-600">
                Deleted Accounts
              </a>
            </li>
            
            <li>
              <a href="#" className="block px-4 py-2 rounded hover:bg-gray-600">
                View Storage
              </a>
            </li>

            <li>
              <a href="#" className="block px-4 py-2 rounded hover:bg-gray-600">
                Settings
              </a>
            </li>
          </ul>
        </nav>

        {/* Bottom Section */}
        <div className="absolute bottom-0 w-64 border-t border-gray-600">
          <div className="p-4">
            <a href="#" className="block px-4 py-2 hover:bg-gray-600 rounded">Get Help</a>
            <button onClick={onLogout} className="block w-full text-left px-4 py-2 hover:bg-gray-600 rounded">
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto">
        <div className="p-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-2xl font-semibold">Admin Dashboard</h1>
          </div>

          {/* Account Types */}
          <div className="mb-8">
            <h2 className="text-lg font-semibold mb-4">Accounts</h2>
            <div className="grid grid-cols-4 gap-4">
              <div className="p-4 border-2 border-dashed rounded-lg hover:border-gray-400 cursor-pointer">
                <div className="aspect-square flex items-center justify-center">
                  <Users size={40} className="text-gray-400" />
                </div>
                <p className="text-center mt-2 text-sm">Premium EndUsers</p>
              </div>
              <div className="p-4 border-2 border-dashed rounded-lg hover:border-gray-400 cursor-pointer">
                <div className="aspect-square flex items-center justify-center">
                  <Users size={40} className="text-gray-400" />
                </div>
                <p className="text-center mt-2 text-sm">End Users</p>
              </div>
            </div>
          </div>

          {/* Recent Viewed Users Table */}
          <div>
            <h2 className="text-lg font-semibold mb-4">Recent Viewed Users</h2>
            <div className="border rounded-lg">
              {/* Table Header */}
              <div className="grid grid-cols-12 gap-4 p-4 border-b bg-gray-50 text-sm font-medium">
                <div className="col-span-1">Select All</div>
                <div className="col-span-3">User ID</div>
                <div className="col-span-3">Name</div>
                <div className="col-span-2">Account Type</div>
                <div className="col-span-2">Last Viewed</div>
                <div className="col-span-1">Actions</div>
              </div>

              {/* Table Body */}
              {recentUsers.map(user => (
                <div key={user.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50 text-sm">
                  <div className="col-span-1">
                    <input type="checkbox" checked={user.selected} className="rounded" />
                  </div>
                  <div className="col-span-3">{user.id}</div>
                  <div className="col-span-3">{user.name} - {user.additionalInfo}</div>
                  <div className="col-span-2">{user.accountType}</div>
                  <div className="col-span-2">{user.lastViewed}</div>
                  <div className="col-span-1">
                    <button className="p-1 hover:bg-gray-100 rounded">
                      <MoreVertical size={16} />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SysAdminDashboard;