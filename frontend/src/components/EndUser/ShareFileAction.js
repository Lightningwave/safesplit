import React, { useState } from 'react';
import { Share2, Copy, X } from 'lucide-react';

const ShareFileAction = ({ file, user }) => {
   const [isModalOpen, setIsModalOpen] = useState(false);
   const [shareLink, setShareLink] = useState('');
   const [password, setPassword] = useState('');
   const [email, setEmail] = useState('');
   const [expiresAt, setExpiresAt] = useState('');
   const [maxDownloads, setMaxDownloads] = useState('');
   const [isLoading, setIsLoading] = useState(false);
   const [error, setError] = useState('');
   const [copySuccess, setCopySuccess] = useState(false);

   const isPremium = user?.role === 'premium_user';

   const handleShare = async (e) => {
       e.preventDefault();
       setIsLoading(true);
       setError('');

       try {
           const token = localStorage.getItem('token');
           if (!token) {
               setError('Please log in to share files.');
               return;
           }

           // Determine share type based on email presence
           const shareType = email ? 'recipient' : 'normal';

           const shareData = {
               password: password,
               share_type: shareType,
               ...(email && { email }),
               ...(isPremium && {
                   expires_at: expiresAt ? new Date(expiresAt).toISOString() : null,
                   max_downloads: maxDownloads ? parseInt(maxDownloads) : null
               })
           };

           const endpoint = isPremium 
               ? `/api/premium/shares/files/${file.id}`
               : `/api/files/${file.id}/share`;

           const response = await fetch(endpoint, {
               method: 'POST',
               headers: {
                   'Content-Type': 'application/json',
                   'Authorization': `Bearer ${token}`,
               },
               body: JSON.stringify(shareData),
           });

           const data = await response.json();

           if (!response.ok) {
               throw new Error(data.error || 'Failed to create share link');
           }

           if (data.status === "success") {
            setShareLink(data.data.share_link);
        } else {
            throw new Error(data.error || 'Failed to create share link');
        }

       } catch (error) {
           setError(error.message);
       } finally {
           setIsLoading(false);
       }
   };

   const copyToClipboard = async () => {
       try {
           await navigator.clipboard.writeText(shareLink);
           setCopySuccess(true);
           setTimeout(() => setCopySuccess(false), 2000);
       } catch (err) {
           setError('Failed to copy to clipboard');
       }
   };

   const closeModal = () => {
       setIsModalOpen(false);
       setShareLink('');
       setPassword('');
       setEmail('');
       setExpiresAt('');
       setMaxDownloads('');
       setError('');
       setCopySuccess(false);
   };

   return (
       <>
           <button
               onClick={() => setIsModalOpen(true)}
               className="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center space-x-2"
           >
               <Share2 size={16} />
               <span>Share</span>
           </button>

           {isModalOpen && (
               <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                   <div className="bg-white rounded-lg p-6 max-w-md w-full">
                       <div className="flex justify-between items-center mb-4">
                           <div>
                               <h2 className="text-xl font-semibold">Share {file.original_name}</h2>
                               {isPremium && (
                                   <span className="text-sm text-blue-600">Premium Share</span>
                               )}
                           </div>
                           <button onClick={closeModal} className="p-1 hover:bg-gray-100 rounded">
                               <X size={20} />
                           </button>
                       </div>

                       {error && (
                           <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                               {error}
                           </div>
                       )}

                       {!shareLink ? (
                           <form onSubmit={handleShare} className="space-y-4">
                               <div>
                                   <label className="block text-sm font-medium text-gray-700 mb-1">
                                       Password
                                   </label>
                                   <input
                                       type="password"
                                       value={password}
                                       onChange={(e) => setPassword(e.target.value)}
                                       required
                                       minLength={6}
                                       className="w-full px-3 py-2 border rounded-md"
                                       placeholder="Minimum 6 characters"
                                   />
                               </div>

                               <div>
                                   <label className="block text-sm font-medium text-gray-700 mb-1">
                                       Recipient Email (Optional)
                                   </label>
                                   <input
                                       type="email"
                                       value={email}
                                       onChange={(e) => setEmail(e.target.value)}
                                       className="w-full px-3 py-2 border rounded-md"
                                       placeholder="Enter recipient email"
                                   />
                                   {email && (
                                       <p className="text-sm text-gray-500 mt-1">
                                           Recipient will receive email verification
                                       </p>
                                   )}
                               </div>

                               {isPremium && (
                                   <>
                                       <div>
                                           <label className="block text-sm font-medium text-gray-700 mb-1">
                                               Expiration Date (Optional)
                                           </label>
                                           <input
                                               type="datetime-local"
                                               value={expiresAt}
                                               onChange={(e) => setExpiresAt(e.target.value)}
                                               className="w-full px-3 py-2 border rounded-md"
                                           />
                                       </div>

                                       <div>
                                           <label className="block text-sm font-medium text-gray-700 mb-1">
                                               Maximum Downloads (Optional)
                                           </label>
                                           <input
                                               type="number"
                                               value={maxDownloads}
                                               onChange={(e) => setMaxDownloads(e.target.value)}
                                               min="1"
                                               className="w-full px-3 py-2 border rounded-md"
                                               placeholder="Unlimited if not set"
                                           />
                                       </div>
                                   </>
                               )}

                               <button
                                   type="submit"
                                   disabled={isLoading}
                                   className="w-full bg-blue-500 text-white py-2 px-4 rounded-md 
                                            hover:bg-blue-600 disabled:bg-blue-300 disabled:cursor-not-allowed"
                               >
                                   {isLoading ? 'Creating Share...' : 'Create Share Link'}
                               </button>
                           </form>
                       ) : (
                           <div className="space-y-4">
                               <div className="flex items-center space-x-2">
                                   <input
                                       type="text"
                                       value={shareLink}
                                       readOnly
                                       className="flex-1 px-3 py-2 border rounded-md bg-gray-50"
                                   />
                                   <button
                                       onClick={copyToClipboard}
                                       className={`p-2 rounded-md transition-colors 
                                           ${copySuccess ? 'bg-green-100 text-green-600' : 'bg-gray-100 hover:bg-gray-200'}`}
                                   >
                                       <Copy size={20} />
                                   </button>
                               </div>
                               {copySuccess && (
                                   <p className="text-sm text-green-600">Copied to clipboard!</p>
                               )}
                               <div className="space-y-2 text-sm text-gray-600">
                                   <p>Password: {password}</p>
                                   {email && <p>Recipient: {email}</p>}
                                   {isPremium && expiresAt && (
                                       <p>Expires: {new Date(expiresAt).toLocaleString()}</p>
                                   )}
                                   {isPremium && maxDownloads && (
                                       <p>Max Downloads: {maxDownloads}</p>
                                   )}
                               </div>
                           </div>
                       )}
                   </div>
               </div>
           )}
       </>
   );
};

export default ShareFileAction;