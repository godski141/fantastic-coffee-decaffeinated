package database

type User struct {
    UserID   string
    Name string
    Photo string
}

type Conversation struct {
    ConvID        string
    Name      string
    Type      string
    CreatorID string
    Photo     string
    LastMessage string
}

type Message struct {
    MessageID      string
    ConversationID string
    SenderID       string
    Content        string
    Timestamp      string
    Status         string
}

type Comment struct {
    Emoji     string
    MessageID string
}