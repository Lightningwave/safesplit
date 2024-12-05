// src/components/EndUserDashboard.js
import React from 'react';
import { File, Settings, MoreVertical } from 'lucide-react';
import SideNavigation from './SideNavigation';

const EndUserDashboard = ({ user, onLogout }) => {
  const folders = [
    { id: 1, name: 'Folder One' },
    { id: 2, name: 'Folder Two' },
  ];

  const recentFiles = [
    { 
      id: 1, 
      name: 'Security_Audit_Reports.txt',
      size: '9 MB',
      type: 'Documents',
      lastModified: 'October 4, 2024',
      selected: true
    },
    { 
      id: 2, 
      name: 'Security_Note_Findings.doc',
      size: '17 MB',
      type: 'Documents',
      lastModified: 'October 4, 2024',
      selected: false
    },
    { 
      id: 3, 
      name: 'Security_Compliance_Checklist.xls',
      size: '644 KB',
      type: 'Documents',
      lastModified: 'October 4, 2024',
      selected: false
    },
    { 
      id: 4, 
      name: 'Security_Incident_Report.pdf',
      size: '2 MB',
      type: 'Documents',
      lastModified: 'October 4, 2024',
      selected: false
    },
  ];

  return (
    <div className="flex h-screen bg-white">
      {/* Dark Sidebar */}
      <div className="w-64 bg-gray-700 text-white">
        <div className="p-4 border-b border-gray-600">
          <img src="/safesplit-logo.png" alt="Logo" className="h-6" />
        </div>
        <nav className="p-4">
          <ul className="space-y-2">
            <li className="bg-gray-600 rounded">
              <a href="#" className="block px-4 py-2">Dashboard</a>
            </li>
            <li>
              <a href="#" className="block px-4 py-2">Files</a>
            </li>
            <li>
              <a href="#" className="block px-4 py-2">Settings</a>
            </li>
          </ul>
        </nav>
        <div className="absolute bottom-0 w-64 border-t border-gray-600">
          <div className="p-4">
            <a href="#" className="block px-4 py-2">Contact Us</a>
            <button onClick={onLogout} className="block w-full text-left px-4 py-2">
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto">
        <div className="p-8">
          <h1 className="text-2xl font-semibold mb-6">My Dashboard</h1>

          {/* Folders Section */}
          <div className="mb-8">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold">Folders</h2>
              <button className="px-4 py-2 bg-gray-100 rounded-md hover:bg-gray-200">
                New Folder
              </button>
            </div>
            <div className="grid grid-cols-6 gap-4">
              {folders.map(folder => (
                <div key={folder.id} className="p-4 border-2 border-dashed rounded-lg hover:border-gray-400 cursor-pointer">
                  <div className="aspect-square flex items-center justify-center">
                    <File size={40} className="text-gray-400" />
                  </div>
                  <p className="text-center mt-2 text-sm">{folder.name}</p>
                </div>
              ))}
            </div>
          </div>

          {/* Recents Section */}
          <div>
            <h2 className="text-lg font-semibold mb-4">Recents</h2>
            <div className="border rounded-lg">
              {/* Table Header */}
              <div className="grid grid-cols-12 gap-4 p-4 border-b bg-gray-50 text-sm font-medium">
                <div className="col-span-5">
                  <div className="flex items-center space-x-4">
                    <input type="checkbox" className="rounded" />
                    <span>Name</span>
                  </div>
                </div>
                <div className="col-span-2">Size</div>
                <div className="col-span-2">Folder</div>
                <div className="col-span-2">Last Modified</div>
                <div className="col-span-1"></div>
              </div>

              {/* Table Body */}
              {recentFiles.map(file => (
                <div key={file.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50 text-sm">
                  <div className="col-span-5">
                    <div className="flex items-center space-x-4">
                      <input type="checkbox" checked={file.selected} className="rounded" />
                      <File size={20} className="text-gray-400" />
                      <span>{file.name}</span>
                    </div>
                  </div>
                  <div className="col-span-2">{file.size}</div>
                  <div className="col-span-2">{file.type}</div>
                  <div className="col-span-2">{file.lastModified}</div>
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

export default EndUserDashboard;