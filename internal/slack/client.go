package slack

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	tea "charm.land/bubbletea/v2"
)

// discardLogger silences slack-go's internal logging so it doesn't
// pollute the TUI output.
var discardLogger = log.New(noopWriter{}, "", 0)

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

// Client wraps the Slack Web API and Socket Mode connections.
type Client struct {
	api       *slack.Client
	socket    *socketmode.Client
	userCache *Cache[User]
	chanCache *Cache[Channel]
	selfID    string
}

func NewClient(userToken, appToken string) (*Client, error) {
	api := slack.New(
		userToken,
		slack.OptionAppLevelToken(appToken),
		slack.OptionLog(discardLogger),
	)

	socket := socketmode.New(
		api,
		socketmode.OptionLog(discardLogger),
	)

	// Verify auth and get our own user ID
	resp, err := api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("auth test failed: %w", err)
	}

	return &Client{
		api:       api,
		socket:    socket,
		userCache: NewCache[User](5 * time.Minute),
		chanCache: NewCache[Channel](2 * time.Minute),
		selfID:    resp.UserID,
	}, nil
}

func (c *Client) SelfID() string {
	return c.selfID
}

// ListChannels returns all channels the user is a member of.
func (c *Client) ListChannels() ([]Channel, error) {
	var channels []Channel

	params := &slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel", "mpim", "im"},
		Limit:           200,
		ExcludeArchived: true,
	}

	for {
		convs, cursor, err := c.api.GetConversations(params)
		if err != nil {
			return nil, fmt.Errorf("list channels: %w", err)
		}

		for _, conv := range convs {
			ch := Channel{
				ID:          conv.ID,
				Name:        c.channelName(conv),
				IsPrivate:   conv.IsPrivate,
				IsDM:        conv.IsIM,
				IsGroupDM:   conv.IsMpIM,
				UnreadCount: conv.UnreadCount,
				MemberCount: conv.NumMembers,
			}
			if conv.Topic.Value != "" {
				ch.Topic = conv.Topic.Value
			}
			channels = append(channels, ch)
			c.chanCache.Set(ch.ID, ch)
		}

		if cursor == "" {
			break
		}
		params.Cursor = cursor
	}

	return channels, nil
}

func (c *Client) channelName(conv slack.Channel) string {
	if conv.IsIM {
		user, err := c.GetUser(conv.User)
		if err == nil {
			return "@" + user.DisplayName
		}
		return "@" + conv.User
	}
	if conv.Name != "" {
		return "#" + conv.Name
	}
	return conv.ID
}

// GetHistory fetches messages for a channel.
func (c *Client) GetHistory(channelID string, limit int) ([]Message, error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     limit,
	}

	resp, err := c.api.GetConversationHistory(params)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}

	messages := make([]Message, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		messages = append(messages, c.convertMessage(channelID, msg))
	}

	// API returns newest first; reverse for display
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetReplies fetches thread replies.
func (c *Client) GetReplies(channelID, threadTS string) ([]Message, error) {
	msgs, _, _, err := c.api.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
		Limit:     100,
	})
	if err != nil {
		return nil, fmt.Errorf("get replies: %w", err)
	}

	messages := make([]Message, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, c.convertMessage(channelID, msg))
	}
	return messages, nil
}

// SendMessage posts a message to a channel.
func (c *Client) SendMessage(channelID, text string) error {
	_, _, err := c.api.PostMessage(channelID, slack.MsgOptionText(text, false))
	return err
}

// SendReply posts a threaded reply.
func (c *Client) SendReply(channelID, threadTS, text string) error {
	_, _, err := c.api.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(threadTS),
	)
	return err
}

// AddReaction adds an emoji reaction to a message.
func (c *Client) AddReaction(channelID, timestamp, emoji string) error {
	return c.api.AddReaction(emoji, slack.ItemRef{
		Channel:   channelID,
		Timestamp: timestamp,
	})
}

// RemoveReaction removes an emoji reaction.
func (c *Client) RemoveReaction(channelID, timestamp, emoji string) error {
	return c.api.RemoveReaction(emoji, slack.ItemRef{
		Channel:   channelID,
		Timestamp: timestamp,
	})
}

// MarkRead marks a channel as read up to the given timestamp.
func (c *Client) MarkRead(channelID, timestamp string) error {
	return c.api.MarkConversation(channelID, timestamp)
}

// GetUser fetches a user, using cache when available.
func (c *Client) GetUser(userID string) (User, error) {
	if u, ok := c.userCache.Get(userID); ok {
		return u, nil
	}

	info, err := c.api.GetUserInfo(userID)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:          info.ID,
		Username:    info.Name,
		DisplayName: info.Profile.DisplayName,
		StatusText:  info.Profile.StatusText,
		StatusEmoji: info.Profile.StatusEmoji,
		IsBot:       info.IsBot,
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	c.userCache.Set(userID, user)
	return user, nil
}

