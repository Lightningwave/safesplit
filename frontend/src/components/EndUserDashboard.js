import React, { useState, useEffect } from 'react';
import { Settings as SettingsIcon, ChevronDown, ChevronRight, Search, Upload, Folder, ChevronLeft, MoreVertical } from 'lucide-react';
import UploadFile from './EndUser/UploadFile';
import ViewFile from './EndUser/ViewFile';
import Settings from './EndUser/Settings';
import ContactUs from './EndUser/ContactUs';
import CreateFolder from './EndUser/CreateFolder';
import DeleteFolder from './EndUser/DeleteFolder';

const EndUserDashboard = ({ user, onLogout }) => {
    const [isFilesOpen, setIsFilesOpen] = useState(true);
    const [selectedSection, setSelectedSection] = useState('Dashboard');
    const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [folders, setFolders] = useState([]);
    const [currentFolder, setCurrentFolder] = useState(null);
    const [folderPath, setFolderPath] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [isCreateFolderOpen, setIsCreateFolderOpen] = useState(false);
    const [isDeleteFolderOpen, setIsDeleteFolderOpen] = useState(false);
    const [folderToDelete, setFolderToDelete] = useState(null);

    useEffect(() => {
        if (selectedSection === 'Dashboard') {
            fetchFolders();
        }
    }, [selectedSection]);

    const fetchFolders = async (parentId = null) => {
        setIsLoading(true);
        try {
            const token = localStorage.getItem('token');
            let url = 'http://localhost:8080/api/folders';
            if (parentId) {
                url += `?parent_id=${parentId}`;
            }

            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });
            
            if (response.status === 401) {
                onLogout();
                return;
            }
            
            if (!response.ok) {
                throw new Error('Failed to fetch folders');
            }
            
            const result = await response.json();
            if (result.status === 'success') {
                setFolders(result.data.folders || []);
            } else {
                throw new Error(result.error);
            }
        } catch (error) {
            console.error('Error fetching folders:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const fetchFolderContents = async (folderId) => {
        setIsLoading(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`http://localhost:8080/api/folders/${folderId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });
            
            if (response.status === 401) {
                onLogout();
                return;
            }

            if (!response.ok) {
                throw new Error('Failed to fetch folder contents');
            }
            
            const result = await response.json();
            if (result.status === 'success') {
                setCurrentFolder(result.data.folder);
                setFolderPath(result.data.path || []);
                // Fetch folders that might be inside this folder
                fetchFolders(folderId);
            } else {
                throw new Error(result.error);
            }
        } catch (error) {
            console.error('Error fetching folder contents:', error);
        }
    };

    const handleSearch = (event) => {
        setSearchQuery(event.target.value);
    };

    const handleUploadComplete = () => {
        if (currentFolder) {
            fetchFolderContents(currentFolder.id);
        } else {
            fetchFolders();
        }
    };

    const handleFolderClick = (folder) => {
        fetchFolderContents(folder.id);
    };

    const handleFolderDelete = (folder) => {
        setFolderToDelete(folder);
        setIsDeleteFolderOpen(true);
    };

    const handleBackClick = () => {
        if (folderPath.length > 1) {
            const parentFolder = folderPath[folderPath.length - 2];
            fetchFolderContents(parentFolder.id);
        } else {
            setCurrentFolder(null);
            setFolderPath([]);
            fetchFolders();
        }
    };

    const renderFolders = () => {
        if (isLoading) {
            return <div className="text-gray-600">Loading...</div>;
        }

        return (
            <div>
                {currentFolder && (
                    <div className="mb-4">
                        <button
                            onClick={handleBackClick}
                            className="flex items-center text-gray-600 hover:text-gray-900"
                        >
                            <ChevronLeft size={20} className="mr-1" />
                            Back
                        </button>
                    </div>
                )}

                {currentFolder && renderBreadcrumbs()}

                <div className="grid grid-cols-4 gap-4">
                    {folders.map((folder) => (
                        <div
                            key={folder.id}
                            className="group relative"
                        >
                            <button
                                onClick={() => handleFolderClick(folder)}
                                className="w-full flex items-center p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                            >
                                <Folder className="text-gray-400 mr-3" size={24} />
                                <span className="text-gray-700 font-medium">{folder.name}</span>
                            </button>
                            <button
                                onClick={() => handleFolderDelete(folder)}
                                className="absolute right-2 top-2 p-2 opacity-0 group-hover:opacity-100 transition-opacity text-gray-500 hover:text-gray-700"
                            >
                                <MoreVertical size={16} />
                            </button>
                        </div>
                    ))}
                </div>
            </div>
        );
    };

    const renderBreadcrumbs = () => {
        if (!currentFolder) return null;

        return (
            <div className="flex items-center space-x-2 text-sm text-gray-600 mb-4">
                <button
                    onClick={() => {
                        setCurrentFolder(null);
                        setFolderPath([]);
                        fetchFolders();
                    }}
                    className="hover:text-gray-900"
                >
                    Dashboard
                </button>
                {folderPath.map((folder, index) => (
                    <React.Fragment key={folder.id}>
                        <ChevronRight size={16} className="text-gray-400" />
                        <button
                            onClick={() => fetchFolderContents(folder.id)}
                            className="hover:text-gray-900"
                        >
                            {folder.name}
                        </button>
                    </React.Fragment>
                ))}
            </div>
        );
    };

    const renderDashboard = () => {
        return (
            <div className="space-y-8">
                <div>
                    <div className="flex justify-between items-center mb-4">
                        <h2 className="text-xl font-semibold">Folders</h2>
                        <button 
                            onClick={() => setIsCreateFolderOpen(true)}
                            className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
                        >
                            New Folder
                        </button>
                    </div>
                    {renderFolders()}
                </div>

                <div>
                    <h2 className="text-xl font-semibold mb-4">Recents</h2>
                    <ViewFile 
                        searchQuery=""
                        user={user}
                        selectedSection={selectedSection}
                        currentFolder={currentFolder}
                        showRecentsOnly={true}
                        maxItems={5}
                    />
                </div>
            </div>
        );
    };

    return (
        <div className="flex h-screen bg-white">
            <div className="w-64 bg-gray-700 text-white relative">
                <div className="p-1 border-b border-gray-600 flex justify-center items-center">
                    <img src="/safesplit-logo_white.png" alt="SafeSplit Logo" className="h-32 w-auto" />
                </div>
                
                <nav className="p-4">
                    <ul className="space-y-2">
                        <li>
                            <button
                                onClick={() => {
                                    setSelectedSection('Dashboard');
                                    setCurrentFolder(null);
                                    setFolderPath([]);
                                    fetchFolders();
                                }}
                                className={`block w-full text-left px-4 py-2 rounded ${
                                    selectedSection === 'Dashboard' ? 'bg-gray-600' : 'hover:bg-gray-600'
                                } transition-colors`}
                            >
                                Dashboard
                            </button>
                        </li>
                        
                        <li>
                            <button 
                                onClick={() => setIsFilesOpen(!isFilesOpen)}
                                className="w-full flex items-center justify-between px-4 py-2 rounded hover:bg-gray-600 transition-colors"
                            >
                                <span>Files</span>
                                {isFilesOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                            </button>
                            
                            {isFilesOpen && (
                                <ul className="ml-4 mt-2 space-y-1">
                                    {['Uploaded Files', 'Shared Files', 'Archives'].map((section) => (
                                        <li key={section}>
                                            <button 
                                                onClick={() => {
                                                    setSelectedSection(section);
                                                    setCurrentFolder(null);
                                                    setFolderPath([]);
                                                }}
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
                            <button 
                                onClick={() => setSelectedSection('Settings')}
                                className={`block w-full text-left px-4 py-2 rounded ${
                                    selectedSection === 'Settings' ? 'bg-gray-600' : 'hover:bg-gray-600'
                                } transition-colors`}
                            >
                                Settings
                            </button>
                        </li>
                    </ul>
                </nav>

                <div className="absolute bottom-0 w-64 border-t border-gray-600">
                    <div className="p-4">
                        <button
                            onClick={() => setSelectedSection('Contact Us')}
                            className={`block w-full text-left px-4 py-2 hover:bg-gray-600 rounded transition-colors ${
                                selectedSection === 'Contact Us' ? 'bg-gray-500' : ''
                            }`}
                        >
                            Contact Us
                        </button>
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
                        {selectedSection !== 'Settings' && selectedSection !== 'Contact Us' && (
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
                        )}
                    </div>

                    {selectedSection === 'Settings' && <Settings user={user} />}
                    {selectedSection === 'Contact Us' && <ContactUs onSubmit={(formData) => console.log("Form Submitted:", formData)} />}
                    {selectedSection === 'Dashboard' && renderDashboard()}
                    {selectedSection !== 'Settings' && 
                     selectedSection !== 'Contact Us' && 
                     selectedSection !== 'Dashboard' && (
                        <ViewFile 
                            searchQuery={searchQuery}
                            user={user}
                            selectedSection={selectedSection}
                            currentFolder={currentFolder}
                        />
                    )}
                </div>
            </div>

            <UploadFile
                isOpen={isUploadModalOpen}
                onClose={() => setIsUploadModalOpen(false)}
                onUpload={handleUploadComplete}
                user={user}
                currentFolder={currentFolder}
            />

            <CreateFolder
                isOpen={isCreateFolderOpen}
                onClose={() => setIsCreateFolderOpen(false)}
                currentFolder={currentFolder}
                onFolderCreated={() => {
                    if (currentFolder) {
                        fetchFolderContents(currentFolder.id);
                    } else {
                        fetchFolders();
                    }
                }}
            />

            <DeleteFolder
                isOpen={isDeleteFolderOpen}
                onClose={() => {
                    setIsDeleteFolderOpen(false);
                    setFolderToDelete(null);
                }}
                folder={folderToDelete}
                onFolderDeleted={() => {
                    if (currentFolder) {
                        fetchFolderContents(currentFolder.id);
                    } else {
                        fetchFolders();
                    }
                }}
            />
        </div>
    );
};

export default EndUserDashboard;