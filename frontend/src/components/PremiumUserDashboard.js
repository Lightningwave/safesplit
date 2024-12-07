// src/components/PremiumUserDashboard.js
import React, { useState } from 'react';
import { File, Settings, MoreVertical, ChevronDown, ChevronRight, Search, Upload } from 'lucide-react';

const PremiumUserDashboard = ({ user, onLogout }) => {
  const [isFilesOpen, setIsFilesOpen] = useState(true);
  const [selectedSection, setSelectedSection] = useState('My Files');

  const files = [
    { 
      id: 1, 
      name: 'Premium_Report_Q4.pdf',
      size: '9 MB',
      folder: 'Documents',
      lastModified: 'October 4, 2024',
      selected: true
    },
    { 
      id: 2, 
      name: 'Financial_Analysis_2024.xlsx',
      size: '17 MB',
      folder: 'Documents',
      lastModified: 'October 4, 2024',
      selected: false
    },
    { 
      id: 3, 
      name: 'Project_Premium_Plan.doc',
      size: '644 KB',
      folder: 'Documents',
      lastModified: 'October 4, 2024',
      selected: false
    },
    { 
      id: 4, 
      name: 'Budget_Overview_2024.pdf',
      size: '2 MB',
      folder: 'Documents',
      lastModified: 'October 4, 2024',
      selected: false
    },
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
            
            {/* Files Dropdown */}
            <li>
              <button 
                onClick={() => setIsFilesOpen(!isFilesOpen)}
                className="w-full flex items-center justify-between px-4 py-2 rounded hover:bg-gray-600 bg-gray-600"
              >
                <span>Files</span>
                {isFilesOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
              </button>
              
              {isFilesOpen && (
                <ul className="ml-4 mt-2 space-y-1">
                  <li>
                    <a 
                      href="#" 
                      onClick={() => setSelectedSection('My Files')}
                      className={`block px-4 py-2 rounded ${selectedSection === 'My Files' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      My Files
                    </a>
                  </li>
                  <li>
                    <a 
                      href="#" 
                      onClick={() => setSelectedSection('Uploaded Files')}
                      className={`block px-4 py-2 rounded ${selectedSection === 'Uploaded Files' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      Uploaded Files
                    </a>
                  </li>
                  <li>
                    <a 
                      href="#" 
                      onClick={() => setSelectedSection('Shared Files')}
                      className={`block px-4 py-2 rounded ${selectedSection === 'Shared Files' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      Shared Files
                    </a>
                  </li>
                  <li>
                    <a 
                      href="#" 
                      onClick={() => setSelectedSection('Archives')}
                      className={`block px-4 py-2 rounded ${selectedSection === 'Archives' ? 'bg-gray-500' : 'hover:bg-gray-600'}`}
                    >
                      Archives
                    </a>
                  </li>
                </ul>
              )}
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
            <a href="#" className="block px-4 py-2 hover:bg-gray-600 rounded">Contact Us</a>
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
          <div className="flex justify-between items-center mb-8">
            <h1 className="text-2xl font-semibold">My Files</h1>
            <div className="flex items-center space-x-4">
              {/* Search Bar */}
              <div className="flex items-center bg-gray-100 rounded-md px-3 py-2">
                <Search size={20} className="text-gray-400 mr-2" />
                <input
                  type="text"
                  placeholder="Search"
                  className="bg-transparent outline-none"
                />
              </div>
              
              {/* Upload Button */}
              <button className="flex items-center space-x-2 px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700">
                <Upload size={20} />
                <span>Upload File</span>
              </button>
            </div>
          </div>

          {/* Files Table */}
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
            {files.map(file => (
              <div key={file.id} className="grid grid-cols-12 gap-4 p-4 border-b hover:bg-gray-50 text-sm">
                <div className="col-span-5">
                  <div className="flex items-center space-x-4">
                    <input type="checkbox" checked={file.selected} className="rounded" />
                    <File size={20} className="text-gray-400" />
                    <span>{file.name}</span>
                  </div>
                </div>
                <div className="col-span-2">{file.size}</div>
                <div className="col-span-2">{file.folder}</div>
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
  );
};

export default PremiumUserDashboard;