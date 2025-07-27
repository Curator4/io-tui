package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Init() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "data.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign key constraints in SQLite
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	if err := CreateIo(db); err != nil {
		return nil, fmt.Errorf("failed to create default ai: %w", err)
	}

	// Clear any active conversations on startup - fresh slate every time
	if err := ClearActiveConversations(db); err != nil {
		return nil, fmt.Errorf("failed to clear active conversations: %w", err)
	}

	return db, nil
}


func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS ais (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		system_prompt TEXT,
		api TEXT NOT NULL,
		model TEXT NOT NULL,
		ascii_art_path TEXT DEFAULT 'ascii/io_ascii.txt',
		is_active BOOLEAN DEFAULT FALSE,
		created DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ai_id INTEGER,
		name TEXT NOT NULL,
		is_active BOOLEAN DEFAULT FALSE,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (ai_id) REFERENCES ais(id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id)
	);`

	_, err := db.Exec(schema)
	return err
}
