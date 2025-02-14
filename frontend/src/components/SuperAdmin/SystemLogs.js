import React, { useState, useEffect } from 'react';
import { Search, ArrowUpDown, AlertCircle, Loader2 } from 'lucide-react';

const SystemLogs = () => {
    const [logs, setLogs] = useState([]);
    const [filteredLogs, setFilteredLogs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [filters, setFilters] = useState({
        search: '',
        activity_type: '',
        user_id: ''
    });

    const fetchLogs = async () => {
        try {
            setLoading(true);
            const response = await fetch('http://localhost:8080/api/admin/system-logs', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch system logs');
            }

            const data = await response.json();
            setLogs(data.logs);
            setFilteredLogs(data.logs);
            setError(null);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchLogs();
    }, []);

    useEffect(() => {
        let result = [...logs];

        if (filters.search) {
            const searchTerm = filters.search.toLowerCase();
            result = result.filter(log => 
                log.error_message?.toLowerCase().includes(searchTerm) ||
                log.ip_address?.toLowerCase().includes(searchTerm) ||
                log.user_id?.toString().toLowerCase().includes(searchTerm)
            );
        }

        if (filters.activity_type) {
            result = result.filter(log => 
                log.activity_type === filters.activity_type
            );
        }

        if (filters.user_id) {
            result = result.filter(log => 
                log.user_id?.toString().includes(filters.user_id)
            );
        }

        setFilteredLogs(result);
    }, [filters, logs]);

    const handleFilterChange = (field, value) => {
        setFilters(prev => ({ ...prev, [field]: value }));
    };

    const getActivityTypeStyle = (activityType) => {
        switch (activityType) {
            case 'login':
            case 'logout':
                return 'bg-blue-100 text-blue-800';
            case 'upload':
            case 'download':
                return 'bg-green-100 text-green-800';
            case 'delete':
                return 'bg-red-100 text-red-800';
            case 'share':
                return 'bg-purple-100 text-purple-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    };

    return (
        <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
                <div className="flex gap-4">
                    <input
                        type="text"
                        placeholder="Search logs..."
                        className="flex-1 p-2 border rounded-md"
                        value={filters.search}
                        onChange={(e) => handleFilterChange('search', e.target.value)}
                    />
                    <select
                        className="p-2 border rounded-md"
                        value={filters.activity_type}
                        onChange={(e) => handleFilterChange('activity_type', e.target.value)}
                    >
                        <option value="">All Activity Types</option>
                        <option value="login">Login</option>
                        <option value="logout">Logout</option>
                        <option value="upload">Upload</option>
                        <option value="download">Download</option>
                        <option value="delete">Delete</option>
                        <option value="share">Share</option>
                    </select>
                    <input
                        type="text"
                        placeholder="User ID"
                        className="p-2 border rounded-md"
                        value={filters.user_id}
                        onChange={(e) => handleFilterChange('user_id', e.target.value)}
                    />
                </div>
            </div>

            <div className="overflow-x-auto">
                <table className="w-full">
                    <thead className="bg-gray-50">
                        <tr>
                            <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Timestamp</th>
                            <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Activity Type</th>
                            <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Message</th>
                            <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">IP Address</th>
                            <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">User ID</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                        {filteredLogs.map((log) => (
                            <tr key={log.id} className="hover:bg-gray-50">
                                <td className="px-6 py-4 text-sm text-gray-900">
                                    {new Date(log.created_at).toLocaleString()}
                                </td>
                                <td className="px-6 py-4 text-sm">
                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getActivityTypeStyle(log.activity_type)}`}>
                                        {log.activity_type}
                                    </span>
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-900">
                                  {log.details || log.error_message || '-'}

                                </td>
                                <td className="px-6 py-4 text-sm text-gray-900">
                                    {log.ip_address}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-900">
                                    {log.user_id}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>

                {filteredLogs.length === 0 && !loading && (
                    <div className="text-center py-8 text-gray-500">
                        No logs found matching your criteria
                    </div>
                )}
            </div>

            {loading && (
                <div className="flex items-center justify-center p-8">
                    <Loader2 className="animate-spin mr-2" />
                    <span>Loading logs...</span>
                </div>
            )}

            {error && (
                <div className="p-4 m-4 bg-red-50 border border-red-200 rounded-lg">
                    <div className="flex items-center text-red-600">
                        <AlertCircle className="h-5 w-5 mr-2" />
                        <span>{error}</span>
                    </div>
                </div>
            )}
        </div>
    );
};

export default SystemLogs;