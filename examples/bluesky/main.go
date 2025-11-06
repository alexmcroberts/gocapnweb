package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gocapnweb"
)

const blueskyAPIBase = "https://public.api.bsky.app/xrpc"

// sanitizeJSON recursively removes or renames keys starting with "$" to avoid
// conflicts with Cap'n Web RPC's special value handling
func sanitizeJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			// Rename keys starting with "$" to avoid Cap'n Web protocol conflicts
			newKey := key
			if len(key) > 0 && key[0] == '$' {
				newKey = "_" + key[1:] // Replace $ with _ (e.g., $type -> _type)
			}
			result[newKey] = sanitizeJSON(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = sanitizeJSON(item)
		}
		return result
	default:
		return v
	}
}

// BlueskyProfile represents a Bluesky profile response
type BlueskyProfile struct {
	DID            string `json:"did"`
	Handle         string `json:"handle"`
	DisplayName    string `json:"displayName,omitempty"`
	Description    string `json:"description,omitempty"`
	Avatar         string `json:"avatar,omitempty"`
	Banner         string `json:"banner,omitempty"`
	FollowersCount int    `json:"followersCount"`
	FollowsCount   int    `json:"followsCount"`
	PostsCount     int    `json:"postsCount"`
}

// BlueskyPost represents a single post in a feed
type BlueskyPost struct {
	URI         string                 `json:"uri"`
	CID         string                 `json:"cid"`
	Author      BlueskyPostAuthor      `json:"author"`
	Record      map[string]interface{} `json:"record"`
	ReplyCount  int                    `json:"replyCount,omitempty"`
	RepostCount int                    `json:"repostCount,omitempty"`
	LikeCount   int                    `json:"likeCount,omitempty"`
	IndexedAt   string                 `json:"indexedAt"`
}

