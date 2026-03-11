import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { getUserPostsById } from '../services/api';
import { Post } from '../types';
import PostCard from '../components/posts/PostCard';

const ProfileView: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    const fetchData = async () => {
      if (!userId) {
        setError('No user specified');
        setLoading(false);
        return;
      }
      try {
        const userPosts = await getUserPostsById(userId);
        setPosts(userPosts);
      } catch {
        setError('Failed to load posts');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [userId]);

  if (loading) {
    return (
      <div className="w-full max-w-3xl mx-auto">
        <div className="flex items-center justify-center h-64">
          <div className="flex flex-col items-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
            <p className="mt-4 text-gray-500">Loading profile...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-3xl mx-auto">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-6">
        <div className="h-32 bg-gradient-to-r from-primary-500 to-primary-700"></div>
        <div className="px-6 pb-6">
          <div className="relative -mt-16 mb-4">
            <div className="h-28 w-28 rounded-full bg-white p-1">
              <div className="h-full w-full rounded-full bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center text-white text-4xl font-bold">
                ?
              </div>
            </div>
          </div>
          
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">User Profile</h1>
              <p className="text-gray-500">@{userId?.slice(0, 8)}</p>
            </div>
            <div className="flex items-center space-x-2 bg-gray-100 px-4 py-2 rounded-lg">
              <span className="font-semibold text-gray-900">{posts.length}</span>
              <span className="text-gray-500">posts</span>
            </div>
          </div>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-4">
          <p>{error}</p>
        </div>
      )}

      <h2 className="text-lg font-semibold text-gray-800 mb-4">Posts</h2>

      {posts.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl border border-gray-200">
          <div className="bg-gray-100 rounded-full p-6 w-24 h-24 mx-auto mb-4 flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-gray-700 mb-2">No posts yet</h3>
        </div>
      ) : (
        <div>
          {posts.map((post) => (
            <PostCard key={post.id} post={post} />
          ))}
        </div>
      )}
    </div>
  );
};

export default ProfileView;
