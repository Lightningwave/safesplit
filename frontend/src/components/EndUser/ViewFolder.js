import React, { useState, useEffect } from 'react';
import { Folder } from 'lucide-react';

const ViewFolder = ({ currentFolderId = null, onFolderClick, selectedSection }) => {
  const [folders, setFolders] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  const fetchFolders = async () => {
    try {
      setLoading(true);
      const endpoint = currentFolderId 
        ? `/api/folders/${currentFolderId}`
        : '/api/folders';

      const response = await fetch(endpoint, {
        credentials: 'include', // This is important for cookie-based auth
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      // Handle different response structures
      let folderData;
      if (currentFolderId) {
        folderData = data.data.folder.sub_folders || [];
      } else {
        folderData = data.data.folders || [];
      }

      setFolders(folderData);
      setError(null);
    } catch (err) {
      console.error('Fetch error:', err);
      setError('Failed to load folders. Please try again later.');
      setFolders([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFolders();
  }, [currentFolderId, selectedSection]);

  if (loading) return (
    <div className="p-4 flex items-center space-x-2">
      <div className="animate-spin rounded-full h-4 w-4 border-2 border-gray-500 border-t-transparent"></div>
      <span>Loading folders...</span>
    </div>
  );

  if (error) return (
    <div className="p-4 mb-4 text-red-700 bg-red-100 rounded-md flex items-center justify-between">
      <div>{error}</div>
      <button 
        onClick={fetchFolders}
        className="px-3 py-1 bg-red-700 text-white rounded-md text-sm hover:bg-red-800"
      >
        Retry
      </button>
    </div>
  );

  if (folders.length === 0) {
    return (
      <div className="p-4 text-gray-500 bg-gray-50 rounded-md">
        No folders found. Create a new folder to get started.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-8">
      {folders.map((folder) => (
        <div
          key={folder.id}
          onClick={() => onFolderClick(folder.id)}
          className="p-4 border rounded-lg hover:bg-gray-50 cursor-pointer transition-colors flex items-center space-x-3"
        >
          <Folder className="text-gray-400" size={24} />
          <span className="font-medium">{folder.name}</span>
        </div>
      ))}
    </div>
  );
};

export default ViewFolder;