package types

import "time"

// ChatItem represents the detailed chat info
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

type Author struct {
	Name      string
	Thumbnail *ImageItem
	ChannelID string
	Badge     *Badge
}

type Badge struct {
	Thumbnail ImageItem
	Label     string
}

type SuperChat struct {
	Amount  string
	Color   string
	Sticker *ImageItem
}

// MessageItem represents a chat message string or emoji
// TypeScript: export type MessageItem = { text: string } | EmojiItem
type MessageItem struct {
	Text      string
	EmojiItem *EmojiItem
}

// ImageItem represents an image
type ImageItem struct {
	URL string
	Alt string
}

// EmojiItem represents an emoji
type EmojiItem struct {
	ImageItem
	EmojiText     string
	IsCustomEmoji bool
}

// YoutubeId union type in TS: { channelId: string } | { liveId: string } | { handle: string }
// In Go we can use a struct with optional fields
type YoutubeId struct {
	ChannelID string
	LiveID    string
	Handle    string
}
