import React from 'react';
import PropTypes from 'prop-types';
import { X } from 'lucide-react';

const ConfirmationAction = ({ isOpen, onClose, onConfirm, message }) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-white rounded-lg shadow-lg w-11/12 md:w-1/3">
        <div className="flex justify-between items-center p-4 border-b">
          <h3 className="text-xl font-semibold">Confirm Action</h3>
          <button onClick={onClose} aria-label="Close modal">
            <X size={24} />
          </button>
        </div>

        <div className="p-4">
          <p>{message}</p>
        </div>

        <div className="flex justify-end p-4 border-t">
          <button 
            onClick={onClose} 
            className="px-4 py-2 bg-gray-300 text-gray-700 rounded mr-2 hover:bg-gray-400"
          >
            Cancel
          </button>
          <button 
            onClick={onConfirm}
            className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );
};

ConfirmationAction.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired,
  message: PropTypes.string.isRequired,
};

export default ConfirmationAction;
