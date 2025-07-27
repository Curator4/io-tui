package db

import (
	"database/sql"
)

type Message struct {
	ID int
	ConversationID int
	Role string
	Content string
	Created string
}

func SaveMessage(db *sql.DB, conversation_id int, role string, content string) error {
	_ , err := db.Exec(`
		INSERT INTO messages (conversation_id, role, content)
		VALUES (?, ?, ?)
	`, conversation_id, role, content)
	return err
}

func LoadMessages(db *sql.DB, conversation_id int) ([]Message, error) {
	rows, err := db.Query(`
		SELECT id, conversation_id, role, content, created
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created ASC
	`, conversation_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		msg, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func DeleteMessagesByConversation(db *sql.DB, conversationID int) error {
	_, err := db.Exec("DELETE FROM messages WHERE conversation_id = ?", conversationID)
	return err
}

// Helper function to scan Message from database row
func scanMessage(scanner interface{ Scan(...interface{}) error }) (Message, error) {
	var msg Message
	err := scanner.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.Created)
	return msg, err
}
