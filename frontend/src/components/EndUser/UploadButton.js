import React, { useState, useRef, useEffect } from 'react';
import { Upload, ChevronDown } from 'lucide-react';

const UploadButton = ({ onSingleUpload, onMassUpload }) => {
    const [showDropdown, setShowDropdown] = useState(false);
    const dropdownRef = useRef(null);

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setShowDropdown(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

    return (
        <div className="relative" ref={dropdownRef}>
            <div className="flex">
                <button
                    onClick={onSingleUpload}
                    className="flex items-center space-x-2 px-4 py-2 bg-gray-600 text-white rounded-l-md hover:bg-gray-700 transition-colors"
                >
                    <Upload size={20} />
                    <span>Upload</span>
                </button>
                <button
                    onClick={() => setShowDropdown(!showDropdown)}
                    className="flex items-center px-2 bg-gray-600 text-white rounded-r-md border-l border-gray-500 hover:bg-gray-700 transition-colors"
                >
                    <ChevronDown size={20} />
                </button>
            </div>

            {showDropdown && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg z-20 border">
                    <button
                        onClick={() => {
                            onSingleUpload();
                            setShowDropdown(false);
                        }}
                        className="w-full text-left px-4 py-2 hover:bg-gray-100 text-sm"
                    >
                        Single File Upload
                    </button>
                    <button
                        onClick={() => {
                            onMassUpload();
                            setShowDropdown(false);
                        }}
                        className="w-full text-left px-4 py-2 hover:bg-gray-100 text-sm"
                    >
                        Multiple Files Upload
                    </button>
                </div>
            )}
        </div>
    );
};

export default UploadButton;