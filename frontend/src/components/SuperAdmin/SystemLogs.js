import React, { useState, useEffect } from 'react';
import { Search, ArrowUpDown, AlertCircle, Loader2 } from 'lucide-react';

const SystemLogs = () => {
    const [logs, setLogs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [filters, setFilters] = useState({
        search: '',
        timestamp: '',
        activity_type: '',
        source: '',
        user_id: ''
    });

    const fetchLogs = async () => {
        try {
            setLoading(true);
            const queryParams = new URLSearchParams();
            Object.entries(filters).forEach(([key, value]) => {
                if (value) queryParams.append(key, value);
            });

            const response = await fetch(`/api/admin/system-logs?${queryParams}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch system logs');
            }

            const data = await response.json();
            setLogs(data.logs);
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

    const handleSearch = (e) => {
        e.preventDefault();
        fetchLogs();
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
                <form onSubmit={handleSearch} className="flex gap-4">
                    <input
                        type="text"
                        placeholder="Search logs..."
                        className="flex-1 p-2 border rounded-md"
                        value={filters.search}
                        onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                    />
                    <button
                        type="submit"
                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                    >
                        Search
                    </button>
                </form>
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
                        {logs.map((log) => (
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
                                    {log.error_message || '-'}
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

                {logs.length === 0 && !loading && (
                    <div className="text-center py-8 text-gray-500">
                        No logs found
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