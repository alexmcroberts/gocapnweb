<script>
  import { newHttpBatchRpcSession } from 'capnweb';

  let handle = $state('bsky.app');
  let loading = $state(false);
  let error = $state(null);
  let profile = $state(null);
  let feed = $state(null);
  let mode = $state('batched'); // 'batched' or 'sequential'
  let timingInfo = $state(null);

  async function loadProfileAndFeed() {
    loading = true;
    error = null;
    profile = null;
    feed = null;
    timingInfo = null;

    try {
      const startTime = performance.now();

      if (mode === 'batched') {
        // Batched mode: both calls in a single HTTP request with pipelining
        // Create a new session for this batch
        const api = newHttpBatchRpcSession('http://localhost:3000/rpc');
        
        try {
          const profilePromise = api.getProfile(handle);
          const feedPromise = api.getFeed(handle, 10);
          
          const [profileResult, feedResult] = await Promise.all([
            profilePromise,
            feedPromise,
          ]);
          
          const endTime = performance.now();
          
          console.log('Batched results:', { profileResult, feedResult });
          
          profile = profileResult;
          // Unwrap the double-wrapped posts array (Cap'n Web escaping)
          if (feedResult.posts && Array.isArray(feedResult.posts) && feedResult.posts.length > 0 && Array.isArray(feedResult.posts[0])) {
            feedResult.posts = feedResult.posts[0];
          }
          feed = feedResult;
          timingInfo = {
            mode: 'Batched (Pipeline)',
            duration: Math.round(endTime - startTime),
            requests: 1,
            description: 'Both profile and feed fetched in a single HTTP request using Cap\'n Web RPC pipelining'
          };
        } catch (batchError) {
          console.error('Batched mode error:', batchError);
          throw batchError;
        }
      } else {
        // Sequential mode: two separate HTTP requests
        // Create a new session for each request
        const start1 = performance.now();
        const api1 = newHttpBatchRpcSession('http://localhost:3000/rpc');
        profile = await api1.getProfile(handle);
        const end1 = performance.now();
        
        const start2 = performance.now();
        const api2 = newHttpBatchRpcSession('http://localhost:3000/rpc');
        feed = await api2.getFeed(handle, 10);
        // Unwrap the double-wrapped posts array (Cap'n Web escaping)
        if (feed.posts && Array.isArray(feed.posts) && feed.posts.length > 0 && Array.isArray(feed.posts[0])) {
          feed.posts = feed.posts[0];
        }
        const end2 = performance.now();
        
        const totalDuration = Math.round(end2 - start1);
        
        timingInfo = {
          mode: 'Sequential',
          duration: totalDuration,
          requests: 2,
          description: 'Profile and feed fetched in two separate HTTP requests',
          request1Duration: Math.round(end1 - start1),
          request2Duration: Math.round(end2 - start2),
        };
      }
    } catch (err) {
      error = err.message || 'Failed to load profile and feed';
      console.error('Error:', err);
    } finally {
      loading = false;
    }
  }

  function formatDate(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    
    return date.toLocaleDateString();
  }
</script>

<h1>🦋 Bluesky Feed Reader</h1>

