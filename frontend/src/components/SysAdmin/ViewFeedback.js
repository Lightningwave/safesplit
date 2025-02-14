import React, { useState, useEffect } from 'react';
import { MessageSquare, Check, ClockIcon, RefreshCw } from 'lucide-react';

const ViewFeedback = () => {
    const [feedbacks, setFeedbacks] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [stats, setStats] = useState({
        total: 0,
        pending: 0,
        review: 0,
        resolved: 0
    });
    const [selectedStatus, setSelectedStatus] = useState('all');
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [selectedFeedback, setSelectedFeedback] = useState(null);

    const fetchFeedbacks = async (page, status) => {
        try {
            const token = localStorage.getItem('token');
            let url = `/api/system/feedback?page=${page}&page_size=10`;
            if (status && status !== 'all') {
                url += `&status=${status}`;
            }

            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch feedbacks');
            }

            const data = await response.json();
            setFeedbacks(data.data.feedbacks);
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
            const response = await fetch('/api/system/feedback/stats', {
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

    const updateStatus = async (feedbackId, newStatus) => {
        const updatedFeedback = feedbacks.find(f => f.id === feedbackId);
        const oldStatus = updatedFeedback.status;
        
        setFeedbacks(prevFeedbacks => 
            prevFeedbacks.map(feedback => 
                feedback.id === feedbackId 
                    ? { ...feedback, status: newStatus }
                    : feedback
            )
        );
    
        setStats(prevStats => ({
            ...prevStats,
            [oldStatus]: prevStats[oldStatus] - 1,
            [newStatus]: prevStats[newStatus] + 1
        }));
    
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`/api/system/feedback/${feedbackId}/status`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    status: newStatus
                })
            });
    
            if (!response.ok) {
                throw new Error('Failed to update status');
            }
    
        } catch (err) {
            setFeedbacks(prevFeedbacks => 
                prevFeedbacks.map(feedback => 
                    feedback.id === feedbackId 
                        ? { ...feedback, status: oldStatus }
                        : feedback
                )
            );
    
            setStats(prevStats => ({
                ...prevStats,
                [oldStatus]: prevStats[oldStatus] + 1,
                [newStatus]: prevStats[newStatus] - 1
            }));
    
            setError(err.message);
        }
    };

    useEffect(() => {
        fetchFeedbacks(currentPage, selectedStatus);
        fetchStats();
    }, [currentPage, selectedStatus]);

    const getStatusColor = (status) => {
        switch (status) {
            case 'pending':
                return 'bg-yellow-100 text-yellow-800';
            case 'in_review':
                return 'bg-blue-100 text-blue-800';
            case 'resolved':
                return 'bg-green-100 text-green-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    };

    return (
        <div className="space-y-6">
            {/* Stats Section */}
            <div className="grid grid-cols-4 gap-4">
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">Total Feedback</h3>
                    <p className="text-2xl font-semibold">{stats.total}</p>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">Pending</h3>
                    <p className="text-2xl font-semibold text-yellow-600">{stats.pending}</p>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">In Review</h3>
                    <p className="text-2xl font-semibold text-blue-600">{stats.review}</p>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <h3 className="text-sm text-gray-500">Resolved</h3>
                    <p className="text-2xl font-semibold text-green-600">{stats.resolved}</p>
                </div>
            </div>

            {/* Filters */}
            <div className="flex items-center space-x-4 bg-white p-4 rounded-lg shadow">
                <span className="text-gray-500">Filter by status:</span>
                <select
                    value={selectedStatus}
                    onChange={(e) => {
                        setSelectedStatus(e.target.value);
                        setCurrentPage(1);
                    }}
                    className="border rounded px-3 py-1"
                >
                    <option value="all">All</option>
                    <option value="pending">Pending</option>
                    <option value="in_review">In Review</option>
                    <option value="resolved">Resolved</option>
                </select>
            </div>

            {/* Feedback List */}
            <div className="bg-white rounded-lg shadow overflow-hidden">
                {loading ? (
                    <div className="p-4 text-center">Loading...</div>
                ) : error ? (
                    <div className="p-4 text-red-500 text-center">{error}</div>
                ) : (
                    <div>
                        <table className="min-w-full">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Subject</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                                {feedbacks.map((feedback) => (
                                    <tr key={feedback.id} className="hover:bg-gray-50">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm">
                                                <div className="font-medium text-gray-900">
                                                    {feedback.user?.username || 'Unknown User'}
                                                </div>
                                                <div className="text-gray-500">
                                                    {feedback.user?.email}
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <button 
                                                onClick={() => setSelectedFeedback(feedback)}
                                                className="text-blue-600 hover:text-blue-900"
                                            >
                                                {feedback.subject}
                                            </button>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusColor(feedback.status)}`}>
                                                {feedback.status}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {new Date(feedback.created_at).toLocaleString()}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                            <div className="flex space-x-2">
                                                {feedback.status === 'pending' && (
                                                    <button
                                                        onClick={() => updateStatus(feedback.id, 'in_review')}
                                                        className="text-blue-600 hover:text-blue-900"
                                                    >
                                                        Review
                                                    </button>
                                                )}
                                                {feedback.status !== 'resolved' && (
                                                    <button
                                                        onClick={() => updateStatus(feedback.id, 'resolved')}
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

            {/* Feedback Details Modal */}
            {selectedFeedback && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                    <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
                        <div className="p-6">
                            <div className="flex justify-between items-start">
                                <h3 className="text-lg font-semibold">{selectedFeedback.subject}</h3>
                                <button 
                                    onClick={() => setSelectedFeedback(null)}
                                    className="text-gray-400 hover:text-gray-500"
                                >
                                    ×
                                </button>
                            </div>
                            <div className="mt-4 space-y-4">
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500">Message</h4>
                                    <p className="mt-1">{selectedFeedback.message}</p>
                                </div>
                                {selectedFeedback.details && (
                                    <div>
                                        <h4 className="text-sm font-medium text-gray-500">Additional Details</h4>
                                        <pre className="mt-1 whitespace-pre-wrap bg-gray-50 p-3 rounded">
                                            {selectedFeedback.details}
                                        </pre>
                                    </div>
                                )}
                                <div className="mt-4 pt-4 border-t">
                                    <div className="flex justify-between items-center">
                                        <div className="text-sm text-gray-500">
                                            Submitted on: {new Date(selectedFeedback.created_at).toLocaleString()}
                                        </div>
                                        <div className="flex space-x-3">
                                            {selectedFeedback.status !== 'resolved' && (
                                                <button
                                                    onClick={() => {
                                                        updateStatus(selectedFeedback.id, 'resolved');
                                                        setSelectedFeedback(null);
                                                    }}
                                                    className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
                                                >
                                                    Resolve Feedback
                                                </button>
                                            )}
                                            <button
                                                onClick={() => setSelectedFeedback(null)}
                                                className="px-4 py-2 border rounded-md hover:bg-gray-50"
                                            >
                                                Close
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ViewFeedback;