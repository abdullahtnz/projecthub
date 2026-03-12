import { useState, useEffect, useCallback } from 'react';
import { getFeed, getUsername } from '../services/api';
import { Post } from '../types';
import PostCard from '../components/posts/PostCard';
import CreatePost from '../components/posts/CreatePost';
import { useAuth } from '../contexts/AuthContext';

const Feed: React.FC = () => {
  const { token } = useAuth();
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchPosts = useCallback(async () => {
    setIsLoading(true);
    setError('');

    try {
      const data = await getFeed();
      
      // Get unique user IDs
      const userIds = Array.from(new Set(data.map(p => p.user_id).filter(Boolean)));
      
      // Fetch usernames for all unique users
      const usernameMap: Record<string, string> = {};
      await Promise.all(
        userIds.map(async (userId) => {
          try {
            const userData = await getUsername(userId);
            usernameMap[userId] = userData.username;
          } catch {
            usernameMap[userId] = 'Unknown User';
          }
        })
      );
      
      // Attach username to each post
      const postsWithUsernames = data.map(post => ({
        ...post,
        user: {
          id: post.user_id,
          username: usernameMap[post.user_id] || 'Unknown User',
          email: ''
        }
      }));
      
      setPosts(postsWithUsernames);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string; error?: string } } };
      setError(error.response?.data?.message || error.response?.data?.error || 'Failed to load posts');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  const handlePostCreated = () => {
    fetchPosts();
  };

  if (isLoading) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center justify-center h-64">
          <div className="flex flex-col items-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
            <p className="mt-4 text-gray-500">Loading posts...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-3xl mx-auto">
      {token && (
        <CreatePost onPostCreated={handlePostCreated} />
      )}

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-4">
          <p>{error}</p>
          <button 
            onClick={() => fetchPosts()} 
            className="text-sm underline mt-2 hover:text-red-800"
          >
            Try again
          </button>
        </div>
      )}

      {posts.length === 0 ? (
        <div className="text-center py-12">
          <div className="bg-gray-100 rounded-full p-6 w-24 h-24 mx-auto mb-4 flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-gray-700 mb-2">No posts yet</h3>
          <p className="text-gray-500">
            {token ? 'Be the first to share something!' : 'Login to see posts from the community'}
          </p>
        </div>
      ) : (
        <div>
          {posts.map((post) => (
            <PostCard 
              key={post.id} 
              post={post} 
              onPostUpdate={() => fetchPosts()}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default Feed;
