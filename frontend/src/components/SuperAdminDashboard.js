import React, { useState } from 'react';
import { Users, ChevronDown, ChevronRight, MoreVertical, Settings, FileText, UserPlus } from 'lucide-react';
import CreateSysAdminForm from '../components/Superadmin/CreateSysAdminForm';

const SuperAdminDashboard = ({ user, onLogout }) => {
  const [isSysAdminOpen, setIsSysAdminOpen] = useState(true);
  const [selectedSection, setSelectedSection] = useState('Dashboard');
  const [sysAdmins, setSysAdmins] = useState([
    { 
      id: 1, 
      name: 'John Admin', 
      email: 'john@admin.com',
      status: 'Active', 
      lastActive: 'October 4, 2024',
      selected: false 
    },
    { 
      id: 2, 
      name: 'Sarah Admin', 
      email: 'sarah@admin.com',
      status: 'Active', 
      lastActive: 'October 4, 2024',
      selected: false 
    },
    { 
      id: 3, 
      name: 'Mike Admin', 
      email: 'mike@admin.com',
      status: 'Inactive', 
      lastActive: 'October 3, 2024',
      selected: false 
    }
  ]);

  const handleQuickAction = (action) => {
    setSelectedSection(action);
  };

  const renderDashboardContent = () => (
    <>
      {/* Quick Actions */}
      <div className="mb-8">
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

      {/* SysAdmins Table */}
      <div>
        <h2 className="text-lg font-semibold mb-4">System Administrators</h2>
        <div className="border rounded-lg">
          {/* Table Header */}
          <div className="grid grid-cols-12 gap-4 p-4 border-b bg-gray-50 text-sm font-medium">
            <div className="col-span-1">Select</div>
            <div className="col-span-2">Admin ID</div>
            <div className="col-span-3">Name</div>
            <div className="col-span-3">Email</div>
            <div className="col-span-1">Status</div>
            <div className="col-span-1">Last Active</div>
            <div className="col-span-1">Actions</div>
          </div>

          {/* Table Body */}
          {sysAdmins.map(admin => (
            <div key={admin.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50 text-sm">
              <div className="col-span-1">
                <input 
                  type="checkbox" 
                  checked={admin.selected} 
                  onChange={() => {
                    setSysAdmins(sysAdmins.map(a => 
                      a.id === admin.id ? { ...a, selected: !a.selected } : a
                    ));
                  }}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
              </div>
              <div className="col-span-2">{admin.id}</div>
              <div className="col-span-3">{admin.name}</div>
              <div className="col-span-3">{admin.email}</div>
              <div className="col-span-1">
                <span className={`px-2 py-1 rounded-full text-xs ${
                  admin.status === 'Active' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  {admin.status}
                </span>
              </div>
              <div className="col-span-1">{admin.lastActive}</div>
              <div className="col-span-1">
                <button className="p-1 hover:bg-gray-100 rounded transition-colors duration-200">
                  <MoreVertical size={16} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );

  const renderMainContent = () => {
    switch (selectedSection) {
      case 'Create SysAdmin':
        return <CreateSysAdminForm onSuccess={() => setSelectedSection('Dashboard')} />;
      case 'Dashboard':
        return renderDashboardContent();
      default:
        return <div>Section under development</div>;
    }
  };

  return (
    <div className="flex h-screen bg-white">
      {/* Sidebar */}
      <div className="w-64 bg-gray-700 text-white flex flex-col">
        <div className="p-6 border-b border-gray-600 flex justify-center items-center">
          <img src="/safesplit-logo_nobg.png" alt="Logo" className="h-24 w-auto" />
        </div>
        
        {/* Navigation */}
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
            
            {/* SysAdmin Management Dropdown */}
            <li>
              <button 
                onClick={() => setIsSysAdminOpen(!isSysAdminOpen)}
                className="w-full flex items-center justify-between px-4 py-2 rounded hover:bg-gray-600 transition-colors duration-200"
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

        {/* Bottom Section */}
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

      {/* Main Content */}
      <div className="flex-1 overflow-auto">
        <div className="p-8">
          {/* Header */}
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