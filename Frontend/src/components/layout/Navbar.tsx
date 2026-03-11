import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/login');
    setIsMenuOpen(false);
  };

  return (
    <nav className="bg-primary-900 border-b border-primary-700 shadow-xl sticky top-0 z-50">
      <div className="w-full px-4">
        <div className="flex justify-between items-center h-16">
          <Link to="/" className="text-xl md:text-2xl font-bold text-white hover:text-primary-400 transition flex items-center">
            <span className="bg-primary-600 text-white px-2 md:px-3 py-1 rounded-lg mr-2">S</span>
            <span className="hidden sm:inline">SocialApp</span>
          </Link>

          <button
            className="md:hidden text-gray-300 hover:text-white p-2"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              {isMenuOpen ? (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              )}
            </svg>
          </button>

          <div className="hidden md:flex items-center space-x-2">
            {user ? (
              <>
                <Link 
                  to="/" 
                  className="text-gray-300 hover:text-white hover:bg-primary-800 transition px-4 py-2 rounded-lg text-sm font-medium"
                >
                  Feed
                </Link>
                <Link 
                  to="/profile" 
                  className="text-gray-300 hover:text-white hover:bg-primary-800 transition px-4 py-2 rounded-lg text-sm font-medium"
                >
                  Profile
                </Link>
                
                <div className="h-6 w-px bg-primary-700 mx-2"></div>
                
                  <div className="flex items-center space-x-3">
                    <Link to={`/profile/${user.id}`} className="flex items-center">
                      <div className="h-8 w-8 rounded-full bg-primary-700 flex items-center justify-center text-white font-semibold text-sm border border-primary-600">
                        {user.username?.charAt(0).toUpperCase()}
                      </div>
                    </Link>
                    <Link to={`/profile/${user.id}`} className="ml-2 text-sm text-gray-300 hover:underline hidden md:inline">
                      {user.username}
                    </Link>
                  
                  <button
                    onClick={handleLogout}
                    className="bg-primary-800 hover:bg-primary-700 text-white px-4 py-2 rounded-lg transition text-sm font-medium border border-primary-600 flex items-center"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                    </svg>
                    Logout
                  </button>
                </div>
              </>
            ) : (
              <div className="flex items-center space-x-3">
                <Link
                  to="/login"
                  className="text-gray-300 hover:text-white hover:bg-primary-800 transition px-4 py-2 rounded-lg text-sm font-medium"
                >
                  Login
                </Link>
                <Link
                  to="/signup"
                  className="bg-primary-700 hover:bg-primary-600 text-white px-4 py-2 rounded-lg transition text-sm font-medium border border-primary-600 flex items-center"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
                  </svg>
                  Sign Up
                </Link>
              </div>
            )}
          </div>
        </div>

        {isMenuOpen && (
          <div className="md:hidden py-4 border-t border-primary-700">
            {user ? (
              <div className="space-y-2">
                <Link 
                  to="/" 
                  className="block text-gray-300 hover:text-white hover:bg-primary-800 transition px-4 py-2 rounded-lg text-sm font-medium"
                  onClick={() => setIsMenuOpen(false)}
                >
                  Feed
                </Link>
                <Link 
                  to="/profile" 
                  className="block text-gray-300 hover:text-white hover:bg-primary-800 transition px-4 py-2 rounded-lg text-sm font-medium"
                  onClick={() => setIsMenuOpen(false)}
                >
                  Profile
                </Link>
                <div className="flex items-center px-4 py-2">
                  <div className="h-8 w-8 rounded-full bg-primary-700 flex items-center justify-center text-white font-semibold text-sm border border-primary-600 mr-2">
                    {user.username?.charAt(0).toUpperCase()}
                  </div>
                  <span className="text-sm text-gray-300">{user.username}</span>
                </div>
                <button
                  onClick={handleLogout}
                  className="w-full text-left bg-primary-800 hover:bg-primary-700 text-white px-4 py-2 rounded-lg transition text-sm font-medium border border-primary-600"
                >
                  Logout
                </button>
              </div>
            ) : (
              <div className="space-y-2">
                <Link
                  to="/login"
                  className="block text-gray-300 hover:text-white hover:bg-primary-800 transition px-4 py-2 rounded-lg text-sm font-medium"
                  onClick={() => setIsMenuOpen(false)}
                >
                  Login
                </Link>
                <Link
                  to="/signup"
                  className="block bg-primary-700 hover:bg-primary-600 text-white px-4 py-2 rounded-lg transition text-sm font-medium border border-primary-600 text-center"
                  onClick={() => setIsMenuOpen(false)}
                >
                  Sign Up
                </Link>
              </div>
            )}
          </div>
        )}
      </div>
    </nav>
  );
};

export default Navbar;
