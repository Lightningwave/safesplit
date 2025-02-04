import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertCircle, Check, CreditCard } from 'lucide-react';

const BillingPage = ({ user, onUpgradeSuccess }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const [billingInfo, setBillingInfo] = useState({
    cardNumber: '',
    cvv: '',
    expiryMonth: '',
    expiryYear: '',
    cardHolder: '',
    billingCycle: 'monthly',
    billingName: '',
    billingEmail: '',
    billingAddress: '',
    countryCode: ''
  });

  const formatCardNumber = (number) => {
    const digits = number.replace(/\D/g, '');
    const groups = [];
    for (let i = 0; i < digits.length && i < 16; i += 4) {
      groups.push(digits.slice(i, i + 4));
    }
    return groups.join(' ');
  };

  const getCardType = (number) => {
    const cleaned = number.replace(/\D/g, '');
    if (!cleaned) return null;
    
    const patterns = {
      visa: /^4/,
      mastercard: /^5[1-5]/,
      amex: /^3[47]/,
      discover: /^6(?:011|5)/
    };

    for (const [card, pattern] of Object.entries(patterns)) {
      if (pattern.test(cleaned)) return card;
    }
    return null;
  };

  const validateCard = () => {
    const cardNumber = billingInfo.cardNumber.replace(/\s/g, '');
    if (cardNumber.length < 12 || cardNumber.length > 19) {
      throw new Error('Card number must be between 12 and 19 digits');
    }
    
    const cardType = getCardType(cardNumber);
    if (!cardType) {
      throw new Error('Invalid card type');
    }

    const currentDate = new Date();
    const currentYear = currentDate.getFullYear() % 100;
    const currentMonth = currentDate.getMonth() + 1;

    const expiryYear = parseInt(billingInfo.expiryYear);
    const expiryMonth = parseInt(billingInfo.expiryMonth);

    if (expiryYear < currentYear || 
        (expiryYear === currentYear && expiryMonth < currentMonth)) {
      throw new Error('Card has expired');
    }

    return true;
  };

  const handleUpgrade = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      validateCard();
      const token = localStorage.getItem('token');

      const paymentResponse = await fetch('/api/payment/upgrade', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          cardNumber: billingInfo.cardNumber.replace(/\s/g, ''),
          cvv: billingInfo.cvv,
          expiryMonth: parseInt(billingInfo.expiryMonth),
          expiryYear: parseInt('20' + billingInfo.expiryYear),
          cardHolder: billingInfo.cardHolder,
          billingCycle: billingInfo.billingCycle,
          billingName: billingInfo.billingName,
          billingEmail: billingInfo.billingEmail,
          billingAddress: billingInfo.billingAddress,
          countryCode: billingInfo.countryCode
        })
      });

      if (!paymentResponse.ok) {
        const errorData = await paymentResponse.json();
        throw new Error(errorData.error || 'Payment processing failed');
      }

      if (onUpgradeSuccess) onUpgradeSuccess();
      navigate('/premium-dashboard?upgraded=true');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold mb-6">Upgrade to Premium</h1>
      
      <div className="grid md:grid-cols-3 gap-8">
        <div className="md:col-span-2 space-y-6">
          {error && (
            <div className="flex items-center space-x-2 p-4 bg-red-50 text-red-700 rounded-md">
              <AlertCircle size={20} />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleUpgrade} className="space-y-6">
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold mb-4">Select Plan</h2>
              <div className="grid grid-cols-2 gap-4">
                <button
                  type="button"
                  className={`p-4 border rounded-lg ${billingInfo.billingCycle === 'monthly' ? 'border-blue-500 bg-blue-50' : ''}`}
                  onClick={() => setBillingInfo({...billingInfo, billingCycle: 'monthly'})}
                >
                  <div className="font-semibold">Monthly Premium</div>
                  <div className="text-xl font-bold mt-2">$8.99/month</div>
                  <div className="text-sm text-gray-600 mt-1">$107.88/year</div>
                </button>
                <button
                  type="button"
                  className={`p-4 border rounded-lg ${billingInfo.billingCycle === 'yearly' ? 'border-blue-500 bg-blue-50' : ''}`}
                  onClick={() => setBillingInfo({...billingInfo, billingCycle: 'yearly'})}
                >
                  <div className="font-semibold">Yearly Premium</div>
                  <div className="text-xl font-bold mt-2">$89.99/year</div>
                  <div className="text-sm text-green-600 mt-1">Save ~17%</div>
                </button>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold mb-4">Billing Information</h2>
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium mb-1">Name</label>
                    <input className="w-full p-2 border rounded-md" required
                      value={billingInfo.billingName}
                      onChange={e => setBillingInfo({...billingInfo, billingName: e.target.value})}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1">Email</label>
                    <input type="email" className="w-full p-2 border rounded-md" required
                      value={billingInfo.billingEmail}
                      onChange={e => setBillingInfo({...billingInfo, billingEmail: e.target.value})}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Billing Address</label>
                  <textarea className="w-full p-2 border rounded-md" required
                    value={billingInfo.billingAddress}
                    onChange={e => setBillingInfo({...billingInfo, billingAddress: e.target.value})}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Country</label>
                  <select className="w-full p-2 border rounded-md" required
                    value={billingInfo.countryCode}
                    onChange={e => setBillingInfo({...billingInfo, countryCode: e.target.value})}
                  >
                    <option value="">Select country</option>
                    <option value="US">United States</option>
                    <option value="CA">Canada</option>
                    <option value="GB">United Kingdom</option>
                  </select>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold mb-4">Payment Information</h2>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1">Card Number</label>
                  <div className="relative">
                    <input 
                      className="w-full p-2 pr-10 border rounded-md font-mono"
                      required
                      value={formatCardNumber(billingInfo.cardNumber)}
                      onChange={e => setBillingInfo({
                        ...billingInfo,
                        cardNumber: e.target.value.replace(/\D/g, '')
                      })}
                      placeholder="1234 5678 9012 3456"
                    />
                    <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                      {getCardType(billingInfo.cardNumber) ? (
                        <span className="font-medium text-gray-700 uppercase">
                          {getCardType(billingInfo.cardNumber)}
                        </span>
                      ) : (
                        <CreditCard className="h-5 w-5 text-gray-400" />
                      )}
                    </div>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium mb-1">Expiry (MM/YY)</label>
                    <input className="w-full p-2 border rounded-md" placeholder="MM/YY" required
                      value={`${billingInfo.expiryMonth}/${billingInfo.expiryYear}`}
                      onChange={e => {
                        const [month, year] = e.target.value.split('/');
                        setBillingInfo({
                          ...billingInfo,
                          expiryMonth: month?.slice(0, 2) || '',
                          expiryYear: year?.slice(0, 2) || ''
                        });
                      }}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1">CVV</label>
                    <input className="w-full p-2 border rounded-md" type="password" maxLength="4" required
                      value={billingInfo.cvv}
                      onChange={e => setBillingInfo({
                        ...billingInfo,
                        cvv: e.target.value.replace(/\D/g, '').slice(0, 4)
                      })}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Card Holder Name</label>
                  <input className="w-full p-2 border rounded-md" required
                    value={billingInfo.cardHolder}
                    onChange={e => setBillingInfo({...billingInfo, cardHolder: e.target.value})}
                  />
                </div>
              </div>
            </div>

            <button type="submit" disabled={loading}
              className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50">
              {loading ? 'Processing...' : `Upgrade Now - ${billingInfo.billingCycle === 'yearly' ? '$89.99/year' : '$8.99/month'}`}
            </button>
          </form>
        </div>

        <div className="space-y-6">
          <div className="bg-blue-50 rounded-lg shadow p-6 border-2 border-blue-500">
            <h3 className="text-xl font-semibold mb-4">Premium Features</h3>
            <div className="space-y-4">
              <div>50 GB Storage</div>
              <ul className="space-y-2">
                {[
                  "File Operations",
                  "File Management",
                  "Security Features",
                  "Data Encryption",
                  "Activity Tracking",
                  "File Sharing",
                  "File Recovery",
                  "Fast Download",
                  "Premium Features",
                  "Priority Support"
                ].map(feature => (
                  <li key={feature} className="flex items-center">
                    <Check className="h-4 w-4 text-green-500 mr-2" />
                    {feature}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default BillingPage;