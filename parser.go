package youtubechat

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DiegPS/youtube-chat-go/types"
)

var (
	regexCanonical    = regexp.MustCompile(`<link rel="canonical" href="https:\/\/www.youtube.com\/watch\?v=(.+?)">`)
	regexIsReplay     = regexp.MustCompile(`['"]isReplay['"]:\s*(true)`)
	regexAPIKey       = regexp.MustCompile(`['"]INNERTUBE_API_KEY['"]:\s*['"](.+?)['"]`)
	regexClientVer    = regexp.MustCompile(`['"]clientVersion['"]:\s*['"]([\d.]+?)['"]`)
	regexContinuation = regexp.MustCompile(`['"]continuation['"]:\s*['"](.+?)['"]`)
)

func GetOptionsFromLivePage(data string) (types.FetchOptions, error) {
	var opts types.FetchOptions

	// LiveID
	liveIDMatch := regexCanonical.FindStringSubmatch(data)
	if len(liveIDMatch) > 1 {
		opts.LiveID = liveIDMatch[1]
	} else {
		return opts, errors.New("Live Stream was not found")
	}

	// Replay
	if regexIsReplay.MatchString(data) {
		return opts, fmt.Errorf("%s is finished live", opts.LiveID)
	}

	// API Key
	apiKeyMatch := regexAPIKey.FindStringSubmatch(data)
	if len(apiKeyMatch) > 1 {
		opts.ApiKey = apiKeyMatch[1]
	} else {
		return opts, errors.New("API Key was not found")
	}

	// Client Version
	clientVerMatch := regexClientVer.FindStringSubmatch(data)
	if len(clientVerMatch) > 1 {
		opts.ClientVersion = clientVerMatch[1]
	} else {
		return opts, errors.New("Client Version was not found")
	}

	// Continuation
	continuationMatch := regexContinuation.FindStringSubmatch(data)
	if len(continuationMatch) > 1 {
		opts.Continuation = continuationMatch[1]
	} else {
		return opts, errors.New("Continuation was not found")
	}

	return opts, nil
}

func ParseChatData(data types.GetLiveChatResponse) ([]types.ChatItem, string) {
	var chatItems []types.ChatItem

	if data.ContinuationContents.LiveChatContinuation.Actions != nil {
		for _, action := range data.ContinuationContents.LiveChatContinuation.Actions {
			item := parseActionToChatItem(action)
			if item != nil {
				chatItems = append(chatItems, *item)
			}
		}
	}

	continuation := ""
	if len(data.ContinuationContents.LiveChatContinuation.Continuations) > 0 {
		contData := data.ContinuationContents.LiveChatContinuation.Continuations[0]
		if contData.InvalidationContinuationData != nil {
			continuation = contData.InvalidationContinuationData.Continuation
		} else if contData.TimedContinuationData != nil {
			continuation = contData.TimedContinuationData.Continuation
		}
	}

	return chatItems, continuation
}

func parseThumbnailToImageItem(data []types.Thumbnail, alt string) *types.ImageItem {
	if len(data) == 0 {
		return &types.ImageItem{URL: "", Alt: ""}
	}
	// ts: data.pop() -> last element
	formattedThumb := data[len(data)-1]
	return &types.ImageItem{
		URL: formattedThumb.URL,
		Alt: alt,
	}
}

func convertColorToHex6(colorNum int) string {
	// hex string from int, strip alpha?
	// TS: colorNum.toString(16).slice(2).toLocaleUpperCase()
	// Go: fmt.Sprintf("%X", colorNum)
	hex := fmt.Sprintf("%08X", colorNum) // assumes ARGB or RGBA 32bit int?
	// TS slice(2) implies skipping first 2 chars (Alpha).
	if len(hex) >= 2 {
		return "#" + strings.ToUpper(hex[2:])
	}
	return "#" + hex
}

func parseMessages(runs []types.MessageRun) []types.MessageItem {
	var items []types.MessageItem
	for _, run := range runs {
		if run.Text != "" {
			items = append(items, types.MessageItem{Text: run.Text})
		} else if run.Emoji != nil {
			// Emoji
			var thumb *types.ImageItem
			if len(run.Emoji.Image.Thumbnails) > 0 {
				// shift() implies first element
				firstThumb := run.Emoji.Image.Thumbnails[0]
				thumb = &types.ImageItem{URL: firstThumb.URL}
			} else {
				thumb = &types.ImageItem{URL: "", Alt: ""}
			}

			isCustom := run.Emoji.IsCustomEmoji
			shortcut := ""
			if len(run.Emoji.Shortcuts) > 0 {
				shortcut = run.Emoji.Shortcuts[0]
			}

			// Update thumb alt
			thumb.Alt = shortcut

			items = append(items, types.MessageItem{
				EmojiItem: &types.EmojiItem{
					ImageItem: *thumb,
					EmojiText: func() string {
						if isCustom {
							return shortcut
						} else {
							return run.Emoji.EmojiId
						}
					}(),
					IsCustomEmoji: isCustom,
				},
			})
		}
	}
	return items
}

