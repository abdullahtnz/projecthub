export interface User {
  id: string;
  email: string;
  username: string;
}

export interface Post {
  id: string;
  user_id: string;
  content: string;
  images: string[];
  created_at: string;
  like_count: number;
  liked_by_me?: boolean;
  user?: User; // We'll add this when fetching user details
}

export interface LikeResponse {
  liked: boolean;
  like_count: number;
  post_id: string;
}

export interface LikeUser {
  user_id: string;
  username: string;
  email: string;
  liked_at: string;
}

export interface AuthResponse {
  message: string;
  user_id: string;
  email: string;
  username: string;
  token?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface SignupCredentials {
  email: string;
  username: string;
  password: string;
}

export interface CreatePostData {
  content: string;
  images?: File[];
}