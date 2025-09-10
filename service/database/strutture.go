package database

type User struct {
    UserID   string `json:"UserID"`
    Name string `json:"Name"`
    Photo string `json:"Photo"`
}

type Conversation struct {
    ConvID        string `json:"ConvID"`
    Name      string `json:"Name"`
    Type      string `json:"Type"`
    CreatorID string `json:"CreatorID"`
    Photo     string `json:"Photo"`
    LastMessage string `json:"LastMessage"`
}

type Message struct {
    MessageID      string `json:"MessageID"`
    ConversationID string `json:"ConversationID"`
    SenderID       string `json:"SenderID"`
    SenderUsername string `json:"SenderUsername"`  // NUOVO CAMPO
    Content        string `json:"Content"`
    Timestamp      string `json:"Timestamp"`
    Status         string `json:"Status"`
    Type           string `json:"Type"`
    ReplyToMessageID *string `json:"ReplyToMessageID"` // Nullable per messaggi che non sono reply
    Reactions []Reaction `json:"Reactions"`
}

type Comment struct {
    Emoji     string `json:"Emoji"`
    MessageID string `json:"MessageID"`
}

type Reaction struct {
    UserID   string `json:"UserID"`
    Reaction string `json:"Reaction"`
}