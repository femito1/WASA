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
	Id             uint64     `json:"id"`
	ConversationId uint64     `json:"-"`
	SenderId       uint64     `json:"senderId"`
	Content        string     `json:"content"`
	Format         string     `json:"format"`
	State          string     `json:"state"`
	Reactions      []Reaction `json:"reactions"`
	Timestamp      string     `json:"timestamp,omitempty"`
}

// Reaction represents an emoji reaction to a message.
type Reaction struct {
	Emoji string `json:"emoji"`
}

// Comment represents a comment on a message.
type Comment struct {
	Id          uint64 `json:"id"`
	MessageId   uint64 `json:"messageId"`
	UserId      uint64 `json:"userId"`
	CommentText string `json:"commentText"`
	Timestamp   string `json:"timestamp,omitempty"`
}
