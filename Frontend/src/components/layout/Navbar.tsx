import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="bg-primary-600 text-white shadow-lg">
      <div className="container mx-auto px-4">
        <div className="flex justify-between items-center h-16">
          <Link to="/" className="text-2xl font-bold">
            SocialApp
          </Link>

          <div className="flex items-center space-x-4">
            {user ? (
              <>
                <Link to="/" className="hover:text-primary-200 transition">
                  Feed
                </Link>
                <Link to="/profile" className="hover:text-primary-200 transition">
                  Profile
                </Link>
                <span className="text-sm">Welcome, {user.username}</span>
                <button
                  onClick={handleLogout}
                  className="bg-primary-700 px-4 py-2 rounded-lg hover:bg-primary-800 transition"
                >
                  Logout
                </button>
              </>
            ) : (
              <>
                <Link to="/login" className="hover:text-primary-200 transition">
                  Login
                </Link>
                <Link
                  to="/signup"
                  className="bg-primary-700 px-4 py-2 rounded-lg hover:bg-primary-800 transition"
                >
                  Sign Up
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;