// SearchMessages searches the workspace.
func (c *Client) SearchMessages(query string) ([]Message, error) {
	params := slack.SearchParameters{
		Sort:  "timestamp",
		Count: 20,
	}
	result, err := c.api.SearchMessages(query, params)
	if err != nil {
		return nil, err
	}

	messages := make([]Message, 0, len(result.Matches))
	for _, match := range result.Matches {
		messages = append(messages, Message{
			ChannelID: match.Channel.ID,
			Username:  match.Username,
			Text:      match.Text,
			Timestamp: parseSlackTimestamp(match.Timestamp),
		})
	}
	return messages, nil
}

// StartSocketMode begins listening for real-time events and relays them as tea.Msg.
func (c *Client) StartSocketMode(p *tea.Program) {
	handler := socketmode.NewSocketmodeHandler(c.socket)

	handler.Handle(socketmode.EventTypeEventsAPI, func(evt *socketmode.Event, client *socketmode.Client) {
		client.Ack(*evt.Request)
		c.handleEvent(evt, p)
	})

	handler.Handle(socketmode.EventTypeHello, func(evt *socketmode.Event, client *socketmode.Client) {
		// Suppress the "Unexpected event type: hello" log
	})

	handler.Handle(socketmode.EventTypeConnecting, func(evt *socketmode.Event, client *socketmode.Client) {
		p.Send(ConnectionStatusEvent{Connected: false})
	})

	handler.Handle(socketmode.EventTypeConnected, func(evt *socketmode.Event, client *socketmode.Client) {
		p.Send(ConnectionStatusEvent{Connected: true})
	})

	handler.Handle(socketmode.EventTypeConnectionError, func(evt *socketmode.Event, client *socketmode.Client) {
		p.Send(ConnectionStatusEvent{Connected: false})
	})

	_ = handler.RunEventLoop()
}

func (c *Client) handleEvent(evt *socketmode.Event, p *tea.Program) {
	eventsAPI, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return
	}

	switch ev := eventsAPI.InnerEvent.Data.(type) {
	case *slack.MessageEvent:
		user, _ := c.GetUser(ev.User)
		msg := Message{
			ID:        ev.ClientMsgID,
			ChannelID: ev.Channel,
			UserID:    ev.User,
			Username:  user.DisplayName,
			Text:      ev.Text,
			Timestamp: parseSlackTimestamp(ev.Timestamp),
			ThreadTS:  ev.ThreadTimestamp,
		}

		switch ev.SubType {
		case "":
			p.Send(NewMessageEvent{Message: msg})
		case "message_changed":
			p.Send(MessageEditedEvent{Message: msg})
		case "message_deleted":
			p.Send(MessageDeletedEvent{
				ChannelID: ev.Channel,
				Timestamp: ev.Timestamp,
			})
		}

	case *slack.ReactionAddedEvent:
		p.Send(ReactionAddedEvent{
			ChannelID: ev.Item.Channel,
			MessageTS: ev.Item.Timestamp,
			Reaction:  ev.Reaction,
			UserID:    ev.User,
		})

	case *slack.ReactionRemovedEvent:
		p.Send(ReactionRemovedEvent{
			ChannelID: ev.Item.Channel,
			MessageTS: ev.Item.Timestamp,
			Reaction:  ev.Reaction,
			UserID:    ev.User,
		})
	}
}

func (c *Client) convertMessage(channelID string, msg slack.Message) Message {
	user, _ := c.GetUser(msg.User)

	var reactions []Reaction
	for _, r := range msg.Reactions {
		reactions = append(reactions, Reaction{
			Name:  r.Name,
			Count: r.Count,
			Users: r.Users,
		})
	}

	var files []File
	for _, f := range msg.Files {
		files = append(files, File{
			ID:       f.ID,
			Name:     f.Name,
			Mimetype: f.Mimetype,
			Size:     int64(f.Size),
			URL:      f.URLPrivateDownload,
		})
	}

	return Message{
		ID:         msg.ClientMsgID,
		ChannelID:  channelID,
		UserID:     msg.User,
		Username:   user.DisplayName,
		Text:       msg.Text,
		Timestamp:  parseSlackTimestamp(msg.Timestamp),
		ThreadTS:   msg.ThreadTimestamp,
		ReplyCount: msg.ReplyCount,
		Reactions:  reactions,
		Edited:     msg.Edited != nil,
		Files:      files,
		IsBot:      msg.BotID != "",
	}
}

func parseSlackTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}
	// Slack timestamps are Unix epoch with microseconds: "1234567890.123456"
	parts := ts
	if dot := len(ts) - 7; dot > 0 && ts[dot] == '.' {
		parts = ts[:dot] + ts[dot+1:]
	}
	n, err := strconv.ParseInt(parts, 10, 64)
	if err != nil {
		// Fallback: try parsing just the seconds part
		if dot := 10; dot < len(ts) {
			n, _ = strconv.ParseInt(ts[:dot], 10, 64)
			return time.Unix(n, 0)
		}
		return time.Time{}
	}
	return time.UnixMicro(n)
}
