package types

// GetLiveChatResponse represents the API response
type GetLiveChatResponse struct {
	ResponseContext      interface{} `json:"responseContext"`
	TrackingParams       string      `json:"trackingParams,omitempty"`
	ContinuationContents struct {
		LiveChatContinuation struct {
			Continuations []Continuation `json:"continuations"`
			Actions       []Action       `json:"actions"`
		} `json:"liveChatContinuation"`
	} `json:"continuationContents"`
}

type Continuation struct {
	InvalidationContinuationData *struct {
		InvalidationId struct {
			ObjectSource             int    `json:"objectSource"`
			ObjectId                 string `json:"objectId"`
			Topic                    string `json:"topic"`
			SubscribeToGcmTopics     bool   `json:"subscribeToGcmTopics"`
			ProtoCreationTimestampMs string `json:"protoCreationTimestampMs"`
		} `json:"invalidationId"`
		TimeoutMs    int    `json:"timeoutMs"`
		Continuation string `json:"continuation"`
	} `json:"invalidationContinuationData,omitempty"`
	TimedContinuationData *struct {
		TimeoutMs           int    `json:"timeoutMs"`
		Continuation        string `json:"continuation"`
		ClickTrackingParams string `json:"clickTrackingParams"`
	} `json:"timedContinuationData,omitempty"`
}

type Action struct {
	AddChatItemAction           *AddChatItemAction `json:"addChatItemAction,omitempty"`
	AddLiveChatTickerItemAction interface{}        `json:"addLiveChatTickerItemAction,omitempty"`
}

type AddChatItemAction struct {
	Item     ActionItem `json:"item"`
	ClientId string     `json:"clientId"`
}

type ActionItem struct {
	LiveChatTextMessageRenderer             *LiveChatTextMessageRenderer    `json:"liveChatTextMessageRenderer,omitempty"`
	LiveChatPaidMessageRenderer             *LiveChatPaidMessageRenderer    `json:"liveChatPaidMessageRenderer,omitempty"`
	LiveChatMembershipItemRenderer          *LiveChatMembershipItemRenderer `json:"liveChatMembershipItemRenderer,omitempty"`
	LiveChatPaidStickerRenderer             *LiveChatPaidStickerRenderer    `json:"liveChatPaidStickerRenderer,omitempty"`
	LiveChatViewerEngagementMessageRenderer interface{}                     `json:"liveChatViewerEngagementMessageRenderer,omitempty"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type MessageText struct {
	Text string `json:"text"`
}

type MessageEmoji struct {
	EmojiId          string   `json:"emojiId"`
	Shortcuts        []string `json:"shortcuts"`
	SearchTerms      []string `json:"searchTerms"`
	SupportsSkinTone bool     `json:"supportsSkinTone"`
	Image            struct {
		Thumbnails    []Thumbnail `json:"thumbnails"`
		Accessibility struct {
			AccessibilityData struct {
				Label string `json:"label"`
			} `json:"accessibilityData"`
		} `json:"accessibility"`
	} `json:"image"`
	VariantIds    []string `json:"variantIds"`
	IsCustomEmoji bool     `json:"isCustomEmoji,omitempty"` // boolean or true? TypeScript says true, treated as bool
}

type MessageRun struct {
	Text  string        `json:"text,omitempty"`
	Emoji *MessageEmoji `json:"emoji,omitempty"`
}

type AuthorBadge struct {
	LiveChatAuthorBadgeRenderer struct {
		CustomThumbnail *struct {
			Thumbnails []Thumbnail `json:"thumbnails"`
		} `json:"customThumbnail,omitempty"`
		Icon *struct {
			IconType string `json:"iconType"`
		} `json:"icon,omitempty"`
		Tooltip       string `json:"tooltip"`
		Accessibility struct {
			AccessibilityData struct {
				Label string `json:"label"`
			} `json:"accessibilityData"`
		} `json:"accessibility"`
	} `json:"liveChatAuthorBadgeRenderer"`
}

type MessageRendererBase struct {
	AuthorName *struct {
		SimpleText string `json:"simpleText"`
	} `json:"authorName,omitempty"`
	AuthorPhoto struct {
		Thumbnails []Thumbnail `json:"thumbnails"`
	} `json:"authorPhoto"`
	AuthorBadges        []AuthorBadge `json:"authorBadges,omitempty"`
	ContextMenuEndpoint struct {
		ClickTrackingParams string `json:"clickTrackingParams"`
		CommandMetadata     struct {
			WebCommandMetadata struct {
				IgnoreNavigation bool `json:"ignoreNavigation"`
			} `json:"webCommandMetadata"`
		} `json:"commandMetadata"`
		LiveChatItemContextMenuEndpoint struct {
			Params string `json:"params"`
		} `json:"liveChatItemContextMenuEndpoint"`
	} `json:"contextMenuEndpoint"`
	ID                       string `json:"id"`
	TimestampUsec            string `json:"timestampUsec"`
	AuthorExternalChannelId  string `json:"authorExternalChannelId"`
	ContextMenuAccessibility struct {
		AccessibilityData struct {
			Label string `json:"label"`
		} `json:"accessibilityData"`
	} `json:"contextMenuAccessibility"`
}

type LiveChatTextMessageRenderer struct {
	MessageRendererBase
	Message struct {
		Runs []MessageRun `json:"runs"`
	} `json:"message"`
}

type LiveChatPaidMessageRenderer struct {
	LiveChatTextMessageRenderer
	PurchaseAmountText struct {
		SimpleText string `json:"simpleText"`
	} `json:"purchaseAmountText"`
	HeaderBackgroundColor int `json:"headerBackgroundColor"`
	HeaderTextColor       int `json:"headerTextColor"`
	BodyBackgroundColor   int `json:"bodyBackgroundColor"`
	BodyTextColor         int `json:"bodyTextColor"`
	AuthorNameTextColor   int `json:"authorNameTextColor"`
}

type LiveChatPaidStickerRenderer struct {
	MessageRendererBase
	PurchaseAmountText struct {
		SimpleText string `json:"simpleText"`
	} `json:"purchaseAmountText"`
	Sticker struct {
		Thumbnails    []Thumbnail `json:"thumbnails"`
		Accessibility struct {
			AccessibilityData struct {
				Label string `json:"label"`
			} `json:"accessibilityData"`
		} `json:"accessibility"`
	} `json:"sticker"`
	MoneyChipBackgroundColor int `json:"moneyChipBackgroundColor"`
	MoneyChipTextColor       int `json:"moneyChipTextColor"`
	StickerDisplayWidth      int `json:"stickerDisplayWidth"`
	StickerDisplayHeight     int `json:"stickerDisplayHeight"`
	BackgroundColor          int `json:"backgroundColor"`
	AuthorNameTextColor      int `json:"authorNameTextColor"`
}

type LiveChatMembershipItemRenderer struct {
	MessageRendererBase
	HeaderSubtext struct {
		Runs []MessageRun `json:"runs"`
	} `json:"headerSubtext"`
}

// FetchOptions for get_live_chat
type FetchOptions struct {
	ApiKey        string
	ClientVersion string
	Continuation  string
	LiveID        string // Added to store liveID as in parser.ts return type
}
