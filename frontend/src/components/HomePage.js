import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowRight, Upload, Lock, Shield, Check } from 'lucide-react';

function HomePage() {
  const navigate = useNavigate();
  
  const handleGetStarted = (e) => {
    e.preventDefault();
    navigate('/register');
  };

  return (
    <div className="flex flex-col min-h-screen">
      <main className="flex-1">
        {/* Hero Section */}
        <section className="w-full py-12 md:py-24 lg:py-32 xl:py-48">
          <div className="container mx-auto max-w-[1200px] px-4">
            <div className="flex flex-col items-center space-y-8 text-center">
              <h1 className="text-[2.5rem] sm:text-5xl md:text-6xl lg:text-7xl font-bold leading-tight tracking-tight">
                Secure File Sharing with Shamir's Secret
              </h1>
              <p className="max-w-[700px] text-xl text-gray-600">
                Split your files into secure shares, share them safely, and recover them with ease. 
                Safesplit uses Shamir's Secret Sharing for unparalleled security.
              </p>
              <div className="w-full max-w-sm space-y-4">
                <form onSubmit={handleGetStarted} className="flex space-x-2">
                  <input
                    type="email"
                    placeholder="Enter your email"
                    className="flex-1 h-12 px-4 rounded-md border border-gray-200 focus:outline-none focus:ring-2 focus:ring-black"
                  />
                  <button
                    type="submit"
                    className="h-12 px-6 rounded-md bg-black text-white hover:bg-gray-800 flex items-center"
                  >
                    Get Started
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </button>
                </form>
                <p className="text-sm text-gray-500">
                  Start securing your files today. No credit card required.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* How it Works Section */}
        <section id="how-it-works" className="w-full py-24 bg-gray-50">
          <div className="container mx-auto max-w-[1200px] px-4">
            <h2 className="text-4xl md:text-5xl font-bold text-center mb-16">
              How Safesplit Works
            </h2>
            <div className="grid md:grid-cols-3 gap-12">
              <div className="flex flex-col items-center text-center space-y-4">
                <Upload className="h-6 w-6" />
                <h3 className="text-xl font-bold">Upload & Split</h3>
                <p className="text-gray-600">
                  Upload your file and split it into multiple secure shares using Shamir's Secret Sharing.
                </p>
              </div>
              <div className="flex flex-col items-center text-center space-y-4">
                <Lock className="h-6 w-6" />
                <h3 className="text-xl font-bold">Distribute Shares</h3>
                <p className="text-gray-600">
                  Safely distribute the shares to trusted parties or store them in different locations.
                </p>
              </div>
              <div className="flex flex-col items-center text-center space-y-4">
                <Shield className="h-6 w-6" />
                <h3 className="text-xl font-bold">Recover Securely</h3>
                <p className="text-gray-600">
                  Reconstruct your original file by combining a threshold number of shares.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Pricing Section */}
        <section id="pricing" className="w-full py-24">
          <div className="container mx-auto max-w-[1200px] px-4">
            <h2 className="text-4xl md:text-5xl font-bold text-center mb-16">
              Pricing Plans
            </h2>
            <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
              <div className="p-8 bg-white rounded-lg border">
                <h3 className="text-2xl font-bold text-center mb-6">Free Plan</h3>
                <ul className="space-y-4 mb-8">
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Up to 100MB file size</span>
                  </li>
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>5 shares per file</span>
                  </li>
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>7-day file retention</span>
                  </li>
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Basic email support</span>
                  </li>
                </ul>
                <button 
                  onClick={() => navigate('/register')}
                  className="w-full py-2 border border-black rounded hover:bg-gray-50 transition-colors"
                >
                  Get Started
                </button>
              </div>

              <div className="p-8 bg-white rounded-lg border-2 border-black">
                <h3 className="text-2xl font-bold text-center mb-6">Premium Plan</h3>
                <ul className="space-y-4 mb-8">
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Up to 10GB file size</span>
                  </li>
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Unlimited shares per file</span>
                  </li>
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>30-day file retention</span>
                  </li>
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Priority support</span>
                  </li>
                  <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Advanced security features</span>
                  </li>
                </ul>
                <button
                  onClick={() => navigate('/register')}
                  className="w-full py-2 bg-black text-white rounded hover:bg-gray-800 transition-colors"
                >
                  Upgrade to Premium
                </button>
              </div>
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="w-full py-24 bg-gray-50">
          <div className="container mx-auto max-w-[1200px] px-4 text-center">
            <h2 className="text-4xl font-bold mb-4">
              Ready to Secure Your Files?
            </h2>
            <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto">
              Join Safesplit today and experience the next level of file security and sharing.
            </p>
            <div className="flex justify-center gap-4">
              <button
                onClick={() => navigate('/register')}
                className="px-8 py-3 bg-black text-white rounded hover:bg-gray-800 transition-colors"
              >
                Get Started
              </button>
              <button
                onClick={() => navigate('/login')}
                className="px-8 py-3 border rounded hover:bg-gray-50 transition-colors"
              >
                Login
              </button>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="py-6 border-t">
        <div className="container mx-auto max-w-[1200px] px-4 flex flex-col sm:flex-row justify-between items-center">
          <p className="text-sm text-gray-500">© 2024 Safesplit. All rights reserved.</p>
          <div className="flex gap-6 mt-4 sm:mt-0">
            <Link to="#" className="text-sm text-gray-500 hover:text-gray-600">Terms of Service</Link>
            <Link to="#" className="text-sm text-gray-500 hover:text-gray-600">Privacy Policy</Link>
            <Link to="#" className="text-sm text-gray-500 hover:text-gray-600">Contact Us</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default HomePage;