import React, { useState } from 'react';

const ContactUs = ({ onSubmit }) => {
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [messageType, setMessageType] = useState('Feedback');
    const [message, setMessage] = useState('');
    const [attachment, setAttachment] = useState(null);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleFileChange = (e) => {
        setAttachment(e.target.files[0]);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();

        // Set submitting state
        setIsSubmitting(true);

        // Prepare form data for submission
        const formData = new FormData();
        formData.append('name', name);
        formData.append('email', email);
        formData.append('messageType', messageType);
        formData.append('message', message);
        if (attachment) {
            formData.append('attachment', attachment);
        }

        // If using an onSubmit callback passed in as a prop:
        // await onSubmit(formData);

        // For demonstration, just log the data
        console.log('Contact Us Submission:', { name, email, messageType, message, attachment });

        // Reset the form or show a success message
        setName('');
        setEmail('');
        setMessageType('Feedback');
        setMessage('');
        setAttachment(null);
        setIsSubmitting(false);
    };

    return (
        <div className="max-w-xl mx-auto bg-white p-8 rounded-md shadow-md">
            <h2 className="text-2xl font-semibold mb-6">Contact Us</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label className="block mb-1 font-medium" htmlFor="name">
                        Name
                    </label>
                    <input
                        type="text"
                        id="name"
                        className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:border-blue-500"
                        placeholder="Your Name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        required
                    />
                </div>

                <div>
                    <label className="block mb-1 font-medium" htmlFor="email">
                        Email
                    </label>
                    <input
                        type="email"
                        id="email"
                        className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:border-blue-500"
                        placeholder="Your Email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        required
                    />
                </div>

                <div>
                    <label className="block mb-1 font-medium" htmlFor="messageType">
                        Message Type
                    </label>
                    <select
                        id="messageType"
                        className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:border-blue-500"
                        value={messageType}
                        onChange={(e) => setMessageType(e.target.value)}
                    >
                        <option value="Feedback">Feedback</option>
                        <option value="Report">Report</option>
                    </select>
                </div>

                <div>
                    <label className="block mb-1 font-medium" htmlFor="message">
                        Message
                    </label>
                    <textarea
                        id="message"
                        className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:border-blue-500"
                        placeholder="Describe your feedback or issue..."
                        rows="5"
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        required
                    ></textarea>
                </div>

                <div>
                    <label className="block mb-1 font-medium" htmlFor="attachment">
                        Attachment (optional)
                    </label>
                    <input
                        type="file"
                        id="attachment"
                        className="w-full"
                        onChange={handleFileChange}
                    />
                </div>

                <button
                    type="submit"
                    disabled={isSubmitting}
                    className={`px-4 py-2 rounded text-white ${
                        isSubmitting ? 'bg-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700'
                    }`}
                >
                    {isSubmitting ? 'Sending...' : 'Submit'}
                </button>
            </form>
        </div>
    );
};

export default ContactUs;