func parseActionToChatItem(data types.Action) *types.ChatItem {
	if data.AddChatItemAction == nil {
		return nil
	}
	item := data.AddChatItemAction.Item

	var messageRenderer *types.MessageRendererBase
	// Identifying renderer
	if item.LiveChatTextMessageRenderer != nil {
		messageRenderer = &item.LiveChatTextMessageRenderer.MessageRendererBase
	} else if item.LiveChatPaidMessageRenderer != nil {
		messageRenderer = &item.LiveChatPaidMessageRenderer.MessageRendererBase
	} else if item.LiveChatPaidStickerRenderer != nil {
		messageRenderer = &item.LiveChatPaidStickerRenderer.MessageRendererBase
	} else if item.LiveChatMembershipItemRenderer != nil {
		messageRenderer = &item.LiveChatMembershipItemRenderer.MessageRendererBase
	} else {
		return nil
	}

	var messageRuns []types.MessageRun
	if item.LiveChatTextMessageRenderer != nil {
		messageRuns = item.LiveChatTextMessageRenderer.Message.Runs
	} else if item.LiveChatMembershipItemRenderer != nil {
		messageRuns = item.LiveChatMembershipItemRenderer.HeaderSubtext.Runs
	} else if item.LiveChatPaidMessageRenderer != nil {
		// Paid message also has message
		messageRuns = item.LiveChatPaidMessageRenderer.Message.Runs
	}

	authorNameText := ""
	if messageRenderer.AuthorName != nil {
		authorNameText = messageRenderer.AuthorName.SimpleText
	}

	// Author thumbnails
	authorThumb := parseThumbnailToImageItem(messageRenderer.AuthorPhoto.Thumbnails, authorNameText)

	timestamp := time.Now()
	if ts, err := strconv.ParseInt(messageRenderer.TimestampUsec, 10, 64); err == nil {
		timestamp = time.Unix(ts/1000000, (ts%1000000)*1000)
	}

	idx := types.ChatItem{
		ID: messageRenderer.ID,
		Author: types.Author{
			Name:      authorNameText,
			Thumbnail: authorThumb,
			ChannelID: messageRenderer.AuthorExternalChannelId,
		},
		Message:      parseMessages(messageRuns),
		Timestamp:    timestamp,
		IsMembership: false,
		IsOwner:      false,
		IsVerified:   false,
		IsModerator:  false,
	}

	if messageRenderer.AuthorBadges != nil {
		for _, entry := range messageRenderer.AuthorBadges {
			badge := entry.LiveChatAuthorBadgeRenderer
			if badge.CustomThumbnail != nil {
				idx.Author.Badge = &types.Badge{
					Thumbnail: *parseThumbnailToImageItem(badge.CustomThumbnail.Thumbnails, badge.Tooltip),
					Label:     badge.Tooltip,
				}
				idx.IsMembership = true
			} else {
				if badge.Icon != nil {
					switch badge.Icon.IconType {
					case "OWNER":
						idx.IsOwner = true
					case "VERIFIED":
						idx.IsVerified = true
					case "MODERATOR":
						idx.IsModerator = true
					}
				}
			}
		}
	}

	// Superchat logic
	if item.LiveChatPaidStickerRenderer != nil {
		r := item.LiveChatPaidStickerRenderer
		idx.SuperChat = &types.SuperChat{
			Amount: r.PurchaseAmountText.SimpleText,
			Color:  convertColorToHex6(r.BackgroundColor),
			Sticker: parseThumbnailToImageItem(
				r.Sticker.Thumbnails,
				r.Sticker.Accessibility.AccessibilityData.Label,
			),
		}
	} else if item.LiveChatPaidMessageRenderer != nil {
		r := item.LiveChatPaidMessageRenderer
		idx.SuperChat = &types.SuperChat{
			Amount: r.PurchaseAmountText.SimpleText,
			Color:  convertColorToHex6(r.BodyBackgroundColor),
		}
	}

	return &idx
}
