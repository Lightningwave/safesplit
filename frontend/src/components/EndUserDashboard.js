import React, { useState } from 'react';
import { Settings, ChevronDown, ChevronRight, Search, Upload } from 'lucide-react';
import UploadFile from './EndUser/UploadFile';
import ViewFile from './EndUser/ViewFile';

const EndUserDashboard = ({ user, onLogout }) => {
    const [isFilesOpen, setIsFilesOpen] = useState(true);
    const [selectedSection, setSelectedSection] = useState('My Files');
    const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');

    const handleSearch = (event) => {
        setSearchQuery(event.target.value);
    };

    const handleUploadComplete = () => {
        // Force a refresh of the ViewFile component
        setSearchQuery(searchQuery);
    };

    return (
        <div className="flex h-screen bg-white">
            <div className="w-64 bg-gray-700 text-white">
                <div className="p-1 border-b border-gray-600 flex justify-center items-center">
                    <img src="/safesplit-logo_white.png" alt="SafeSplit Logo" className="h-32 w-auto" />
                </div>
                
                <nav className="p-4">
                    <ul className="space-y-2">
                        <li>
                            <a href="#" className="block px-4 py-2 rounded hover:bg-gray-600 transition-colors">
                                Dashboard
                            </a>
                        </li>
                        
                        <li>
                            <button 
                                onClick={() => setIsFilesOpen(!isFilesOpen)}
                                className="w-full flex items-center justify-between px-4 py-2 rounded hover:bg-gray-600 bg-gray-600 transition-colors"
                            >
                                <span>Files</span>
                                {isFilesOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                            </button>
                            
                            {isFilesOpen && (
                                <ul className="ml-4 mt-2 space-y-1">
                                    {['My Files', 'Uploaded Files', 'Shared Files', 'Archives'].map((section) => (
                                        <li key={section}>
                                            <button 
                                                onClick={() => setSelectedSection(section)}
                                                className={`w-full text-left px-4 py-2 rounded ${
                                                    selectedSection === section ? 'bg-gray-500' : 'hover:bg-gray-600'
                                                } transition-colors`}
                                            >
                                                {section}
                                            </button>
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </li>
                        
                        <li>
                            <a href="#" className="block px-4 py-2 rounded hover:bg-gray-600 transition-colors">
                                Settings
                            </a>
                        </li>
                    </ul>
                </nav>

                <div className="absolute bottom-0 w-64 border-t border-gray-600">
                    <div className="p-4">
                        <a href="#" className="block px-4 py-2 hover:bg-gray-600 rounded transition-colors">
                            Contact Us
                        </a>
                        <button 
                            onClick={onLogout} 
                            className="block w-full text-left px-4 py-2 hover:bg-gray-600 rounded transition-colors"
                        >
                            Logout
                        </button>
                    </div>
                </div>
            </div>

            <div className="flex-1 overflow-auto">
                <div className="p-8">
                    <div className="flex justify-between items-center mb-8">
                        <h1 className="text-2xl font-semibold">{selectedSection}</h1>
                        <div className="flex items-center space-x-4">
                            <div className="flex items-center bg-gray-100 rounded-md px-3 py-2">
                                <Search size={20} className="text-gray-400 mr-2" />
                                <input
                                    type="text"
                                    placeholder="Search files..."
                                    value={searchQuery}
                                    onChange={handleSearch}
                                    className="bg-transparent outline-none w-64"
                                />
                            </div>
                            
                            <button 
                                onClick={() => setIsUploadModalOpen(true)}
                                className="flex items-center space-x-2 px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
                                aria-label="Upload file"
                            >
                                <Upload size={20} />
                                <span>Upload File</span>
                            </button>
                        </div>
                    </div>

                    <ViewFile 
                        searchQuery={searchQuery}
                        user={user}
                    />
                </div>
            </div>

            <UploadFile
                isOpen={isUploadModalOpen}
                onClose={() => setIsUploadModalOpen(false)}
                onUpload={handleUploadComplete}
                user={user}
            />
        </div>
    );
};

export default EndUserDashboard;