package youtubechat

import (
	"errors"
	"testing"
	"time"

	"github.com/DiegPS/youtube-chat-go/types"
)

// Mock data
var mockChatItems = []types.ChatItem{
	{
		ID: "id",
		Author: types.Author{
			Name: "authorName",
			Thumbnail: &types.ImageItem{
				URL: "https://author.thumbnail.url",
				Alt: "authorName",
			},
			ChannelID: "channelId",
		},
		Message: []types.MessageItem{
			{Text: "Hello, World!"},
		},
		Timestamp: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

var mockOptions = types.FetchOptions{
	LiveID:        "liveId",
	ApiKey:        "apiKey",
	ClientVersion: "clientVersion",
	Continuation:  "continuation",
}

func TestConstructor(t *testing.T) {
	t.Run("LiveID", func(t *testing.T) {
		lc, err := NewLiveChat(types.YoutubeId{LiveID: "liveId"}, 1000)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if lc == nil {
			t.Error("Expected LiveChat instance")
		}
	})

	t.Run("ChannelID", func(t *testing.T) {
		lc, err := NewLiveChat(types.YoutubeId{ChannelID: "channelId"}, 1000)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if lc == nil {
			t.Error("Expected LiveChat instance")
		}
	})

	t.Run("No IDs Error", func(t *testing.T) {
		lc, err := NewLiveChat(types.YoutubeId{}, 1000)
		if err == nil {
			t.Error("Expected error for empty ID")
		}
		if lc != nil {
			t.Error("Expected nil LiveChat")
		}
	})
}

func TestStart(t *testing.T) {
	lc, _ := NewLiveChat(types.YoutubeId{ChannelID: "channelId"}, 100)

	// Mock fetchers
	lc.FetchLivePageFunc = func(id types.YoutubeId) (types.FetchOptions, error) {
		return mockOptions, nil
	}
	lc.FetchChatFunc = func(opts types.FetchOptions) ([]types.ChatItem, string, error) {
		return []types.ChatItem{}, "continuation", nil
	}

	if !lc.Start() {
		t.Error("Failed to start")
	}

	select {
	case id := <-lc.StartChan:
		if id != "liveId" {
			t.Errorf("Expected start event with liveId, got %s", id)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StartChan")
	}

	lc.Stop("test")
}

func TestStartSecondTime(t *testing.T) {
	lc, _ := NewLiveChat(types.YoutubeId{ChannelID: "channelId"}, 100)
	lc.FetchLivePageFunc = func(id types.YoutubeId) (types.FetchOptions, error) { return mockOptions, nil }
	lc.FetchChatFunc = func(opts types.FetchOptions) ([]types.ChatItem, string, error) { return nil, "", nil }

	lc.Start()
	if lc.Start() {
		t.Error("Should not allow start second time")
	}
	lc.Stop("stop")

	// Wait a bit for stop to propagate logic?
	// actually Stop() logic is sync regarding setting 'running=false'

	if !lc.Start() {
		t.Error("Should allow start after stop")
	}
	lc.Stop("stop")
}

func TestStop(t *testing.T) {
	lc, _ := NewLiveChat(types.YoutubeId{ChannelID: "channelId"}, 100)
	lc.FetchLivePageFunc = func(id types.YoutubeId) (types.FetchOptions, error) { return mockOptions, nil }
	lc.FetchChatFunc = func(opts types.FetchOptions) ([]types.ChatItem, string, error) { return nil, "", nil }

	lc.Start()
	// Drain start chan
	<-lc.StartChan

	lc.Stop("STOP")

	select {
	case reason := <-lc.EndChan:
		if reason != "STOP" {
			t.Errorf("Expected EndChan reason STOP, got %s", reason)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for EndChan")
	}
}

func TestOnChat(t *testing.T) {
	lc, _ := NewLiveChat(types.YoutubeId{ChannelID: "channelId"}, 50) // fast interval
	lc.FetchLivePageFunc = func(id types.YoutubeId) (types.FetchOptions, error) { return mockOptions, nil }

	// Mock FetchChat to return items once
	called := false
	lc.FetchChatFunc = func(opts types.FetchOptions) ([]types.ChatItem, string, error) {
		if !called {
			called = true
			return mockChatItems, "continuation", nil
		}
		return []types.ChatItem{}, "continuation", nil
	}

	lc.Start()
	defer lc.Stop("done")

	select {
	case chat := <-lc.ChatChan:
		if chat.ID != "id" {
			t.Errorf("Expected chat ID 'id', got %s", chat.ID)
		}
		if chat.Message[0].Text != "Hello, World!" {
			t.Errorf("Message mismatch")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for ChatChan")
	}
}

func TestOnError_FetchLivePage(t *testing.T) {
	lc, _ := NewLiveChat(types.YoutubeId{ChannelID: "channelId"}, 100)
	lc.FetchLivePageFunc = func(id types.YoutubeId) (types.FetchOptions, error) {
		return types.FetchOptions{}, errors.New("ERROR")
	}

	if lc.Start() {
		t.Error("Start should return false on error")
	}

	select {
	case err := <-lc.ErrorChan:
		if err.Error() != "ERROR" {
			t.Errorf("Expected error 'ERROR', got %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for ErrorChan")
	}
}

func TestOnError_FetchChat(t *testing.T) {
	lc, _ := NewLiveChat(types.YoutubeId{ChannelID: "channelId"}, 50)
	lc.FetchLivePageFunc = func(id types.YoutubeId) (types.FetchOptions, error) { return mockOptions, nil }

	lc.FetchChatFunc = func(opts types.FetchOptions) ([]types.ChatItem, string, error) {
		return nil, "", errors.New("ERROR")
	}

	lc.Start()
	defer lc.Stop("done")

	select {
	case err := <-lc.ErrorChan:
		if err.Error() != "ERROR" {
			t.Errorf("Expected error 'ERROR', got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for ErrorChan")
	}
}
