import React from 'react';
import { ArrowUp, ArrowDown } from 'lucide-react';

// Compoment to render the sorting UI for each column
function SortingHeader({ label, sortType, currentSortType, sortOrder, onSortChange, onToggleSortOrder }) {
    return (
        <div className="flex items-center space-x-4">
            <span>{label}</span>
            {sortType === currentSortType && (
                <button 
                    type="button" 
                    onClick={onToggleSortOrder}
                    className="ml-2"
                >
                    {sortOrder === 'asc' ? <ArrowUp size={14} /> : <ArrowDown size={14} />}
                </button>
            )}
        </div>
    );
}

export default SortingHeader;
