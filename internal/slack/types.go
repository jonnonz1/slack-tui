package slack

import "time"

// Channel represents a Slack channel, DM, or group DM.
type Channel struct {
	ID          string
	Name        string
	Topic       string
	IsPrivate   bool
	IsDM        bool
	IsGroupDM   bool
	UnreadCount int
	LastMessage time.Time
	MemberCount int
}

// Message represents a single Slack message.
type Message struct {
	ID        string
	ChannelID string
	UserID    string
	Username  string
	Text      string
	Timestamp time.Time
	ThreadTS  string
	ReplyCount int
	Reactions []Reaction
	Edited    bool
	Files     []File
	IsBot     bool
}

// Reaction on a message.
type Reaction struct {
	Name  string
	Count int
	Users []string
}

// File attachment.
type File struct {
	ID       string
	Name     string
	Mimetype string
	Size     int64
	URL      string
}

// User in the workspace.
type User struct {
	ID          string
	Username    string
	DisplayName string
	StatusText  string
	StatusEmoji string
	IsOnline    bool
	IsBot       bool
	Avatar      string
}

// Event types sent from Socket Mode into bubbletea.

type NewMessageEvent struct {
	Message Message
}

type MessageEditedEvent struct {
	Message Message
}

type MessageDeletedEvent struct {
	ChannelID string
	Timestamp string
}

type ReactionAddedEvent struct {
	ChannelID string
	MessageTS string
	Reaction  string
	UserID    string
}

type ReactionRemovedEvent struct {
	ChannelID string
	MessageTS string
	Reaction  string
	UserID    string
}

type ChannelMarkedEvent struct {
	ChannelID string
	Timestamp string
}

type PresenceChangeEvent struct {
	UserID   string
	IsOnline bool
}

type TypingEvent struct {
	ChannelID string
	UserID    string
}

type ConnectionStatusEvent struct {
	Connected bool
}
