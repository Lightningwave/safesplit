import React, { useState, useEffect } from 'react';
import { Settings, ChevronDown, ChevronRight, Search, Upload, Folder, ChevronLeft, MoreVertical } from 'lucide-react';
import UploadFile from './EndUser/UploadFile';
import ViewFile from './EndUser/ViewFile';
import ViewFolder from './EndUser/ViewFolder';
import SettingsPage from './EndUser/Settings';
import ContactUs from './EndUser/ContactUs';
import CreateFolder from './EndUser/CreateFolder';
import DeleteFolder from './EndUser/DeleteFolder';
import TrashBin from './PremiumUser/TrashBin';  

const PremiumUserDashboard = ({ user, onLogout }) => {
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
    const [error, setError] = useState(null);
    const [refreshTrigger, setRefreshTrigger] = useState(0);

    useEffect(() => {
        if (selectedSection === 'Dashboard') {
            refreshCurrentView();
        }
    }, [selectedSection]);

    const refreshCurrentView = async () => {
        if (currentFolder) {
            await fetchFolderContents(currentFolder.id);
        } else {
            await fetchFolders();
        }
    };

    const triggerRefresh = () => {
        setRefreshTrigger(prev => prev + 1);
    };

    const fetchFolders = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/folders', {
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
            setError('Failed to load folders. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    const fetchFolderContents = async (folderId) => {
        setIsLoading(true);
        setError(null);
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
                setFolders(result.data.folder.sub_folders || []);
            } else {
                throw new Error(result.error);
            }
        } catch (error) {
            console.error('Error fetching folder contents:', error);
            setError('Failed to load folder contents. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    const handleSearch = (event) => {
        setSearchQuery(event.target.value);
    };

    const handleUploadComplete = async () => {
        triggerRefresh();
        await refreshCurrentView();
    };

    const handleFolderClick = async (folder) => {
        try {
            await fetchFolderContents(folder.id);
        } catch (error) {
            setError('Failed to open folder. Please try again.');
        }
    };

    const handleFolderDelete = (folder) => {
        setFolderToDelete(folder);
        setIsDeleteFolderOpen(true);
    };

    const handleBackClick = async () => {
        try {
            if (folderPath.length > 1) {
                const parentFolder = folderPath[folderPath.length - 2];
                await fetchFolderContents(parentFolder.id);
            } else {
                setCurrentFolder(null);
                setFolderPath([]);
                await fetchFolders();
            }
        } catch (error) {
            setError('Failed to navigate back. Please try again.');
        }
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
                    Root
                </button>
                {folderPath.map((folder, index) => (
                    <React.Fragment key={folder.id}>
                        <ChevronRight size={16} className="text-gray-400" />
                        <button
                            onClick={() => fetchFolderContents(folder.id)}
                            className={`hover:text-gray-900 ${
                                index === folderPath.length - 1 ? 'font-medium' : ''
                            }`}
                        >
                            {folder.name}
                        </button>
                    </React.Fragment>
                ))}
            </div>
        );
    };

    const renderDashboard = () => (
        <div className="space-y-8">
            <div>
                <div className="flex justify-between items-center mb-4">
                    <div className="flex items-center space-x-2">
                        <h2 className="text-xl font-semibold">
                            {currentFolder ? currentFolder.name : 'Folders'}
                        </h2>
                        <span className="text-sm text-gray-500">
                            ({folders.length} {folders.length === 1 ? 'folder' : 'folders'})
                        </span>
                    </div>
                    <button 
                        onClick={() => setIsCreateFolderOpen(true)}
                        className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
                    >
                        New Folder
                    </button>
                </div>
                <ViewFolder 
                    currentFolder={currentFolder}
                    onFolderClick={handleFolderClick}
                    onFolderDelete={handleFolderDelete}
                    onBackClick={handleBackClick}
                    selectedSection={selectedSection}
                    refreshTrigger={refreshTrigger}
                />
            </div>

            <div>
                <h2 className="text-xl font-semibold mb-4">Recent</h2>
                <ViewFile 
                    searchQuery={searchQuery}
                    user={user}
                    selectedSection={selectedSection}
                    currentFolder={currentFolder}
                    showRecentsOnly={true} 
                    refreshTrigger={refreshTrigger}
                />
            </div>
        </div>
    );

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
                                onClick={() => {
                                    setSelectedSection('Trash Bin');
                                    setCurrentFolder(null);
                                    setFolderPath([]);
                                }}
                                className={`block w-full text-left px-4 py-2 rounded ${
                                    selectedSection === 'Trash Bin' ? 'bg-gray-600' : 'hover:bg-gray-600'
                                } transition-colors`}
                            >
                                Trash Bin
                            </button>
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

                    {selectedSection === 'Settings' && <SettingsPage user={user} />}
                    {selectedSection === 'Contact Us' && <ContactUs onSubmit={(formData) => console.log("Form Submitted:", formData)} />}
                    {selectedSection === 'Dashboard' && renderDashboard()}
                    {['Uploaded Files', 'Shared Files', 'Archives'].includes(selectedSection) && (
                    <ViewFile 
                        searchQuery={searchQuery}
                        user={user}
                        selectedSection={selectedSection}
                        currentFolder={currentFolder}
                    />
                )}
                {selectedSection === 'Trash Bin' && (
                    <TrashBin user={user} />
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

            {error && (
                <div className="fixed bottom-4 right-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                    <span>{error}</span>
                    <button
                        onClick={() => setError(null)}
                        className="ml-2 font-bold"
                    >
                        ×
                    </button>
                </div>
            )}

            {isLoading && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
                    <div className="bg-white p-4 rounded-lg">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default PremiumUserDashboard;