import React from 'react';
import { Link } from 'react-router-dom';
import { Github } from 'lucide-react';

function AboutPage() {
  const team = [
    {
      name: "HO RONG HUI",
      role: "Leader / Backend Developer",
      bio: "Guides the team to achieve project goals. Develops server-side systems and creates essential documentation.",
      image: "/Sample1.png",
      github: "https://github.com/username1"
    },
    {
      name: "CHEN XIAOJUAN",
      role: "Frontend Developer",
      bio: "Creates and updates essential documentation. Designs wireframe and front-end of the website.",
      image: "/Sample5.png",
      github: "https://github.com/xiaojuan-c"
    },
    {
      name: "LAI NIM LOONG RYAN",
      role: "Frontend Developer",
      bio: "Creates and updates essential documentation. Designs wireframe and front-end of the website.",
      image: "/Sample3.png",
      github: "https://github.com/Nimloongs "
    },
    {
      name: "PANG WAI PIN",
      role: "Frontend Developer",
      bio: "Creates and updates essential documentation. Designs wireframe and front-end of the website.",
      image: "/Sample4.png",
      github: "https://github.com/username4"
    },
    {
      name: "VAN LE NGUYEN",
      role: "Full Stack Developer",
      bio: "Creates and updates essential documentation. Develops both front-end and backend components of the website.",
      image: "/Sample2.png",
      github: "https://github.com/Lightningwave"
    },
    {
      name: "ZENG LIHONG",
      role: "Backend Developer",
      bio: "Creates and updates essential documentation. Develops server-side systems.",
      image: "/Sample6.png",
      github: "https://github.com/username6"
    }
  ];

  return (
    <div className="flex flex-col min-h-screen">
      <main className="flex-1">
        {/* About Section */}
        <section className="w-full py-12 md:py-24">
          <div className="container mx-auto max-w-[1200px] px-4">
            <div className="flex flex-col items-center text-center">
              <h1 className="text-4xl md:text-5xl font-bold mb-6">
                About Safesplit
              </h1>
              <p className="text-xl text-gray-600 max-w-[800px]">
                We're a team of passionate individuals dedicated to revolutionizing secure 
                file sharing using Shamir's Secret Sharing scheme.
              </p>
            </div>
          </div>
        </section>

        {/* Team Section */}
        <section className="w-full py-12 md:py-24 bg-gray-50">
          <div className="container mx-auto max-w-[1200px] px-4">
            <h2 className="text-3xl md:text-4xl font-bold text-center mb-16">
              Our Team
            </h2>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              {team.map((member) => (
                <div key={member.name} className="bg-white rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow">
                  {/* Image container with fixed dimensions */}
                  <div className="relative w-full h-64">
                    <img 
                      src={member.image} 
                      alt={member.name} 
                      className="w-full h-full object-cover" 
                      style={{ objectPosition: 'center top' }}
                      onError={(e) => {
                        e.target.src = "/placeholder.png";
                        e.target.onerror = null;
                      }}
                    />
                  </div>
                  <div className="p-6">
                    <h3 className="text-xl font-bold mb-1">{member.name}</h3>
                    <p className="text-sm text-gray-600 mb-4">{member.role}</p>
                    <p className="text-gray-600 mb-6">{member.bio}</p>
                    <div className="flex">
                      <a 
                        href={member.github} 
                        target="_blank" 
                        rel="noopener noreferrer" 
                        className="text-gray-600 hover:text-black transition-colors"
                      >
                        <Github className="h-5 w-5" />
                      </a>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Mission Section */}
        <section className="w-full py-12 md:py-24">
          <div className="container mx-auto max-w-[1200px] px-4">
            <div className="grid md:grid-cols-2 gap-12">
              <div>
                <h2 className="text-3xl font-bold mb-6">Our Mission</h2>
                <p className="text-gray-600">
                  We're committed to making secure file sharing accessible to everyone. 
                  By leveraging advanced cryptographic techniques like Shamir's Secret Sharing, 
                  we ensure your files remain private and secure while being easily shareable 
                  with those you trust.
                </p>
              </div>
              <div>
                <h2 className="text-3xl font-bold mb-6">Our Vision</h2>
                <p className="text-gray-600">
                  We envision a future where secure file sharing is the norm, not the exception. 
                  Our goal is to create innovative solutions that make cryptographic security 
                  intuitive and accessible to users worldwide.
                </p>
              </div>
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

export default AboutPage;