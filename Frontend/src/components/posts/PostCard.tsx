import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Post, LikeUser } from '../../types';
import { likePost, unlikePost, getLikeStatus, getPostLikes, getImageUrl } from '../../services/api';
import { useAuth } from '../../contexts/AuthContext';

interface PostCardProps {
  post: Post;
  onPostUpdate?: () => void;
}

const PostCard: React.FC<PostCardProps> = ({ post, onPostUpdate }) => {
  const { token } = useAuth();
  const [liked, setLiked] = useState(post.liked_by_me || false);
  const [likeCount, setLikeCount] = useState(post.like_count);
  const [isLoading, setIsLoading] = useState(false);
  const [showLikes, setShowLikes] = useState(false);
  const [likesList, setLikesList] = useState<LikeUser[]>([]);
  const [isLoadingLikes, setIsLoadingLikes] = useState(false);

  useEffect(() => {
    if (token) {
      getLikeStatus(post.id)
        .then((res) => {
          setLiked(res.liked);
          setLikeCount(res.like_count);
        })
        .catch(console.error);
    }
  }, [post.id, token]);

  const handleLike = async () => {
    if (!token || isLoading) return;
    setIsLoading(true);

    try {
      if (liked) {
        const res = await unlikePost(post.id);
        setLiked(res.liked);
        setLikeCount(res.like_count);
      } else {
        const res = await likePost(post.id);
        setLiked(res.liked);
        setLikeCount(res.like_count);
      }
      onPostUpdate?.();
    } catch (error) {
      console.error('Error toggling like:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleShowLikes = async () => {
    if (!token || showLikes) {
      setShowLikes(!showLikes);
      return;
    }
    setIsLoadingLikes(true);
    try {
      const users = await getPostLikes(post.id);
      setLikesList(users);
      setShowLikes(true);
    } catch (error) {
      console.error('Error fetching likes:', error);
    } finally {
      setIsLoadingLikes(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-4">
      <div className="p-4">
        <div className="flex items-center mb-3">
          <Link to={`/profile/${post.user_id || ''}`} className="flex items-center">
            <div className="h-10 w-10 rounded-full bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center text-white font-semibold">
              {post.user?.username?.charAt(0).toUpperCase() || '?'}
            </div>
          </Link>
          <div className="ml-3">
            <Link to={`/profile/${post.user_id || ''}`} className="font-semibold text-gray-900 hover:underline">{post.user?.username || 'Unknown User'}</Link>
            <p className="text-xs text-gray-500">{formatDate(post.created_at)}</p>
          </div>
        </div>

        <p className="text-gray-800 whitespace-pre-wrap mb-3">{post.content}</p>

        {post.images && post.images.length > 0 && (
          <div className={`grid gap-1 mb-3 ${post.images.length === 1 ? 'grid-cols-1' : 'grid-cols-2'}`}>
            {post.images.map((image, index) => (
              <img
                key={index}
                src={getImageUrl(image)}
                alt={`Post image ${index + 1}`}
                className="w-full max-h-96 object-contain rounded-lg bg-gray-100"
              />
            ))}
          </div>
        )}

        <div className="flex items-center pt-2 border-t border-gray-100">
          <button
            onClick={handleLike}
            disabled={!token || isLoading}
            className={`flex items-center space-x-2 px-3 py-2 rounded-lg transition-colors ${
              liked
                ? 'text-red-500 bg-red-50'
                : 'text-gray-500 hover:text-red-500 hover:bg-red-50'
            } ${!token ? 'opacity-50 cursor-not-allowed' : ''}`}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className={`h-5 w-5 ${liked ? 'fill-current' : ''}`}
              fill={liked ? 'currentColor' : 'none'}
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
              />
            </svg>
            <span className="font-medium">{likeCount}</span>
          </button>

          <button
            onClick={handleShowLikes}
            disabled={isLoadingLikes}
            className="ml-4 text-gray-500 hover:text-primary-600 text-sm font-medium"
          >
            {likeCount > 0 ? `${likeCount} ${likeCount === 1 ? 'like' : 'likes'}` : 'No likes yet'}
          </button>
        </div>

        {showLikes && (
          <div className="mt-3 p-3 bg-gray-50 rounded-lg">
            <p className="text-sm font-medium text-gray-700 mb-2">
              {likesList.length} {likesList.length === 1 ? 'Like' : 'Likes'}
            </p>
            {isLoadingLikes ? (
              <p className="text-sm text-gray-500">Loading...</p>
            ) : (
              <div className="space-y-2">
                {likesList.map((like) => (
                  <div key={like.user_id} className="flex items-center text-sm">
                    <div className="h-6 w-6 rounded-full bg-primary-100 flex items-center justify-center text-primary-700 text-xs font-semibold mr-2">
                      {like.username.charAt(0).toUpperCase()}
                    </div>
                    <span className="text-gray-700">{like.username}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default PostCard;
