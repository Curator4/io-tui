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
	AsciiArtPath string
	IsActive bool
	Created string
}

func GetAIByID(db *sql.DB, id int) (AI, error) {
	row := db.QueryRow(`
		SELECT id, name, system_prompt, api, model, ascii_art_path, is_active, created
		FROM ais WHERE id = ?
	`, id)
	return scanAI(row)
}

func GetAIByName(db *sql.DB, name string) (AI, error) {
	row := db.QueryRow(`
		SELECT id, name, system_prompt, api, model, ascii_art_path, is_active, created
		FROM ais WHERE name = ?
	`, name)
	return scanAI(row)
}

func ListAIs(db *sql.DB) ([]AI, error) {
	rows, err := db.Query(`
		SELECT id, name, system_prompt, api, model, ascii_art_path, is_active, created
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
		SELECT id, name, system_prompt, api, model, ascii_art_path, is_active, created
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


func GetAIAsciiPath(db *sql.DB, aiID int) (string, error) {
	var asciiPath string
	row := db.QueryRow(`
		SELECT ascii_art_path 
		FROM ais WHERE id = ?
	`, aiID)
	
	err := row.Scan(&asciiPath)
	if err != nil {
		return "", err
	}
	
	return asciiPath, nil
}

func CreateAI(db *sql.DB, name, prompt, api, model, asciiPath string, isActive bool) error {
	_, err := db.Exec(`
		INSERT INTO ais (name, system_prompt, api, model, ascii_art_path, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, name, prompt, api, model, asciiPath, isActive)
	return err
}

func UpdateActiveAIAPI(db *sql.DB, apiName string, defaultModel string) (AI, error) {
	// Update the active AI's API and set it to the default model
	_, err := db.Exec(`
		UPDATE ais 
		SET api = ?, model = ?
		WHERE is_active = true
	`, apiName, defaultModel)
	if err != nil {
		return AI{}, err
	}
	
	// Return the updated active AI
	return GetActiveAI(db)
}

func UpdateActiveAIModel(db *sql.DB, model string) (AI, error) {
	// Update the active AI's model
	_, err := db.Exec(`
		UPDATE ais 
		SET model = ?
		WHERE is_active = true
	`, model)
	if err != nil {
		return AI{}, err
	}
	
	// Return the updated active AI
	return GetActiveAI(db)
}

func UpdateActiveAIPrompt(db *sql.DB, prompt string) (AI, error) {
	// Update the active AI's system prompt
	_, err := db.Exec(`
		UPDATE ais 
		SET system_prompt = ?
		WHERE is_active = true
	`, prompt)
	if err != nil {
		return AI{}, err
	}
	
	// Return the updated active AI
	return GetActiveAI(db)
}

// Helper function to scan AI from database row
func scanAI(scanner interface{ Scan(...interface{}) error }) (AI, error) {
	var ai AI
	err := scanner.Scan(&ai.ID, &ai.Name, &ai.SystemPrompt, &ai.API, &ai.Model, &ai.AsciiArtPath, &ai.IsActive, &ai.Created)
	return ai, err
}
