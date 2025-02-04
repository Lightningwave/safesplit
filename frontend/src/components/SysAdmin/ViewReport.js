import React, { useState, useEffect } from 'react';
import { AlertTriangle, Search, AlertCircle } from 'lucide-react';

const ViewReport = () => {
    const [reports, setReports] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [stats, setStats] = useState({
        total: {
            all: 0,
            today: 0,
            pending: 0,
            review: 0,
            resolved: 0
        }
    });
    const [selectedStatus, setSelectedStatus] = useState('all');
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [selectedReport, setSelectedReport] = useState(null);

    const fetchReports = async (page, status) => {
        try {
            const token = localStorage.getItem('token');
            let url = `http://localhost:8080/api/system/reports?page=${page}&page_size=10`;
            if (status && status !== 'all') {
                url += `&status=${status}`;
            }

            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch reports');
            }

            const data = await response.json();
            setReports(data.data.reports);
            setTotalPages(data.data.meta.total_pages);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const fetchStats = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/system/reports/stats', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch statistics');
            }

            const data = await response.json();
            setStats(data.data);
        } catch (err) {
            console.error('Failed to fetch stats:', err);
        }
    };

    const updateReportStatus = async (reportId, newStatus) => {
        try {
            const comment = prompt('Please enter a comment for this status update:');
            if (!comment) return;

            const token = localStorage.getItem('token');
            const response = await fetch(`http://localhost:8080/api/system/reports/${reportId}/status`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    status: newStatus,
                    comment: comment
                })
            });

            if (!response.ok) {
                throw new Error('Failed to update status');
            }

            // Refresh data
            fetchReports(currentPage, selectedStatus);
            fetchStats();
        } catch (err) {
            setError(err.message);
        }
    };

    useEffect(() => {
        fetchReports(currentPage, selectedStatus);
        fetchStats();
    }, [currentPage, selectedStatus]);

    const getStatusColor = (status) => {
        switch (status) {
            case 'pending':
                return 'bg-red-100 text-red-800';
            case 'in_review':
                return 'bg-yellow-100 text-yellow-800';
            case 'resolved':
                return 'bg-green-100 text-green-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    };

    return (
        <div className="space-y-6">
            {/* Stats Section */}
            <div className="grid grid-cols-5 gap-4">
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">Total Reports</h3>
                    <p className="text-2xl font-semibold">{stats.total.all}</p>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">Today's Reports</h3>
                    <p className="text-2xl font-semibold text-blue-600">{stats.total.today}</p>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">Pending</h3>
                    <p className="text-2xl font-semibold text-red-600">{stats.total.pending}</p>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">In Review</h3>
                    <p className="text-2xl font-semibold text-yellow-600">{stats.total.review}</p>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">Resolved</h3>
                    <p className="text-2xl font-semibold text-green-600">{stats.total.resolved}</p>
                </div>
            </div>

            {/* Filters */}
            <div className="flex items-center justify-between bg-white p-4 rounded-lg shadow">
                <div className="flex items-center space-x-4">
                    <span className="text-gray-500">Filter by status:</span>
                    <select
                        value={selectedStatus}
                        onChange={(e) => {
                            setSelectedStatus(e.target.value);
                            setCurrentPage(1);
                        }}
                        className="border rounded px-3 py-1"
                    >
                        <option value="all">All Reports</option>
                        <option value="pending">Pending</option>
                        <option value="in_review">In Review</option>
                        <option value="resolved">Resolved</option>
                    </select>
                </div>
                <div className="flex items-center space-x-2">
                    <AlertCircle className="text-red-500" size={20} />
                    <span className="text-red-500 font-medium">
                        {stats.total.pending} reports need attention
                    </span>
                </div>
            </div>

            {/* Reports List */}
            <div className="bg-white rounded-lg shadow overflow-hidden">
                {loading ? (
                    <div className="p-4 text-center">Loading...</div>
                ) : error ? (
                    <div className="p-4 text-red-500 text-center">{error}</div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="min-w-full">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Reporter</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Subject</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-200">
                                {reports.map((report) => (
                                    <tr key={report.id} className="hover:bg-gray-50">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm">
                                                <div className="font-medium text-gray-900">
                                                    {report.reporter.username}
                                                </div>
                                                <div className="text-gray-500">
                                                    {report.reporter.email}
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <button 
                                                onClick={() => setSelectedReport(report)}
                                                className="text-blue-600 hover:text-blue-900"
                                            >
                                                {report.subject}
                                            </button>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusColor(report.status)}`}>
                                                {report.status}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {new Date(report.created_at).toLocaleString()}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                            <div className="flex space-x-2">
                                                {report.status === 'pending' && (
                                                    <button
                                                        onClick={() => updateReportStatus(report.id, 'in_review')}
                                                        className="text-yellow-600 hover:text-yellow-900"
                                                    >
                                                        Start Review
                                                    </button>
                                                )}
                                                {report.status !== 'resolved' && (
                                                    <button
                                                        onClick={() => updateReportStatus(report.id, 'resolved')}
                                                        className="text-green-600 hover:text-green-900"
                                                    >
                                                        Resolve
                                                    </button>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>

                        {/* Pagination */}
                        <div className="px-6 py-4 flex items-center justify-between border-t">
                            <button
                                onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                                disabled={currentPage === 1}
                                className="px-3 py-1 border rounded disabled:opacity-50"
                            >
                                Previous
                            </button>
                            <span>Page {currentPage} of {totalPages}</span>
                            <button
                                onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                                disabled={currentPage === totalPages}
                                className="px-3 py-1 border rounded disabled:opacity-50"
                            >
                                Next
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Report Details Modal */}
            {selectedReport && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
                    <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
                        <div className="p-6">
                            <div className="flex justify-between items-start">
                                <h3 className="text-lg font-semibold">{selectedReport.subject}</h3>
                                <button 
                                    onClick={() => setSelectedReport(null)}
                                    className="text-gray-400 hover:text-gray-500"
                                >
                                    ×
                                </button>
                            </div>
                            <div className="mt-4 space-y-4">
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500">Message</h4>
                                    <p className="mt-1">{selectedReport.message}</p>
                                </div>
                                {selectedReport.details && (
                                    <div>
                                        <h4 className="text-sm font-medium text-gray-500">History</h4>
                                        <pre className="mt-1 whitespace-pre-wrap bg-gray-50 p-3 rounded">
                                            {selectedReport.details}
                                        </pre>
                                    </div>
                                )}
                                <div className="flex justify-end space-x-3 pt-4 border-t">
                                    <button
                                        onClick={() => {
                                            setSelectedReport(null);
                                        }}
                                        className="px-4 py-2 border rounded-md hover:bg-gray-50"
                                    >
                                        Close
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ViewReport;