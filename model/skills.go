package model

import (
	"database/sql"
	"errors"
	"log"
)

type Skills struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`
}

func InsertSkills(db *sql.DB, skills Skills) error {
	if db == nil {
		log.Println("Error: Database is nil")
		return ErrDBNil
	}

	query := `INSERT INTO skills (id, name, image) VALUES ($1, $2, $3);`
	_, err := db.Exec(query, skills.ID, skills.Name, skills.Image)

	if err != nil {
		log.Printf("Error inserting skills: %v", err)
		return err
	}

	return nil
}

func GetListSkills(db *sql.DB, offset int, limit int) ([]Skills, error) {
	if db == nil {
		log.Println("Error: Database is nil")
		return nil, ErrDBNil
	}

	query := `SELECT id, name, image FROM skills LIMIT $1 OFFSET $2`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		log.Printf("Error querying skills: %v", err)
		return nil, err
	}
	defer rows.Close()

	var skillsList []Skills
	for rows.Next() {
		var skill Skills
		if err := rows.Scan(&skill.ID, &skill.Name, &skill.Image); err != nil {
			log.Printf("Error scanning skills: %v", err)
			return nil, err
		}
		skillsList = append(skillsList, skill)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		return nil, err
	}

	return skillsList, nil
}

func DeleteSkill(db *sql.DB, skillID string) error {
	if db == nil {
		log.Println("Error: Database is nil")
		return ErrDBNil
	}

	query := `DELETE FROM skills WHERE id = $1`
	_, err := db.Exec(query, skillID)

	if err != nil {
		log.Printf("Error deleting skill: %v", err)
		return err
	}

	return nil
}

func UpdateSkill(db *sql.DB, skill *Skills) error {
	if db == nil {
		log.Println("Error: Database is nil")
		return ErrDBNil
	}

	query := `UPDATE skills SET name = $2, image = $3 WHERE id = $1`
	_, err := db.Exec(query, skill.ID, skill.Name, skill.Image)
	if err != nil {
		log.Printf("Error updating skill: %v", err)
		return err
	}

	return nil
}

func GetSkillID(db *sql.DB, skillID string) (*Skills, error) {
	if db == nil {
		log.Println("Error: Database is nil")
		return nil, ErrDBNil
	}

	query := `SELECT id, name, image FROM skills WHERE id = $1`
	row := db.QueryRow(query, skillID)

	var skill Skills
	err := row.Scan(&skill.ID, &skill.Name, &skill.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Error: No skill found")
			return nil, errors.New("no skill found")
		}
		log.Printf("Error scanning skill: %v", err)
		return nil, err
	}

	return &skill, nil
}
