import React, { useState } from 'react';
import { Users, ChevronDown, ChevronRight, MoreVertical, Settings, FileText, UserPlus } from 'lucide-react';
import CreateSysAdminForm from './SuperAdmin/CreateSysAdminForm';
import ViewSysAdminAccount from './SuperAdmin/ViewSysAdminAccount';
import SystemLogs from './SuperAdmin/SystemLogs';

const SuperAdminDashboard = ({ user, onLogout }) => {
  const [isSysAdminOpen, setIsSysAdminOpen] = useState(true);
  const [selectedSection, setSelectedSection] = useState('Dashboard');

  const handleQuickAction = (action) => {
    setSelectedSection(action);
  };

  const renderDashboardContent = () => (
    <div className="space-y-8">
      <div>
        <h2 className="text-lg font-semibold mb-4">Quick Actions</h2>
        <div className="grid grid-cols-4 gap-4">
          <div 
            onClick={() => handleQuickAction('Create SysAdmin')}
            className="p-4 border-2 border-dashed rounded-lg hover:border-gray-400 cursor-pointer transition-colors duration-200 hover:bg-gray-50"
          >
            <div className="aspect-square flex items-center justify-center">
              <UserPlus size={40} className="text-gray-400" />
            </div>
            <p className="text-center mt-2 text-sm">Create SysAdmin</p>
          </div>
          <div 
            onClick={() => handleQuickAction('System Logs')}
            className="p-4 border-2 border-dashed rounded-lg hover:border-gray-400 cursor-pointer transition-colors duration-200 hover:bg-gray-50"
          >
            <div className="aspect-square flex items-center justify-center">
              <FileText size={40} className="text-gray-400" />
            </div>
            <p className="text-center mt-2 text-sm">View System Logs</p>
          </div>
        </div>
      </div>

      <div>
        <ViewSysAdminAccount />
      </div>
    </div>
  );

  const renderMainContent = () => {
    switch (selectedSection) {
      case 'Create SysAdmin':
        return (
          <div>
            <div className="mb-6">
              <h2 className="text-xl font-semibold">Create System Administrator</h2>
              <p className="text-gray-600">Add a new system administrator account</p>
            </div>
            <CreateSysAdminForm onSuccess={() => setSelectedSection('Dashboard')} />
          </div>
        );
      case 'View SysAdmin':
        return (
          <div>
            <div className="mb-6">
              <h2 className="text-xl font-semibold">System Administrators</h2>
              <p className="text-gray-600">Manage and monitor system administrator accounts</p>
            </div>
            <ViewSysAdminAccount />
          </div>
        );
      case 'System Logs':
        return (
          <div>
            <div className="mb-6">
              <h2 className="text-xl font-semibold">System Logs</h2>
              <p className="text-gray-600">Monitor system activity and track important events</p>
            </div>
            <SystemLogs />
          </div>
        );
      case 'Dashboard':
        return renderDashboardContent();
      default:
        return (
          <div className="flex items-center justify-center h-64 text-gray-500">
            This section is currently under development
          </div>
        );
    }
  };

  return (
    <div className="flex h-screen bg-white">
      <div className="w-64 bg-gray-900 text-white flex flex-col">
        <div className="p-6 border-b border-gray-600 flex justify-center items-center">
          <img src="/safesplit-logo_white.png" alt="Logo" className="h-24 w-auto" />
        </div>
        
        <nav className="flex-1 p-4">
          <ul className="space-y-2">
            <li>
              <button 
                onClick={() => setSelectedSection('Dashboard')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-600 
                  ${selectedSection === 'Dashboard' ? 'bg-gray-600' : ''}`}
              >
                Dashboard
              </button>
            </li>
            
            <li>
              <button 
                onClick={() => setIsSysAdminOpen(!isSysAdminOpen)}
                className={`w-full flex items-center justify-between px-4 py-2 rounded hover:bg-gray-600 transition-colors duration-200
                  ${selectedSection.includes('SysAdmin') ? 'bg-gray-600' : ''}`}
              >
                <span>SysAdmin Management</span>
                {isSysAdminOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
              </button>
              
              {isSysAdminOpen && (
                <ul className="ml-4 mt-2 space-y-1">
                  <li>
                    <button 
                      onClick={() => setSelectedSection('Create SysAdmin')}
                      className={`w-full text-left px-4 py-2 rounded transition-colors duration-200
                        ${selectedSection === 'Create SysAdmin' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      Create SysAdmin
                    </button>
                  </li>
                  <li>
                    <button 
                      onClick={() => setSelectedSection('View SysAdmin')}
                      className={`w-full text-left px-4 py-2 rounded transition-colors duration-200
                        ${selectedSection === 'View SysAdmin' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      View SysAdmin
                    </button>
                  </li>
                </ul>
              )}
            </li>

            <li>
              <button 
                onClick={() => setSelectedSection('System Logs')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-600
                  ${selectedSection === 'System Logs' ? 'bg-gray-600' : ''}`}
              >
                System Logs
              </button>
            </li>

            <li>
              <button 
                onClick={() => setSelectedSection('Settings')}
                className={`w-full text-left px-4 py-2 rounded transition-colors duration-200 hover:bg-gray-600
                  ${selectedSection === 'Settings' ? 'bg-gray-600' : ''}`}
              >
                Settings
              </button>
            </li>
          </ul>
        </nav>

        <div className="border-t border-gray-600 p-4">
          <button className="block w-full text-left px-4 py-2 hover:bg-gray-600 rounded transition-colors duration-200">
            Get Help
          </button>
          <button 
            onClick={onLogout}
            className="block w-full text-left px-4 py-2 hover:bg-gray-600 rounded transition-colors duration-200"
          >
            Logout
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        <div className="p-8">
          <div className="mb-8">
            <h1 className="text-2xl font-semibold">Super Admin Dashboard</h1>
            <p className="text-gray-600 mt-1">Manage system administrators and monitor system activity</p>
          </div>

          {renderMainContent()}
        </div>
      </div>
    </div>
  );
};

export default SuperAdminDashboard;