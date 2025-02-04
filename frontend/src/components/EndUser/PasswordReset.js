import React, { useState } from 'react';
import { Loader2, Check, AlertCircle } from 'lucide-react';

export default function PasswordReset() {
 const [formData, setFormData] = useState({
   currentPassword: '',
   newPassword: '',
   confirmPassword: ''
 });
 const [loading, setLoading] = useState(false);
 const [success, setSuccess] = useState(false);
 const [error, setError] = useState(null);
 const [validationErrors, setValidationErrors] = useState({});

 const validatePassword = (password) => {
   const errors = [];
   if (password.length < 8) errors.push('At least 8 characters');
   if (!/[A-Z]/.test(password)) errors.push('One uppercase letter');
   if (!/[a-z]/.test(password)) errors.push('One lowercase letter');
   if (!/[0-9]/.test(password)) errors.push('One number');
   return errors;
 };

 const handleChange = (e) => {
   const { name, value } = e.target;
   setFormData(prev => ({
     ...prev,
     [name]: value
   }));
   setError(null);
   setSuccess(false);

   if (name === 'newPassword') {
     const errors = validatePassword(value);
     setValidationErrors(prev => ({
       ...prev,
       newPassword: errors
     }));
   }
 };

 const handleSubmit = async (e) => {
   e.preventDefault();
   setError(null);
   setSuccess(false);

   const passwordErrors = validatePassword(formData.newPassword);
   if (passwordErrors.length > 0) {
     setValidationErrors({ newPassword: passwordErrors });
     setError('Please fix the password errors.');
     return;
   }

   if (formData.newPassword !== formData.confirmPassword) {
     setError('Passwords do not match');
     return;
   }

   if (formData.currentPassword === formData.newPassword) {
     setError('New password must be different from current password');
     return;
   }

   setLoading(true);

   try {
    const token = localStorage.getItem('token');
    if (!token) {
      throw new Error('Not authenticated. Please login again.');
    }
 
    const response = await fetch('http://localhost:8080/api/reset-password', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        current_password: formData.currentPassword,
        new_password: formData.newPassword
      })
    });
 
    if (!response.ok) {
      const errorData = await response.json().catch(() => null);
      throw new Error(errorData?.error || `Server error: ${response.status}`);
    }
 
    const data = await response.json();
    setSuccess(true);
    setFormData({
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    });
    setValidationErrors({});
    
  } catch (err) {
    console.error('Password reset error:', err);
    setError(err.message);
  } finally {
    setLoading(false);
  }
 };

 return (
   <div>
     <h2 className="text-xl font-semibold mb-4">Change Password</h2>
     
     {error && (
       <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md flex items-center text-red-700">
         <AlertCircle className="mr-2 h-5 w-5" />
         {error}
       </div>
     )}

     {success && (
       <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-md flex items-center text-green-700">
         <Check className="mr-2 h-5 w-5" />
         Password updated successfully
       </div>
     )}

     <form onSubmit={handleSubmit} className="space-y-4 max-w-sm">
       <div>
         <label className="block mb-1 font-medium">Current Password</label>
         <input
           type="password"
           name="currentPassword"
           value={formData.currentPassword}
           onChange={handleChange}
           className="w-full border rounded px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
           required
           disabled={loading}
         />
       </div>

       <div>
         <label className="block mb-1 font-medium">New Password</label>
         <input
           type="password"
           name="newPassword"
           value={formData.newPassword}
           onChange={handleChange}
           className={`w-full border rounded px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent ${validationErrors.newPassword && 'border-red-500'}`}
           required
           disabled={loading}
         />
         {validationErrors.newPassword && (
           <ul className="text-sm text-red-600 mt-1 pl-5 list-disc">
             {validationErrors.newPassword.map((err, index) => (
               <li key={index}>{err}</li>
             ))}
           </ul>
         )}
       </div>

       <div>
         <label className="block mb-1 font-medium">Confirm New Password</label>
         <input
           type="password"
           name="confirmPassword"
           value={formData.confirmPassword}
           onChange={handleChange}
           className="w-full border rounded px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
           required
           disabled={loading}
         />
       </div>

       <div className="text-sm text-gray-600 space-y-1">
         <p>Password must:</p>
         <ul className="list-disc pl-5">
           <li>Be at least 8 characters long</li>
           <li>Contain at least one uppercase letter</li>
           <li>Contain at least one lowercase letter</li>
           <li>Contain at least one number</li>
         </ul>
       </div>

       <button 
         type="submit"
         disabled={loading}
         className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
       >
         {loading ? (
           <>
             <Loader2 className="animate-spin mr-2 h-5 w-5" />
             Updating Password...
           </>
         ) : 'Update Password'}
       </button>
     </form>
   </div>
 );
}