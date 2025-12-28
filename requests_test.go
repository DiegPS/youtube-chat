package youtubechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DiegPS/youtube-chat-go/types"
)

func TestFetchChat(t *testing.T) {
	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		key := r.URL.Query().Get("key")
		if key != "apiKey" {
			t.Errorf("Expected key 'apiKey', got '%s'", key)
		}

		// Verify Body
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		ctx, _ := payload["context"].(map[string]interface{})
		client, _ := ctx["client"].(map[string]interface{})

		if client["clientVersion"] != "clientVersion" {
			t.Errorf("Expected clientVersion 'clientVersion'")
		}
		if client["clientName"] != "WEB" {
			t.Errorf("Expected clientName 'WEB'")
		}
		if payload["continuation"] != "continuation" {
			t.Errorf("Expected continuation 'continuation'")
		}

		// Return mock response
		// We can return a minimal valid response or just check the request
		// returns a JSON that parses to empty?
		w.Header().Set("Content-Type", "application/json")
		// minimal response
		fmt.Fprintln(w, `{"continuationContents": {"liveChatContinuation": {"actions": [], "continuations": []}}}`)
	}))
	defer ts.Close()

	// Update BaseURL
	origBaseURL := BaseURL
	BaseURL = ts.URL
	defer func() { BaseURL = origBaseURL }()

	options := types.FetchOptions{
		ApiKey:        "apiKey",
		ClientVersion: "clientVersion",
		Continuation:  "continuation",
	}

	_, _, err := FetchChat(options)
	if err != nil {
		t.Errorf("FetchChat failed: %v", err)
	}
}

func TestFetchLivePage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html></html>")) // dummy response
	}))
	defer ts.Close()

	origYoutubeBaseURL := YoutubeBaseURL
	// YoutubeBaseURL is the domain, e.g. https://www.youtube.com
	// httptest URL is http://127.0.0.1:xxx
	YoutubeBaseURL = ts.URL
	defer func() { YoutubeBaseURL = origYoutubeBaseURL }()

	t.Run("ChannelID request", func(t *testing.T) {
		// We can spy on the request in a custom handler per sub-test or reuse logic
		// But here we just want to verify the URL pattern
		// Since we reuse 'ts', let's check request path in handler?
		// We can't change handler of existing server easily.
		// Let's create a new server for each sub-test or one smart handler.
	})
}

// Redefining TestFetchLivePage with specific handlers
func TestFetchLivePage_Detailed(t *testing.T) {
	origYoutubeBaseURL := YoutubeBaseURL
	defer func() { YoutubeBaseURL = origYoutubeBaseURL }()

	t.Run("ChannelID request", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/channel/channelId/live" {
				t.Errorf("Expected path /channel/channelId/live, got %s", r.URL.Path)
			}
		}))
		defer ts.Close()
		YoutubeBaseURL = ts.URL

		FetchLivePage(types.YoutubeId{ChannelID: "channelId"})
	})

	t.Run("LiveID request", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/watch" || r.URL.Query().Get("v") != "liveId" {
				t.Errorf("Expected path /watch?v=liveId, got %s?%s", r.URL.Path, r.URL.RawQuery)
			}
		}))
		defer ts.Close()
		YoutubeBaseURL = ts.URL

		FetchLivePage(types.YoutubeId{LiveID: "liveId"})
	})

	t.Run("Handle request", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/@handle/live" {
				t.Errorf("Expected path /@handle/live, got %s", r.URL.Path)
			}
		}))
		defer ts.Close()
		YoutubeBaseURL = ts.URL

		FetchLivePage(types.YoutubeId{Handle: "@handle"})
	})

	t.Run("Handle without @", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/@handle/live" {
				t.Errorf("Expected path /@handle/live, got %s", r.URL.Path)
			}
		}))
		defer ts.Close()
		YoutubeBaseURL = ts.URL

		FetchLivePage(types.YoutubeId{Handle: "handle"})
	})
}
