package model

import (
	"database/sql"
	"errors"
)

type Skills struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func InsertSkills(db *sql.DB, skills Skills) error {
	if db == nil {
		return ErrDBNil
	}

	query := `INSERT INTO skills (id, name, image) VALUES ($1, $2, $3);`
	_, err := db.Exec(query, skills.ID, skills.Name, skills.Image)

	if err != nil {
		return err
	}

	return nil
}

func GetListSkills(db *sql.DB, offset int, limit int) ([]Skills, error) {
	if db == nil {
		return nil, ErrDBNil
	}

	query := `SELECT id, name, image FROM skills LIMIT $1 OFFSET $2`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skillsList []Skills
	for rows.Next() {
		var skill Skills
		if err := rows.Scan(&skill.ID, &skill.Name, &skill.Image); err != nil {
			return nil, err
		}
		skillsList = append(skillsList, skill)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return skillsList, nil
}

func DeleteSkill(db *sql.DB, skillID string) error {
	if db == nil {
		return ErrDBNil
	}

	query := `DELETE FROM skills WHERE id = $1`
	_, err := db.Exec(query, skillID)

	if err != nil {
		return err
	}

	return nil
}

func GetSkillID(db *sql.DB, skillID string) (*Skills, error) {
	if db == nil {
		return nil, ErrDBNil
	}

	query := `SELECT id, name, image FROM skills WHERE id = $1`
	row := db.QueryRow(query, skillID)

	var skill Skills
	err := row.Scan(&skill.ID, &skill.Name, &skill.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no user found")
		}
		return nil, err
	}

	return &skill, nil
}
