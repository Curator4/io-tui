package db

import (
	"database/sql"
)

type AI struct {
	ID int
	Name string
	SystemPrompt string
	API string
	Model string
	IsActive bool
	Created string
}

func GetAIByID(db *sql.DB, id int) (AI, error) {
	row := db.QueryRow(`
		SELECT id, name, system_prompt, api, model, is_active, created
		FROM ais WHERE id = ?
	`, id)
	return scanAI(row)
}

func GetAIByName(db *sql.DB, name string) (AI, error) {
	row := db.QueryRow(`
		SELECT id, name, system_prompt, api, model, is_active, created
		FROM ais WHERE name = ?
	`, name)
	return scanAI(row)
}

func ListAIs(db *sql.DB) ([]AI, error) {
	rows, err := db.Query(`
		SELECT id, name, system_prompt, api, model, is_active, created
		FROM ais
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ais []AI
	for rows.Next() {
		ai, err := scanAI(rows)
		if err != nil {
			return nil, err
		}
		ais = append(ais, ai)
	}
	return ais, rows.Err()
}

func GetActiveAI(db *sql.DB) (AI, error) {
	row := db.QueryRow(`
		SELECT id, name, system_prompt, api, model, is_active, created
		FROM ais WHERE is_active = true
	`)
	return scanAI(row)
}

func SetActiveAI(db *sql.DB, name string) (AI, error) {
	// Update in one atomic operation
	_, err := db.Exec(`
		UPDATE ais SET is_active = CASE 
			WHEN name = ? THEN true 
			ELSE false 
		END
	`, name)
	if err != nil {
		return AI{}, err
	}
	
	// Return the newly active AI
	return GetAIByName(db, name)
}


func CreateIo(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO ais (name, system_prompt, api, model, is_active)
		VALUES ('Io', 'You are a helpful AI assistant', 'gemini', 'gemini-2.0-flash', true)
	`)
	return err
}

// Helper function to scan AI from database row
func scanAI(scanner interface{ Scan(...interface{}) error }) (AI, error) {
	var ai AI
	err := scanner.Scan(&ai.ID, &ai.Name, &ai.SystemPrompt, &ai.API, &ai.Model, &ai.IsActive, &ai.Created)
	return ai, err
}
