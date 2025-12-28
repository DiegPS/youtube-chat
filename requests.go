package youtubechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DiegPS/youtube-chat-go/types"
)

var (
	clientName     = "WEB"
	BaseURL        = "https://www.youtube.com/youtubei/v1/live_chat/get_live_chat"
	YoutubeBaseURL = "https://www.youtube.com"
)

func FetchChat(options types.FetchOptions) ([]types.ChatItem, string, error) {
	url := fmt.Sprintf("%s?key=%s", BaseURL, options.ApiKey)

	payload := map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]string{
				"clientVersion": options.ClientVersion,
				"clientName":    clientName,
			},
		},
		"continuation": options.Continuation,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch chat: status %d", resp.StatusCode)
	}

	var parsedResponse types.GetLiveChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResponse); err != nil {
		return nil, "", err
	}

	items, continuation := ParseChatData(parsedResponse)
	return items, continuation, nil
}

func FetchLivePage(id types.YoutubeId) (types.FetchOptions, error) {
	url := generateLiveUrl(id)
	if url == "" {
		return types.FetchOptions{}, fmt.Errorf("id not found")
	}

	// Axios user-agent mimicry might be needed? Usually YouTube needs a User-Agent.
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.FetchOptions{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return types.FetchOptions{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.FetchOptions{}, fmt.Errorf("failed to fetch live page: status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.FetchOptions{}, err
	}

	return GetOptionsFromLivePage(string(bodyBytes))
}

func generateLiveUrl(id types.YoutubeId) string {
	if id.ChannelID != "" {
		return fmt.Sprintf("%s/channel/%s/live", YoutubeBaseURL, id.ChannelID)
	} else if id.LiveID != "" {
		return fmt.Sprintf("%s/watch?v=%s", YoutubeBaseURL, id.LiveID)
	} else if id.Handle != "" {
		handle := id.Handle
		if !strings.HasPrefix(handle, "@") {
			handle = "@" + handle
		}
		return fmt.Sprintf("%s/%s/live", YoutubeBaseURL, handle)
	}
	return ""
}
