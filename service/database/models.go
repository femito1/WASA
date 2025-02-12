package database

// Conversation represents a conversation between users.
type Conversation struct {
	Id       uint64    `json:"id"`
	Name     string    `json:"name"`
	Picture  string    `json:"picture,omitempty"`
	Members  []User    `json:"members"`
	Messages []Message `json:"messages"`
}

// Message represents a message in a conversation.
type Message struct {
	ConversationId uint64     `json:"-"`
	Id             uint64     `json:"id"`
	SenderId       uint64     `json:"senderId"`
	SenderName     string     `json:"senderName,omitempty"`    // NEW field to hold the sender's username.
	SenderPicture  string     `json:"senderPicture,omitempty"` // NEW field for profile picture
	Content        string     `json:"content"`
	Format         string     `json:"format"`
	State          string     `json:"state"`
	Reactions      []Reaction `json:"reactions"`
	Timestamp      string     `json:"timestamp,omitempty"`
	ReplyToID      *uint64    `json:"replyToId,omitempty"`
	ReplyTo        *Message   `json:"replyTo,omitempty"`
	IsForwarded    bool       `json:"isForwarded"`
}

// Reaction represents an emoji reaction to a message.
type Reaction struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count,omitempty"` // Count of this reaction (aggregated)
}

type Comment struct {
	Id          uint64 `json:"id"`
	MessageId   uint64 `json:"messageId"`
	UserId      uint64 `json:"userId"`
	CommentText string `json:"commentText"`
	Timestamp   string `json:"timestamp,omitempty"`
	SenderName  string `json:"senderName,omitempty"` // ADDED: sender's username
}
