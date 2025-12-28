package youtubechat

import (
	"errors"
	"time"

	"github.com/DiegPS/youtube-chat/types"
)

type LiveChat struct {
	// Events exposed as channels
	ChatChan  chan types.ChatItem
	ErrorChan chan error
	StartChan chan string
	EndChan   chan string

	liveID   string
	observer *time.Ticker
	options  *types.FetchOptions
	interval time.Duration
	id       types.YoutubeId
	stopChan chan struct{}
	running  bool

	// Fetch replacement for testing
	FetchLivePageFunc func(types.YoutubeId) (types.FetchOptions, error)
	FetchChatFunc     func(types.FetchOptions) ([]types.ChatItem, string, error)
}

func NewLiveChat(id types.YoutubeId, intervalMs int) (*LiveChat, error) {
	if id.ChannelID == "" && id.LiveID == "" && id.Handle == "" {
		return nil, errors.New("Required channelId or liveId or handle.")
	}

	lc := &LiveChat{
		ChatChan:          make(chan types.ChatItem, 100),
		ErrorChan:         make(chan error, 10),
		StartChan:         make(chan string, 1),
		EndChan:           make(chan string, 1),
		id:                id,
		interval:          time.Duration(intervalMs) * time.Millisecond,
		stopChan:          make(chan struct{}),
		FetchLivePageFunc: FetchLivePage,
		FetchChatFunc:     FetchChat,
	}

	if lc.interval == 0 {
		lc.interval = 1000 * time.Millisecond
	}

	if id.LiveID != "" {
		lc.liveID = id.LiveID
	}

	return lc, nil
}

func (lc *LiveChat) Start() bool {
	if lc.running {
		return false
	}

	// Fetch initial options (FetchLivePage)
	options, err := lc.FetchLivePageFunc(lc.id)
	if err != nil {
		// Emit error? The TS code emits error AND returns false logic.
		// TS: emit("error", err); return false
		// Since we haven't started the loop yet, user might be listening.
		// We should try to send non-blocking or just return false?
		// Better to run this async or blocking? TS start() is async.
		// In Go, usually Start() starts the thing.
		// Let's do the initial fetch synchronously purely to match 'await fetchLivePage' before interval?
		// But in TS it returns a Promise.
		// We can do it synchronously here.
		lc.emitError(err)
		return false
	}

	lc.liveID = options.LiveID
	lc.options = &options

	lc.running = true
	lc.stopChan = make(chan struct{})

	// Start loop
	lc.observer = time.NewTicker(lc.interval)

	// Emit start
	select {
	case lc.StartChan <- lc.liveID:
	default:
	}

	go lc.loop()

	return true
}

func (lc *LiveChat) Stop(reason string) {
	if lc.running {
		lc.running = false
		if lc.observer != nil {
			lc.observer.Stop()
		}
		close(lc.stopChan)

		select {
		case lc.EndChan <- reason:
		default:
		}
	}
}

func (lc *LiveChat) loop() {
	for {
		select {
		case <-lc.stopChan:
			return
		case <-lc.observer.C:
			lc.execute()
		}
	}
}

func (lc *LiveChat) execute() {
	if lc.options == nil {
		msg := "Not found options"
		lc.emitError(errors.New(msg))
		lc.Stop(msg)
		return
	}

	items, continuation, err := lc.FetchChatFunc(*lc.options)
	if err != nil {
		lc.emitError(err)
		return
	}

	for _, item := range items {
		select {
		case lc.ChatChan <- item:
		default:
			// If channel full, drop? Or block?
			// Blocking might stall the ticker. Drop if full is safer for realtime.
			// Or make buffer large enough.
		}
	}

	lc.options.Continuation = continuation
}

func (lc *LiveChat) emitError(err error) {
	select {
	case lc.ErrorChan <- err:
	default:
	}
}
