package model

import (
	"database/sql"
	"log"
	"time"
)

type Experience struct {
	ID          string    `json:"id"`
	CompanyName string    `json:"company_name"`
	Position    string    `json:"position"`
	Image       string    `json:"image"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Location    string    `json:"location"`
	Skills      []Skills  `json:"skills"`
}

type ExperienceSkill struct {
	ExperienceID string `json:"experience_id"`
	SkillID      string `json:"skill_id"`
}

func (p *Experience) AddSkills(db *sql.DB, skillIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v\n", err)
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO experiance_skills (experiance_id, skill_id) VALUES ($1, $2)")
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing sql statement: %v\n", err)
		return err
	}
	defer stmt.Close()

	//execute the statement for each skill id
	for _, SkillID := range skillIDs {
		_, err := stmt.Exec(p.ID, SkillID)
		if err != nil {
			tx.Rollback()
			log.Printf("Error executing sql statement: %v\n", err)
			return err
		}
	}

	//commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error commiting trasaction: %v\n", err)
		return err
	}
	return nil
}

func InsertExperience(db *sql.DB, experience *Experience) error {
	query := `INSERT INTO experiance (id, company_name, position, image, start_date, end_date, location) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.Exec(query, experience.ID, experience.CompanyName, experience.Position, experience.Image, experience.StartDate, experience.EndDate, experience.Location)

	if err != nil {
		log.Printf("Error inserting experience: %v\n", err)
		return err
	}

	return nil
}

func UpdateExperience(db *sql.DB, experiance *Experience) error {
	query := `UPDATE experiance SET company_name = $2, position = $3, image = $4, start_date = $5, end_date = $6, location = $7 WHERE id = $1`
	_, err := db.Exec(query, experiance.ID, experiance.CompanyName, experiance.Position, experiance.Image, experiance.StartDate, experiance.EndDate, experiance.Location)

	if err != nil {
		log.Printf("Error updating experiance: %v\n", err)
		return err
	}
	return nil
}

func DeleteExperience(db *sql.DB, experianceID string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error Starting transaction: %v\n", err)
		return err
	}

	deleteQuery := `DELETE FROM experiance WHERE id = $1`
	if _, err := tx.Exec(deleteQuery, experianceID); err != nil {
		tx.Rollback()
		log.Printf("Error Deleting experince: %v\n", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error Commiting transaction: %v\n", err)
		return err
	}
	return nil
}

func GetExperience(db *sql.DB, offset int, limit int) ([]*Experience, error) {
	query := `SELECT id, company_name, position, image, start_date, end_date, location FROM experiance LIMIT $1 OFFSET $2`

	rows, err := db.Query(query, limit, offset)

	if err != nil {
		log.Printf("Error querying experience: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var experiances []*Experience
	for rows.Next() {
		var experience Experience
		if err := rows.Scan(&experience.ID, &experience.CompanyName, &experience.Position, &experience.Image, &experience.StartDate, &experience.EndDate, &experience.Location); err != nil {
			log.Printf("Error Scanning experence: %v\n", err)
			return nil, err
		}
		experiances = append(experiances, &experience)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iterationL %v\n", err)
		return nil, err
	}

	return experiances, nil
}

func GetExperienceID(db *sql.DB, experienceID string) (*Experience, error) {
	experienceQuery := `SELECT id, company_name, image, position,start_date, end_date, location FROM experiance WHERE id = $1`
	row := db.QueryRow(experienceQuery, experienceID)

	var experience Experience
	err := row.Scan(&experience.ID, &experience.CompanyName, &experience.Image, &experience.Position, &experience.StartDate, &experience.StartDate, &experience.Location)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No experience found with id: %v\n", err)
			return nil, err
		}
		log.Printf("Error retrieving experience: %v\n", err)
		return nil, err
	}

	return &experience, nil
}

func DeleteSkillAndExperienceRelations(db *sql.DB, skillID string, portfolioID string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Delete relations from portfolio_skills table for the given skill ID and portfolio ID
	deleteRelationsQuery := `DELETE FROM experiance_skills WHERE skill_id = $1 AND experiance_id = $2`
	if _, err := tx.Exec(deleteRelationsQuery, skillID, portfolioID); err != nil {
		tx.Rollback()
		log.Printf("Error deleting relations: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}
	return nil
}

func DeleteExperienceAndRelations(db *sql.DB, portfolioID string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Delete relations from portfolio_skills table only, do not delete the skills themselves
	deleteRelationsQuery := `DELETE FROM experiance_skills WHERE experiance_id = $1`
	if _, err := tx.Exec(deleteRelationsQuery, portfolioID); err != nil {
		tx.Rollback()
		log.Printf("Error deleting portfolio-skill relations: %v", err)
		return err
	}

	// Delete the portfolio from portfolio table
	deletePortfolioQuery := `DELETE FROM experiance WHERE id = $1`
	if _, err := tx.Exec(deletePortfolioQuery, portfolioID); err != nil {
		tx.Rollback()
		log.Printf("Error deleting portfolio: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}
	return nil
}

func GetSkillByExperienceID(db *sql.DB, experienceID string) ([]Skills, error) {
	query := `SELECT skills.id, skills.name, skills.image FROM skills INNER JOIN experiance_skills ON skills.id = experiance_skills.skill_id WHERE experiance_skills.experiance_id = $1`
	rows, err := db.Query(query, experienceID)
	if err != nil {
		log.Printf("Error querying skill by experience id: %v", err)
		return nil, err
	}
	defer rows.Close()

	var skills []Skills
	for rows.Next() {
		var skill Skills
		if err := rows.Scan(&skill.ID, &skill.Name, &skill.Image); err != nil {
			log.Printf("Error scanning skill: %v", err)
			return nil, err
		}

		skills = append(skills, skill)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during skill rows iteration: %v", err)
		return nil, err
	}

	return skills, nil
}
