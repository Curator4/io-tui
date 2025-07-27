package db

import (
	"database/sql"
	"time"
)

type Conversation struct {
	ID int
	AIID int
	Name string
	IsActive bool
	Created string
}

func CreateConversation(db *sql.DB, firstMessage string, ai_id int) (Conversation, error) {
	var conversationName string
	if len(firstMessage) >= 20 {
		conversationName = firstMessage[:20] + "..."
	} else {
		conversationName = time.Now().Format("2006-01-02 15:04:05")
	}

	// First, clear any existing active conversations
	if err := ClearActiveConversations(db); err != nil {
		return Conversation{}, err
	}

	// Then create new conversation as active
	result, err := db.Exec(`
		INSERT INTO conversations (ai_id, name, is_active)
		VALUES (?, ?, true)
	`, ai_id, conversationName)
	
	if err != nil {
		return Conversation{}, err
	}

	// Get the ID of the inserted conversation
	id, err := result.LastInsertId()
	if err != nil {
		return Conversation{}, err
	}

	// Return the full conversation struct
	return GetConversationByID(db, int(id))
}

func GetConversationByID(db *sql.DB, id int) (Conversation, error) {
	row := db.QueryRow(`
		SELECT id, ai_id, name, is_active, created
		FROM conversations WHERE id = ?
	`, id)
	return scanConversation(row)
}

func GetConversationByName(db *sql.DB, name string) (Conversation, error) {
	row := db.QueryRow(`
		SELECT id, ai_id, name, is_active, created
		FROM conversations WHERE name = ?
	`, name)
	return scanConversation(row)
}

func GetActiveConversation(db *sql.DB) (Conversation, error) {
	row := db.QueryRow(`
		SELECT id, ai_id, name, is_active, created
		FROM conversations WHERE is_active = true
	`)
	return scanConversation(row)
}

func SetActiveConversation(db *sql.DB, id int) (Conversation, error) {
	// Update in one atomic operation using CASE (SQL ternary!)
	_, err := db.Exec(`
		UPDATE conversations SET is_active = CASE 
			WHEN id = ? THEN true 
			ELSE false 
		END
	`, id)
	if err != nil {
		return Conversation{}, err
	}

	// Return the newly active conversation
	return GetConversationByID(db, id)
}

func ListConversations(db *sql.DB) ([]Conversation, error) {
	rows, err := db.Query(`
		SELECT id, ai_id, name, is_active, created
		FROM conversations
		ORDER BY created DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		conv, err := scanConversation(rows)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}
	return conversations, rows.Err()
}

func DeleteConversation(db *sql.DB, id int) error {
	// First delete all messages in this conversation
	if err := DeleteMessagesByConversation(db, id); err != nil {
		return err
	}
	
	// Then delete the conversation itself
	_, err := db.Exec("DELETE FROM conversations WHERE id = ?", id)
	return err
}

func ClearActiveConversations(db *sql.DB) error {
	_, err := db.Exec("UPDATE conversations SET is_active = false")
	return err
}

// Helper function to scan Conversation from database row
func scanConversation(scanner interface{ Scan(...interface{}) error }) (Conversation, error) {
	var conv Conversation
	err := scanner.Scan(&conv.ID, &conv.AIID, &conv.Name, &conv.IsActive, &conv.Created)
	return conv, err
}