// BlueskyPostAuthor represents the author of a post
type BlueskyPostAuthor struct {
	DID         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

// BlueskyFeedResponse represents the feed response structure
type BlueskyFeedResponse struct {
	Feed []struct {
		Post BlueskyPost `json:"post"`
	} `json:"feed"`
	Cursor string `json:"cursor,omitempty"`
}

// BlueskyServer implements RPC methods for fetching Bluesky data
type BlueskyServer struct {
	*gocapnweb.BaseRpcTarget
	httpClient *http.Client
}

// NewBlueskyServer creates a new BlueskyServer instance
func NewBlueskyServer() *BlueskyServer {
	server := &BlueskyServer{
		BaseRpcTarget: gocapnweb.NewBaseRpcTarget(),
		httpClient:    &http.Client{},
	}

	// Register RPC methods
	server.Method("getProfile", server.getProfile)
	server.Method("getFeed", server.getFeed)

	return server
}

// getProfile fetches a Bluesky profile by handle
func (s *BlueskyServer) getProfile(args json.RawMessage) (interface{}, error) {
	// Extract handle from arguments
	var handle string

	// Try to parse as array first
	var argArray []string
	if err := json.Unmarshal(args, &argArray); err == nil && len(argArray) > 0 {
		handle = argArray[0]
	} else {
		// Try to parse as string
		if err := json.Unmarshal(args, &handle); err != nil {
			return nil, fmt.Errorf("invalid arguments: expected handle")
		}
	}

	if handle == "" {
		return nil, fmt.Errorf("handle is required")
	}

	// Build API URL
	apiURL := fmt.Sprintf("%s/app.bsky.actor.getProfile?actor=%s", blueskyAPIBase, url.QueryEscape(handle))

	log.Printf("Fetching profile for handle: %s", handle)

	// Make API request
	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var profile BlueskyProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	log.Printf("Successfully fetched profile for %s (DID: %s)", profile.Handle, profile.DID)

	return profile, nil
}

// getFeed fetches a user's feed by handle
func (s *BlueskyServer) getFeed(args json.RawMessage) (interface{}, error) {
	// Extract arguments
	var argArray []interface{}
	if err := json.Unmarshal(args, &argArray); err != nil {
		return nil, fmt.Errorf("invalid arguments: expected [handle, limit]")
	}

	if len(argArray) == 0 {
		return nil, fmt.Errorf("handle is required")
	}

	handle, ok := argArray[0].(string)
	if !ok {
		return nil, fmt.Errorf("handle must be a string")
	}

	limit := 10
	if len(argArray) > 1 {
		if limitFloat, ok := argArray[1].(float64); ok {
			limit = int(limitFloat)
		}
	}

	// Build API URL
	apiURL := fmt.Sprintf("%s/app.bsky.feed.getAuthorFeed?actor=%s&limit=%d",
		blueskyAPIBase, url.QueryEscape(handle), limit)

	log.Printf("Fetching feed for handle: %s (limit: %d)", handle, limit)

	// Make API request
	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse as generic JSON first, then sanitize to avoid $ key conflicts
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	// Sanitize the entire response
	sanitized := sanitizeJSON(rawResponse).(map[string]interface{})

	// Extract feed array
	feedArray, ok := sanitized["feed"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected feed response format")
	}

	// Extract and simplify posts - only include fields we actually display
	// Use []interface{} instead of []map[string]interface{} for Cap'n Web compatibility
	posts := make([]interface{}, 0, len(feedArray))
	for _, item := range feedArray {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if postData, ok := itemMap["post"].(map[string]interface{}); ok {
				// Extract only the fields we need for display
				simplifiedPost := map[string]interface{}{
					"uri":         postData["uri"],
					"cid":         postData["cid"],
					"indexedAt":   postData["indexedAt"],
					"replyCount":  getIntOrZero(postData, "replyCount"),
					"repostCount": getIntOrZero(postData, "repostCount"),
					"likeCount":   getIntOrZero(postData, "likeCount"),
				}

				// Extract author info
				if author, ok := postData["author"].(map[string]interface{}); ok {
					simplifiedPost["author"] = map[string]interface{}{
						"did":         author["did"],
						"handle":      author["handle"],
						"displayName": author["displayName"],
						"avatar":      author["avatar"],
					}
				}

				// Extract record.text
				if record, ok := postData["record"].(map[string]interface{}); ok {
					simplifiedPost["record"] = map[string]interface{}{
						"text": record["text"],
					}
				}

				posts = append(posts, simplifiedPost)
			}
		}
	}

	log.Printf("Successfully fetched %d posts for %s", len(posts), handle)

	cursor := ""
	if c, ok := sanitized["cursor"].(string); ok {
		cursor = c
	}

	// Wrap the posts array in another array to escape it for Cap'n Web
	result := map[string]interface{}{
		"posts":  []interface{}{posts}, // Double-wrap: [[{...}]]
		"cursor": cursor,
	}

	return result, nil
}

// Helper to safely extract int values with zero default
func getIntOrZero(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return 0
}

func main() {
	// Default to serving static files from the examples/static directory
	staticPath := "/static"
	if len(os.Args) >= 2 {
		staticPath = os.Args[1]
	}

	port := ":8000"

	// Create Echo server with middleware
	e := gocapnweb.SetupEchoServer()

	// Setup RPC endpoint
	server := NewBlueskyServer()
	gocapnweb.SetupRpcEndpoint(e, "/rpc", server)

	// Setup static file endpoint
	gocapnweb.SetupFileEndpoint(e, "/static", staticPath)

	log.Printf("🚀 Bluesky Feed Reader Go Server starting on port %s", port)
	log.Printf("🔌 HTTP Batch RPC endpoint: http://localhost%s/rpc", port)
	log.Printf("🌐 Demo URL: http://localhost:3000 (available once you start the Svelte development server)")
	log.Println()
	log.Println("Features:")
	log.Println("  🦋 Fetch Bluesky profiles and feeds")
	log.Println("  ⚡ Batch pipelining for optimal performance")
	log.Println("  🔗 AT Protocol/XRPC integration")
	log.Println()
	log.Println("Try it with curl:")
	log.Printf("  curl -X POST http://localhost%s/rpc -d '[\"push\",[\"pipeline\",1,[\"getProfile\"],[\"bsky.app\"]]]'", port)
	log.Printf("  curl -X POST http://localhost%s/rpc -d '[\"pull\",1]'", port)

	if err := e.Start(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
