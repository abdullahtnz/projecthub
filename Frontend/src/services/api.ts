import axios from 'axios';
import { LoginCredentials, SignupCredentials, AuthResponse, Post, LikeResponse, LikeUser } from '../types';

const API_URL = 'http://localhost:8080';

const api = axios.create({
  baseURL: API_URL,
});

// Add token to requests if it exists
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Auth endpoints
export const login = async (credentials: LoginCredentials) => {
  const response = await api.post<AuthResponse>('/login', credentials);
  return response.data;
};

export const signup = async (credentials: SignupCredentials) => {
  const response = await api.post<AuthResponse>('/signup', credentials);
  return response.data;
};

// Posts endpoints
export const getFeed = async () => {
  const response = await api.get<Post[]>('/posts/feed');
  return response.data;
};

export const createPost = async (data: CreatePostData) => {
  const formData = new FormData();
  formData.append('content', data.content);
  
  if (data.images) {
    data.images.forEach((image) => {
      formData.append('images', image);
    });
  }

  const response = await api.post('/posts', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

// Like endpoints
export const likePost = async (postId: string) => {
  const response = await api.post<LikeResponse>(`/posts/${postId}/like`);
  return response.data;
};

export const unlikePost = async (postId: string) => {
  const response = await api.post<LikeResponse>(`/posts/${postId}/unlike`);
  return response.data;
};

export const getLikeStatus = async (postId: string) => {
  const response = await api.get<LikeResponse>(`/posts/${postId}/like-status`);
  return response.data;
};

export const getPostLikes = async (postId: string) => {
  const response = await api.get<LikeUser[]>(`/posts/${postId}/likes`);
  return response.data;
};

// Helper to get image URL
export const getImageUrl = (filename: string) => {
  return `${API_URL}/uploads/${filename}`;
};

export default api;