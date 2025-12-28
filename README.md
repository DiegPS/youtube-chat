# youtube-chat-go

A port of [youtube-chat](https://github.com/LinaTsukusu/youtube-chat) to Go.

## 1. Install
```bash
go get github.com/DiegPS/youtube-chat-go
```

## 2. Import
```go
import (
    "github.com/DiegPS/youtube-chat-go"
    "github.com/DiegPS/youtube-chat-go/types" // optional, for more granular access to types
)
```

## 3. Create instance with ChannelID or LiveID
```go
// If channelId is specified, liveId in the current stream is automatically acquired.
// Recommended
lc, err := youtubechat.NewLiveChat(types.YoutubeId{ChannelID: "CHANNEL_ID_HERE"}, 1000)

// Or specify LiveID (Video ID) manually.
lc, err := youtubechat.NewLiveChat(types.YoutubeId{LiveID: "LIVE_ID_HERE"}, 1000)
```

## 4. Handle events
In Go, instead of an `EventEmitter`, events are handled through channels for type safety and idiomatic concurrency.

```go
// Start observations
ok := lc.Start()
if !ok {
    fmt.Println("Failed to start")
}

for {
    select {
    case liveId := <-lc.StartChan:
        // Emit at start of observation chat.
        fmt.Printf("Started observing: %s\n", liveId)

    case reason := <-lc.EndChan:
        // Emit at end of observation chat.
        fmt.Printf("Ended: %s\n", reason)
        return

    case chatItem := <-lc.ChatChan:
        // Emit at receive chat.
        // chatItem fields match ChatItem interface in JS
        fmt.Printf("[%s]: %v\n", chatItem.Author.Name, chatItem.Message)

    case err := <-lc.ErrorChan:
        // Emit when an error occurs
        fmt.Printf("Error: %v\n", err)
    }
}
```

## 5. Stop loop
```go
lc.Stop("optional manual stop reason")
```

## Types

### ChatItem
```go
type ChatItem struct {
	ID           string
	Author       Author
	Message      []MessageItem
	SuperChat    *SuperChat
	IsMembership bool
	IsVerified   bool
	IsOwner      bool
	IsModerator  bool
	Timestamp    time.Time
}
```

### MessageItem
```go
// MessageItem represents a chat message text or emoji
type MessageItem struct {
	Text      string
	EmojiItem *EmojiItem
}
```

### ImageItem
```go
type ImageItem struct {
	URL string
	Alt string
}
```

### EmojiItem
```go
type EmojiItem struct {
	ImageItem
	EmojiText     string
	IsCustomEmoji bool
}
```
