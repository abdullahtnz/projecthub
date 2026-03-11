/* eslint-disable react-refresh/only-export-components */
import React, { createContext, useState, useContext, ReactNode } from 'react';
import { User, LoginCredentials, SignupCredentials } from '../types';
import { login as loginApi, signup as signupApi } from '../services/api';

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  signup: (credentials: SignupCredentials) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

const getStoredUser = (): User | null => {
  const stored = localStorage.getItem('user');
  return stored ? JSON.parse(stored) : null;
};

const getStoredToken = (): string | null => {
  return localStorage.getItem('token');
};

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(getStoredUser);
  const [token, setToken] = useState<string | null>(getStoredToken);
  const isLoading = false;

  const login = async (credentials: LoginCredentials) => {
    try {
      const response = await loginApi(credentials);
      if (response.token) {
        const userData: User = {
          id: response.user_id,
          email: response.email,
          username: response.username,
        };
        setUser(userData);
        setToken(response.token);
        localStorage.setItem('user', JSON.stringify(userData));
        localStorage.setItem('token', response.token);
      }
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const signup = async (credentials: SignupCredentials) => {
    try {
      const response = await signupApi(credentials);
      const userData: User = {
        id: response.user_id,
        email: response.email,
        username: response.username,
      };
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      // Note: Signup doesn't return token, user must login
    } catch (error) {
      console.error('Signup failed:', error);
      throw error;
    }
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('user');
    localStorage.removeItem('token');
  };

  return (
    <AuthContext.Provider value={{ user, token, isLoading, login, signup, logout }}>
      {children}
    </AuthContext.Provider>
  );
};