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
      <section className="w-full py-12 md:py-24 lg:py-32 xl:py-48">
          <div className="container mx-auto max-w-[1200px] px-4">
            <div className="flex flex-col items-center space-y-8 text-center">
              <h1 className="text-[2.5rem] sm:text-5xl md:text-6xl lg:text-7xl font-bold leading-tight tracking-tight">
               Enterprise-Grade Security with Distributed Storage
              </h1>
              <p className="max-w-[700px] text-xl text-gray-600">
                Protect your files with AES-256-GCM encryption, secured by Shamir's Secret Sharing 
                for key management and Reed–Solomon code for reliable distributed storage.
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
                <Lock className="h-6 w-6" />
                <h3 className="text-xl font-bold">AES-256-GCM Encryption</h3>
                <p className="text-gray-600">
                  Your files are secured using AES-256-GCM encryption, providing 
                  enterprise-grade security with authenticated encryption.
                </p>
              </div>
              <div className="flex flex-col items-center text-center space-y-4">
                <Shield className="h-6 w-6" />
                <h3 className="text-xl font-bold">Distributed Key Management</h3>
                <p className="text-gray-600">
                  Encryption keys are protected using Shamir's Secret Sharing, 
                  ensuring secure and distributed key management.
                </p>
              </div>
              <div className="flex flex-col items-center text-center space-y-4">
                <Upload className="h-6 w-6" />
                <h3 className="text-xl font-bold">Redundant Storage</h3>
                <p className="text-gray-600">
                  Encrypted data is split using Reed–Solomon code and distributed 
                  across multiple storage nodes for reliability.
                </p>
              </div>
            </div>
          </div>
        </section>

    {/* Demo Section */}
    <section className="w-full py-24">
      <div className="container mx-auto max-w-[1200px] px-4">
        <h2 className="text-4xl md:text-5xl font-bold text-center mb-16">
          See How It Works
        </h2>
        <div className="w-full max-w-[1200px] mx-auto">
          <div className="relative w-full pt-[56.25%]">
            <iframe
              src="https://www.youtube.com/embed/0SLUr2ZKELU"
              title="Safesplit Demo"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
              className="absolute inset-0 w-full h-full rounded-lg shadow-lg"
            />
          </div>
        </div>
        <p className="text-center mt-8 text-gray-600 max-w-2xl mx-auto">
          Watch our demo to see how Safesplit makes file sharing secure and simple using Shamir's Secret Sharing and Reed-Solomon algorithms.
        </p>
      </div>
    </section>

        {/* Pricing Section */}
        <section id="pricing" className="w-full py-24">
        <div className="container mx-auto max-w-[1200px] px-4">
            <h2 className="text-4xl md:text-5xl font-bold text-center mb-16">
            Pricing Plans
            </h2>
            <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
            {/* Free Plan */}
            <div className="p-8 bg-white rounded-lg border">
                <h3 className="text-2xl font-bold text-center mb-6">Free Plan</h3>
                <div className="text-center mb-8">
                <span className="text-3xl font-bold">Free</span>
                </div>
                <ul className="space-y-4 mb-8">
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>5 GB Storage Capacity</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>File Operations</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>File Management</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Security Features</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Data Encryption(AES-256-GCM)</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Activity Tracking</span>
                </li>
                <li className="flex items-center opacity-50">
                    <svg className="h-5 w-5 mr-3 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 6L6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span>File Sharing</span>
                </li>
                <li className="flex items-center opacity-50">
                    <svg className="h-5 w-5 mr-3 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 6L6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span>File Recovery</span>
                </li>
                <li className="flex items-center opacity-50">
                    <svg className="h-5 w-5 mr-3 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 6L6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span>Fast Download</span>
                </li>
                <li className="flex items-center opacity-50">
                    <svg className="h-5 w-5 mr-3 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 6L6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span>Premium Features</span>
                </li>
                <li className="flex items-center opacity-50">
                    <svg className="h-5 w-5 mr-3 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 6L6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span>Priority Customer Support</span>
                </li>
                </ul>
                <button 
                onClick={() => navigate('/register')}
                className="w-full py-2 border border-black rounded hover:bg-gray-50 transition-colors"
                >
                Get Started
                </button>
            </div>

            {/* Premium Plan */}
            <div className="p-8 bg-white rounded-lg border-2 border-black">
                <h3 className="text-2xl font-bold text-center mb-6">Premium Plan</h3>
                <div className="text-center mb-8">
                <span className="text-3xl font-bold">$8.99</span>
                <span className="text-gray-500">/month</span>
                </div>
                <ul className="space-y-4 mb-8">
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>50 GB Storage Capacity</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>File Operations</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>File Management</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Security Features</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Data Encryption(AES-256-GCM,ChaCha20,TwoFish)</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Activity Tracking</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>File Sharing</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>File Recovery</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Fast Download</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Premium Features</span>
                </li>
                <li className="flex items-center">
                    <Check className="h-5 w-5 mr-3 text-green-500" />
                    <span>Priority Customer Support</span>
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
          <p className="text-sm text-gray-500">© 2025 Safesplit. All rights reserved.</p>
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