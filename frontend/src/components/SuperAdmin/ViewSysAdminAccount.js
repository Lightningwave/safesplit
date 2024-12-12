import React, { useState, useEffect } from 'react';
import { AlertCircle, CheckCircle2, Loader2, Search } from 'lucide-react';
import DeleteSysAdminAccount from './DeleteSysAdminAccount';

const ViewSysAdminAccount = () => {
  const [sysAdmins, setSysAdmins] = useState([]);
  const [filteredAdmins, setFilteredAdmins] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [successMessage, setSuccessMessage] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchSysAdmins = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/admin/sysadmins', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch system administrators');
      }

      const data = await response.json();
      setSysAdmins(data.sysAdmins);
      setFilteredAdmins(data.sysAdmins);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSysAdmins();
  }, []);

  useEffect(() => {
    if (searchQuery.trim() === '') {
      setFilteredAdmins(sysAdmins);
      return;
    }

    const lowerQuery = searchQuery.toLowerCase();
    const filtered = sysAdmins.filter(admin => 
      admin.username.toLowerCase().includes(lowerQuery) ||
      admin.email.toLowerCase().includes(lowerQuery)
    );
    setFilteredAdmins(filtered);
  }, [searchQuery, sysAdmins]);

  const handleDeleteSuccess = async () => {
    setSuccessMessage('System administrator deleted successfully');
    await fetchSysAdmins();
    setTimeout(() => setSuccessMessage(null), 5000);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="animate-spin mr-2" />
        <span>Loading system administrators...</span>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-xl font-semibold">System Administrators</h2>
            <p className="text-sm text-gray-600 mt-1">
              Manage and monitor system administrator accounts
            </p>
          </div>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search className="h-5 w-5 text-gray-400" />
            </div>
            <input
              type="text"
              placeholder="Search administrators..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 pr-4 py-2 border rounded-md w-72 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        </div>
      </div>

      {error && (
        <div className="m-4 p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center text-red-600">
            <AlertCircle size={20} className="mr-2" />
            <span>{error}</span>
          </div>
        </div>
      )}

      {successMessage && (
        <div className="m-4 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center text-green-600">
            <CheckCircle2 size={20} className="mr-2" />
            <span>{successMessage}</span>
          </div>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Username</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Email</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Status</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Last Login</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredAdmins.map((admin) => (
              <tr key={admin.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 text-sm text-gray-900">{admin.username}</td>
                <td className="px-6 py-4 text-sm text-gray-900">{admin.email}</td>
                <td className="px-6 py-4 text-sm">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                    ${admin.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                    {admin.is_active ? (
                      <><CheckCircle2 size={12} className="mr-1" /> Active</>
                    ) : (
                      <><AlertCircle size={12} className="mr-1" /> Inactive</>
                    )}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm text-gray-900">
                  {admin.last_login ? new Date(admin.last_login).toLocaleString() : 'Never'}
                </td>
                <td className="px-6 py-4 text-sm">
                  <DeleteSysAdminAccount 
                    admin={admin}
                    onSuccess={handleDeleteSuccess}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredAdmins.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            {searchQuery ? 'No administrators found matching your search' : 'No system administrators found'}
          </div>
        )}
      </div>
    </div>
  );
};

export default ViewSysAdminAccount;