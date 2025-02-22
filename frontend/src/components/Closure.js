import React, { useState, useEffect } from 'react';

const ClosureAlert = () => {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    setIsOpen(true);
  }, []);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg max-w-md w-full p-6 shadow-xl">
        <div className="mb-4">
          <h2 className="text-xl font-bold text-red-600 mb-2">
            Important Notice
          </h2>
          <div className="text-gray-600 space-y-2">
            <p>
              We regret to inform you that Safesplit will be permanently closing on March 20, 2025.
            </p>
            <p>
              Please ensure you have downloaded and secured all your files before this date. Thank you for your understanding and support.
            </p>
          </div>
        </div>
        <div className="flex justify-end">
          <button
            onClick={() => setIsOpen(false)}
            className="px-4 py-2 bg-black text-white rounded hover:bg-gray-800 transition-colors"
          >
            I Understand
          </button>
        </div>
      </div>
    </div>
  );
};

export default ClosureAlert;