<div class="container">
  <div class="info-box">
    <h3>Cap'n Web RPC + AT Protocol</h3>
    <p>
      This demo showcases <strong>Cap'n Web RPC</strong> batch pipelining by fetching 
      Bluesky profiles and feeds via <strong>AT Protocol's XRPC</strong> endpoints.
    </p>
    <p>
      Try switching between <strong>Batched</strong> and <strong>Sequential</strong> modes 
      to see how pipelining reduces round trips and improves performance!
    </p>
  </div>

  <div class="input-section">
    <div class="mode-toggle">
      <span><strong>Mode:</strong></span>
      <label>
        <input type="radio" bind:group={mode} value="batched" aria-label="Batched (Pipeline) Mode" />
        Batched (Pipeline)
      </label>
      <label>
        <input type="radio" bind:group={mode} value="sequential" aria-label="Sequential Mode" />
        Sequential
      </label>
    </div>

    <div class="input-group">
      <input
        type="text"
        bind:value={handle}
        placeholder="Enter Bluesky handle (e.g., bsky.app)"
        onkeydown={(e) => e.key === 'Enter' && !loading && loadProfileAndFeed()}
        disabled={loading}
        aria-label="Bluesky handle"
      />
      <button onclick={loadProfileAndFeed} disabled={loading} aria-label="Load Profile & Feed">
        {loading ? 'Loading...' : 'Load Profile & Feed'}
      </button>
    </div>
  </div>

  {#if error}
    <div class="error">
      <strong>Error:</strong> {error}
    </div>
  {/if}

  {#if loading}
    <div class="loading">
      <div>Loading profile and feed...</div>
    </div>
  {/if}

  {#if timingInfo}
    <div class="timing-info">
      <h4>⚡ Performance: {timingInfo.mode}</h4>
      <p><strong>Total Duration:</strong> {timingInfo.duration}ms</p>
      <p><strong>HTTP Requests:</strong> {timingInfo.requests}</p>
      <p>{timingInfo.description}</p>
      {#if timingInfo.request1Duration}
        <p style="margin-top: 12px;">
          <strong>Request 1 (Profile):</strong> {timingInfo.request1Duration}ms<br/>
          <strong>Request 2 (Feed):</strong> {timingInfo.request2Duration}ms
        </p>
      {/if}
    </div>
  {/if}

  {#if profile}
    <div class="profile-section">
      <div class="profile-header">
        {#if profile.avatar}
          <img src={profile.avatar} alt={profile.displayName || profile.handle} class="profile-avatar" />
        {/if}
        <div class="profile-info">
          <h2>{profile.displayName || profile.handle}</h2>
          <div class="profile-handle">@{profile.handle}</div>
          <div class="profile-stats">
            <div class="stat">
              <div class="stat-value">{profile.postsCount || 0}</div>
              <div class="stat-label">Posts</div>
            </div>
            <div class="stat">
              <div class="stat-value">{profile.followersCount || 0}</div>
              <div class="stat-label">Followers</div>
            </div>
            <div class="stat">
              <div class="stat-value">{profile.followsCount || 0}</div>
              <div class="stat-label">Following</div>
            </div>
          </div>
        </div>
      </div>
      {#if profile.description}
        <div class="profile-bio">{profile.description}</div>
      {/if}
    </div>
  {/if}

  {#if feed && feed.posts}
    <div class="feed-section">
      <h3>Recent Posts ({feed.posts.length})</h3>
      {#each feed.posts as post}
        <div class="post">
          <div class="post-author">
            {#if post.author.avatar}
              <img src={post.author.avatar} alt={post.author.handle} class="post-avatar" />
            {/if}
            <div class="post-author-info">
              <div class="post-author-name">{post.author.displayName || post.author.handle}</div>
              <div class="post-author-handle">@{post.author.handle}</div>
            </div>
          </div>
          {#if post.record.text}
            <div class="post-text">{post.record.text}</div>
          {/if}
          <div class="post-stats">
            <div class="post-stat">💬 {post.replyCount || 0}</div>
            <div class="post-stat">🔄 {post.repostCount || 0}</div>
            <div class="post-stat">❤️ {post.likeCount || 0}</div>
          </div>
          {#if post.indexedAt}
            <div class="post-time">{formatDate(post.indexedAt)}</div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
.container {
  background: white;
  border-radius: 12px;
  padding: 30px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.input-section {
  margin-bottom: 30px;
}

.input-group {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

input[type="text"] {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 0.3s;
}

input[type="text"]:focus {
  outline: none;
  border-color: #667eea;
}

button {
  padding: 12px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s, opacity 0.3s;
}

button:hover:not(:disabled) {
  transform: translateY(-2px);
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.mode-toggle {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 15px;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 8px;
}

.mode-toggle label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
}

.mode-toggle input[type="radio"] {
  cursor: pointer;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #666;
  font-size: 18px;
}

.error {
  padding: 15px;
  background: #fee;
  color: #c33;
  border-radius: 8px;
  margin-bottom: 20px;
  border-left: 4px solid #c33;
}

.profile-section {
  margin-bottom: 30px;
  padding: 25px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 12px;
}

.profile-header {
  display: flex;
  gap: 20px;
  align-items: start;
  margin-bottom: 20px;
}

.profile-avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  object-fit: cover;
  border: 4px solid white;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.profile-info h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: #333;
}

.profile-handle {
  color: #667eea;
  font-size: 16px;
  margin-bottom: 12px;
}

.profile-stats {
  display: flex;
  gap: 20px;
  margin-top: 15px;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 20px;
  background: white;
  border-radius: 8px;
  min-width: 80px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #667eea;
}

.stat-label {
  font-size: 12px;
  color: #666;
  text-transform: uppercase;
  margin-top: 4px;
}

.profile-bio {
  margin-top: 15px;
  padding: 15px;
  background: white;
  border-radius: 8px;
  color: #555;
  line-height: 1.6;
}

.feed-section h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #333;
  font-size: 22px;
}

.post {
  padding: 20px;
  background: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 15px;
  border-left: 4px solid #667eea;
}

.post-author {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.post-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.post-author-info {
  flex: 1;
}

.post-author-name {
  font-weight: 600;
  color: #333;
  font-size: 15px;
}

.post-author-handle {
  color: #888;
  font-size: 13px;
}

.post-text {
  margin-bottom: 12px;
  line-height: 1.6;
  color: #333;
  white-space: pre-wrap;
}

.post-stats {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #666;
}

.post-stat {
  display: flex;
  align-items: center;
  gap: 6px;
}

.post-time {
  font-size: 13px;
  color: #999;
  margin-top: 8px;
}

.timing-info {
  margin: 20px 0;
  padding: 15px;
  background: #e8f5e9;
  border-radius: 8px;
  border-left: 4px solid #4caf50;
}

.timing-info h4 {
  margin: 0 0 8px 0;
  color: #2e7d32;
  font-size: 16px;
}

.timing-info p {
  margin: 4px 0;
  color: #555;
  font-size: 14px;
}

.info-box {
  padding: 20px;
  background: #f0f7ff;
  border-radius: 8px;
  border-left: 4px solid #2196f3;
  margin-bottom: 20px;
}

.info-box h3 {
  margin: 0 0 10px 0;
  color: #1565c0;
  font-size: 18px;
}

.info-box p {
  margin: 8px 0;
  color: #555;
  line-height: 1.6;
}


</style>