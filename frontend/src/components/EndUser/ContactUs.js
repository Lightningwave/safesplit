import React, { useState } from 'react';

const FeedbackForm = () => {
    const [formData, setFormData] = useState({
        category: 'feature_request',
        subject: '',
        message: ''
    });
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const categories = [
        {
            id: 'feature_request',
            name: 'Feature Request',
            description: 'Suggest a new feature for SafeSplit'
        },
        {
            id: 'bug_report',
            name: 'Bug Report',
            description: 'Report a problem or issue with the application'
        },
        {
            id: 'general_feedback',
            name: 'General Feedback',
            description: 'Share your thoughts about SafeSplit'
        },
        {
            id: 'improvement',
            name: 'Improvement Suggestion',
            description: 'Suggest improvements to existing features'
        },
        {
            id: 'suggestion',
            name: 'Other Suggestion',
            description: 'Any other suggestions for SafeSplit'
        }
    ];

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsSubmitting(true);
        setError('');
        setSuccess('');

        try {
            const token = localStorage.getItem('token');
            if (!token) {
                throw new Error('Please log in to submit feedback');
            }

            const response = await fetch('/api/feedback', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(formData)
            });

            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Failed to submit feedback');
            }

            setSuccess(data.data.message);
            setFormData({
                category: 'feature_request',
                subject: '',
                message: ''
            });
        } catch (err) {
            setError(err.message);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="max-w-xl mx-auto bg-white p-8 rounded-lg shadow-md">
            <h2 className="text-2xl font-semibold mb-6">Submit Feedback</h2>
            
            {error && (
                <div className="mb-4 p-4 bg-red-50 border-l-4 border-red-500 text-red-700">
                    {error}
                </div>
            )}
            
            {success && (
                <div className="mb-4 p-4 bg-green-50 border-l-4 border-green-500 text-green-700">
                    {success}
                </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-6">
                <div>
                    <label className="block text-sm font-medium text-gray-700">
                        Feedback Category
                    </label>
                    <select
                        name="category"
                        value={formData.category}
                        onChange={handleInputChange}
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    >
                        {categories.map(category => (
                            <option key={category.id} value={category.id}>
                                {category.name}
                            </option>
                        ))}
                    </select>
                    <p className="mt-1 text-sm text-gray-500">
                        {categories.find(c => c.id === formData.category)?.description}
                    </p>
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">
                        Subject
                    </label>
                    <input
                        type="text"
                        name="subject"
                        value={formData.subject}
                        onChange={handleInputChange}
                        required
                        minLength={5}
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Brief subject of your feedback"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">
                        Message
                    </label>
                    <textarea
                        name="message"
                        value={formData.message}
                        onChange={handleInputChange}
                        required
                        minLength={10}
                        rows={5}
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Provide details about your feedback..."
                    />
                </div>

                <button
                    type="submit"
                    disabled={isSubmitting}
                    className={`w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white ${
                        isSubmitting 
                            ? 'bg-gray-400 cursor-not-allowed' 
                            : 'bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500'
                    }`}
                >
                    {isSubmitting ? 'Submitting...' : 'Submit Feedback'}
                </button>
            </form>
        </div>
    );
};

export default FeedbackForm;