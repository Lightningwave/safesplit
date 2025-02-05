import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { X, CreditCard, Mail, Calendar, HardDrive } from 'lucide-react';

const ViewUserAction = ({ isOpen, onClose, userId }) => {
  const [userDetails, setUserDetails] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchUserDetails = async () => {
    if (!userId) return;

    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/api/system/users/${userId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => null);
        if (errorData?.error) {
          throw new Error(errorData.error);
        }
        throw new Error(`Unable to retrieve user information (Status: ${response.status}). Please try again later or contact support if the issue persists.`);
      }

      const data = await response.json();
      if (!data.user_details) {
        throw new Error('The server response was invalid. Please try again or contact support.');
      }
      setUserDetails(data.user_details);
    } catch (err) {
      const errorMessage = err.message || 'An unexpected error occurred while fetching user details. Please try again later.';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchUserDetails();
    }
    return () => {
      setUserDetails(null);
      setError(null);
      setLoading(false);
    };
  }, [isOpen, userId]);

  if (!isOpen) return null;

  const getStatusBadgeClass = (status) => {
    switch (status?.toLowerCase()) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'cancelled':
        return 'bg-red-100 text-red-800';
      case 'free':
        return 'bg-blue-100 text-blue-800';
      case 'premium':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatStorageSize = (bytes) => {
    if (!bytes) return '0 GB';
    const gb = bytes / (1024 * 1024 * 1024);
    return `${gb.toFixed(2)} GB`;
  };

  const renderSubscriptionDetails = () => {
    if (!userDetails?.subscription) return null;
    
    return (
      <div className="mt-6 p-4 bg-gray-50 rounded-lg">
        <h4 className="text-lg font-semibold mb-4">Subscription Information</h4>
        
        <div className="space-y-4">
          <div className="flex items-center space-x-2">
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              getStatusBadgeClass(userDetails.subscription.status)
            }`}>
              {userDetails.subscription.status}
            </span>
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              getStatusBadgeClass(userDetails.subscription.billing_status)
            }`}>
              {userDetails.subscription.billing_status}
            </span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-gray-500 mb-1">Billing Name</p>
              <p className="font-medium">{userDetails.subscription.billing_name || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 mb-1">Billing Cycle</p>
              <p className="font-medium">{userDetails.subscription.billing_cycle || 'N/A'}</p>
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center">
              <CreditCard className="w-4 h-4 mr-2 text-gray-500" />
              <div>
                <p className="text-sm text-gray-500">Payment Method</p>
                <p className="font-medium">{userDetails.subscription.payment_method || 'None'}</p>
              </div>
            </div>

            <div className="flex items-center">
              <Mail className="w-4 h-4 mr-2 text-gray-500" />
              <div>
                <p className="text-sm text-gray-500">Billing Email</p>
                <p className="font-medium">{userDetails.subscription.billing_email || 'N/A'}</p>
              </div>
            </div>

            <div className="flex items-center">
              <Calendar className="w-4 h-4 mr-2 text-gray-500" />
              <div>
                <p className="text-sm text-gray-500">Next Invoice Date</p>
                <p className="font-medium">{userDetails.subscription.next_invoice_date || 'N/A'}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  };

  const renderStorageDetails = () => {
    if (!userDetails?.storage) return null;

    const usagePercentage = Math.round(
      (userDetails.storage.quota_used / userDetails.storage.quota_total) * 100
    );

    return (
      <div className="mt-6 p-4 bg-gray-50 rounded-lg">
        <h4 className="text-lg font-semibold mb-4">Storage Information</h4>
        
        <div className="space-y-4">
          <div className="flex items-center">
            <HardDrive className="w-4 h-4 mr-2 text-gray-500" />
            <div>
              <p className="text-sm text-gray-500">Storage Usage</p>
              <p className="font-medium">
                {formatStorageSize(userDetails.storage.quota_used)} / {formatStorageSize(userDetails.storage.quota_total)}
              </p>
            </div>
          </div>

          <div>
            <div className="flex justify-between text-sm text-gray-500 mb-1">
              <span>Usage</span>
              <span>{usagePercentage}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className="bg-blue-600 rounded-full h-2"
                style={{ width: `${Math.min(usagePercentage, 100)}%` }}
              />
            </div>
          </div>
        </div>
      </div>
    );
  };

  const renderAccessDetails = () => {
    return (
      <div className="mt-6 p-4 bg-gray-50 rounded-lg">
        <h4 className="text-lg font-semibold mb-4">Access Permissions</h4>
        <div className="flex space-x-4">
          <span className={`px-2 py-1 rounded ${userDetails?.read_access ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
            Read: {userDetails?.read_access ? 'Enabled' : 'Disabled'}
          </span>
          <span className={`px-2 py-1 rounded ${userDetails?.write_access ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
            Write: {userDetails?.write_access ? 'Enabled' : 'Disabled'}
          </span>
          <span className={`px-2 py-1 rounded ${userDetails?.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
            Account: {userDetails?.is_active ? 'Active' : 'Inactive'}
          </span>
        </div>
      </div>
    );
  };

  const renderError = () => {
    if (!error) return null;
    
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">Error Loading User Details</h3>
            <div className="mt-2 text-sm text-red-700">
              <p>{error}</p>
              <p className="mt-1">Please verify your connection and permissions, then try again.</p>
            </div>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-white rounded-lg shadow-lg w-11/12 md:w-2/3 lg:w-1/2 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center p-4 border-b">
          <h3 className="text-xl font-semibold">User Account Details</h3>
          <button onClick={onClose} className="hover:bg-gray-100 p-2 rounded-full" aria-label="Close modal">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="p-6">
          {loading && (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            </div>
          )}

          {renderError()}

          {userDetails && (
            <div className="space-y-6">
              {/* Basic Information */}
              <div>
                <h4 className="text-lg font-semibold mb-4">Basic Information</h4>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-gray-500">User ID</p>
                    <p className="font-medium">{userDetails.user_id}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-500">Account Type</p>
                    <p className="font-medium">{userDetails.account_type}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-500">Username</p>
                    <p className="font-medium">{userDetails.username}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-500">Email</p>
                    <p className="font-medium">{userDetails.email}</p>
                  </div>
                </div>
              </div>

              {renderAccessDetails()}
              {renderSubscriptionDetails()}
              {renderStorageDetails()}
            </div>
          )}
        </div>

        <div className="flex justify-end p-4 border-t">
          <button 
            onClick={onClose} 
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

ViewUserAction.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  userId: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
};

ViewUserAction.defaultProps = {
  userId: null,
};

export default ViewUserAction;