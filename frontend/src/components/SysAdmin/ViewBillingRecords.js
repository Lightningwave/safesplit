import React, { useState, useEffect } from 'react';
import { CreditCard, DollarSign, Calendar, AlertCircle, Clock } from 'lucide-react';

const ViewBillingRecords = () => {
    const [records, setRecords] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [stats, setStats] = useState({
        total_subscriptions: 0,
        active_subscriptions: 0,
        monthly_subscriptions: 0,
        yearly_subscriptions: 0
    });
    const [selectedStatus, setSelectedStatus] = useState('all');
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [expiringSubscriptions, setExpiringSubscriptions] = useState([]);

    const fetchBillingRecords = async (page, status) => {
        try {
            const token = localStorage.getItem('token');
            let url = `http://localhost:8080/api/system/billing/records?page=${page}&page_size=10`;
            if (status && status !== 'all') {
                url += `&status=${status}`;
            }

            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch billing records');
            }

            const data = await response.json();
            setRecords(data.data.records);
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
            const response = await fetch('http://localhost:8080/api/system/billing/stats', {
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

    const fetchExpiringSubscriptions = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/system/billing/expiring', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch expiring subscriptions');
            }

            const data = await response.json();
            setExpiringSubscriptions(data.data);
        } catch (err) {
            console.error('Failed to fetch expiring subscriptions:', err);
        }
    };

    useEffect(() => {
        fetchBillingRecords(currentPage, selectedStatus);
        fetchStats();
        fetchExpiringSubscriptions();
    }, [currentPage, selectedStatus]);

    return (
        <div className="space-y-6">
            {/* Stats Section */}
            <div className="grid grid-cols-4 gap-4">
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <h3 className="text-sm text-gray-500">Total Subscriptions</h3>
                            <p className="text-2xl font-semibold">{stats.total_subscriptions}</p>
                        </div>
                        <DollarSign className="text-blue-500" size={24} />
                    </div>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <h3 className="text-sm text-gray-500">Active Subscriptions</h3>
                            <p className="text-2xl font-semibold text-green-600">{stats.active_subscriptions}</p>
                        </div>
                        <CreditCard className="text-green-500" size={24} />
                    </div>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <h3 className="text-sm text-gray-500">Monthly Plans</h3>
                            <p className="text-2xl font-semibold text-purple-600">{stats.monthly_subscriptions}</p>
                        </div>
                        <Calendar className="text-purple-500" size={24} />
                    </div>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                    <div className="flex items-center justify-between">
                        <div>
                            <h3 className="text-sm text-gray-500">Yearly Plans</h3>
                            <p className="text-2xl font-semibold text-orange-600">{stats.yearly_subscriptions}</p>
                        </div>
                        <Calendar className="text-orange-500" size={24} />
                    </div>
                </div>
            </div>

            {/* Expiring Subscriptions Alert */}
            {expiringSubscriptions.length > 0 && (
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                    <div className="flex items-center space-x-2">
                        <Clock className="text-yellow-500" size={20} />
                        <span className="text-yellow-700">
                            {expiringSubscriptions.length} subscriptions expiring in the next 7 days
                        </span>
                    </div>
                </div>
            )}

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
                    <option value="active">Active</option>
                    <option value="pending">Pending</option>
                    <option value="cancelled">Cancelled</option>
                </select>
            </div>

            {/* Billing Records Table */}
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
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Plan</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Next Billing</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                                {records.map((record) => (
                                    <tr key={record.id}>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm">
                                                <div className="font-medium text-gray-900">{record.billing_name}</div>
                                                <div className="text-gray-500">{record.billing_email}</div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className="text-sm text-gray-900">
                                                {record.billing_cycle.charAt(0).toUpperCase() + record.billing_cycle.slice(1)}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full 
                                                ${record.billing_status === 'active' ? 'bg-green-100 text-green-800' : 
                                                record.billing_status === 'pending' ? 'bg-yellow-100 text-yellow-800' : 
                                                'bg-red-100 text-red-800'}`}>
                                                {record.billing_status.charAt(0).toUpperCase() + record.billing_status.slice(1)}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {new Date(record.next_billing_date).toLocaleDateString()}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                            {record.currency} {record.billing_cycle === 'monthly' ? '9.99' : '99.99'}
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
        </div>
    );
};

export default ViewBillingRecords;