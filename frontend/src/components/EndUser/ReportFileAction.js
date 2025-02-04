import React, { useState } from 'react';
import { AlertOctagon } from 'lucide-react';

const ReportFileAction = ({ file }) => {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState('');
    const [showForm, setShowForm] = useState(false);
    const [message, setMessage] = useState('');

    const handleSubmit = async () => {
        if (!message || message.length < 10) {
            setError('Please provide a detailed description (minimum 10 characters)');
            return;
        }

        setIsSubmitting(true);
        setError('');

        try {
            const token = localStorage.getItem('token');
            if (!token) {
                throw new Error('Please log in to report files');
            }

            const response = await fetch(`http://localhost:8080/api/reports/file/${file.id}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    file_id: file.id,
                    message: message
                })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Failed to submit report');
            }

            // Close form and reset state on success
            setShowForm(false);
            setMessage('');
            
            // Show success message (you might want to handle this at a higher level)
            alert('Report submitted successfully. Our team will investigate.');

        } catch (err) {
            setError(err.message);
        } finally {
            setIsSubmitting(false);
        }
    };

    if (showForm) {
        return (
            <div className="p-4 bg-white rounded-md shadow-lg absolute right-0 mt-2 w-80 z-30 border">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold">Report File</h3>
                    <button 
                        onClick={() => setShowForm(false)}
                        className="text-gray-500 hover:text-gray-700"
                    >
                        ×
                    </button>
                </div>

                {error && (
                    <div className="mb-4 p-2 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm">
                        {error}
                    </div>
                )}

                <textarea
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    placeholder="Please describe why you're reporting this file..."
                    className="w-full p-2 border rounded-md mb-4 min-h-[100px]"
                    required
                    minLength={10}
                />

                <div className="flex justify-end space-x-2">
                    <button
                        onClick={() => setShowForm(false)}
                        className="px-3 py-1 text-gray-600 hover:bg-gray-100 rounded-md"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        disabled={isSubmitting}
                        className={`px-3 py-1 rounded-md text-white ${
                            isSubmitting
                                ? 'bg-gray-400 cursor-not-allowed'
                                : 'bg-red-500 hover:bg-red-600'
                        }`}
                    >
                        {isSubmitting ? 'Submitting...' : 'Submit Report'}
                    </button>
                </div>
            </div>
        );
    }

    return (
        <button
            onClick={() => setShowForm(true)}
            className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2 text-red-600"
        >
            <AlertOctagon size={16} />
            <span>Report File</span>
        </button>
    );
};

export default ReportFileAction;