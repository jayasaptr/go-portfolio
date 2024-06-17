package model

import (
	"database/sql"
	"log"
	"time"
)

type Portfolio struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle"`
	Image       string      `json:"image"`
	Content     string      `json:"content"`
	Status      string      `json:"status"`
	DateProject time.Time   `json:"date_project"`
	Skills      []Skills    `json:"skills"`
	Experience  *Experience `json:"experience"`
}

type PortfolioSkill struct {
	PortfolioID string `json:"portfolio_id"`
	SkillID     string `json:"skill_id"`
}

// Function to associate multiple skills with a single portfolio
func (p *Portfolio) AddSkills(db *sql.DB, skillIDs []string) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Prepare the SQL statement for inserting portfolio-skill relationships
	stmt, err := tx.Prepare("INSERT INTO portfolio_skills (portfolio_id, skill_id) VALUES ($1, $2)")
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing SQL statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Execute the statement for each skill ID
	for _, skillID := range skillIDs {
		_, err := stmt.Exec(p.ID, skillID)
		if err != nil {
			tx.Rollback()
			log.Printf("Error executing SQL statement: %v", err)
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}
	return nil
}

// add experience
func (p *Portfolio) AddExperience(db *sql.DB, experienceID string) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Prepare the SQL statement for inserting portfolio-experience relationships
	stmt, err := tx.Prepare("INSERT INTO portfolio_experience (portfolio_id, experiance_id) VALUES ($1, $2)")
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing SQL statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Execute the statement
	_, err = stmt.Exec(p.ID, experienceID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error executing SQL statement: %v", err)
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}
	return nil
}

// update experience
func (p *Portfolio) UpdateExperiencePortfolio(db *sql.DB, experienceID string) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Prepare the SQL statement for inserting portfolio-skill relationships
	stmt, err := tx.Prepare("UPDATE portfolio_experience SET experiance_id = $2 WHERE portfolio_id = $1")
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing SQL statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Execute the statement for each skill ID
	_, err = stmt.Exec(p.ID, experienceID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error executing SQL statement: %v", err)
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}
	return nil

}

// Function to insert a new portfolio into the database
func InsertPortfolio(db *sql.DB, portfolio *Portfolio) error {
	query := `INSERT INTO portfolio (id, title, subtitle, image, content, status, date_project) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.Exec(query, portfolio.ID, portfolio.Title, portfolio.Subtitle, portfolio.Image, portfolio.Content, portfolio.Status, portfolio.DateProject)
	if err != nil {
		log.Printf("Error inserting portfolio: %v", err)
		return err
	}
	return nil
}

// Function to retrieve a portfolio along with its associated skills
func GetPortfoliosPaginated(db *sql.DB, offset int, limit int) ([]*Portfolio, error) {
	query := `SELECT id, title, subtitle, image, content, status, date_project FROM portfolio LIMIT $1 OFFSET $2`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		log.Printf("Error querying portfolios: %v", err)
		return nil, err
	}
	defer rows.Close()

	var portfolios []*Portfolio
	for rows.Next() {
		var portfolio Portfolio
		if err := rows.Scan(&portfolio.ID, &portfolio.Title, &portfolio.Subtitle, &portfolio.Image, &portfolio.Content, &portfolio.Status, &portfolio.DateProject); err != nil {
			log.Printf("Error scanning portfolio: %v", err)
			return nil, err
		}
		portfolios = append(portfolios, &portfolio)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		return nil, err
	}

	return portfolios, nil
}

// Function to delete a skill and its relations from the database
func DeleteSkillAndPortfolioRelations(db *sql.DB, skillID string, portfolioID string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Delete relations from portfolio_skills table for the given skill ID and portfolio ID
	deleteRelationsQuery := `DELETE FROM portfolio_skills WHERE skill_id = $1 AND portfolio_id = $2`
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

// Function to update a portfolio in the database
func UpdatePortfolio(db *sql.DB, portfolio *Portfolio) error {
	query := `UPDATE portfolio SET title = $2, subtitle = $3, image = $4, content = $5, status = $6, date_project = $7 WHERE id = $1`
	_, err := db.Exec(query, portfolio.ID, portfolio.Title, portfolio.Subtitle, portfolio.Image, portfolio.Content, portfolio.Status, portfolio.DateProject)
	if err != nil {
		log.Printf("Error updating portfolio: %v", err)
		return err
	}
	return nil
}

// GetPortfolioByID retrieves a portfolio by its ID along with its associated skills
func GetPortfolioByID(db *sql.DB, portfolioID string) (*Portfolio, error) {
	portfolioQuery := `SELECT id, title, subtitle, image, content, status, date_project FROM portfolio WHERE id = $1`
	row := db.QueryRow(portfolioQuery, portfolioID)

	var portfolio Portfolio
	err := row.Scan(&portfolio.ID, &portfolio.Title, &portfolio.Subtitle, &portfolio.Image, &portfolio.Content, &portfolio.Status, &portfolio.DateProject)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No portfolio found with ID: %v", portfolioID)
			return nil, err
		}
		log.Printf("Error retrieving portfolio: %v", err)
		return nil, err
	}

	// Retrieve associated skills
	skills, err := GetSkillsByPortfolioID(db, portfolioID)
	if err != nil {
		log.Printf("Error retrieving skills for portfolio %s: %v", portfolioID, err)
		return nil, err
	}

	// Retrieve associated experience
	experience, err := GetExperienceByPortfolioID(db, portfolioID)
	if err != nil {
		log.Printf("Error retrieving experience for portfolio %s: %v", portfolioID, err)
		return nil, err

	}

	portfolio.Skills = skills
	portfolio.Experience = experience

	return &portfolio, nil
}

// Function to delete a portfolio and its relations from the database without deleting the master skills
func DeletePortfolioAndRelations(db *sql.DB, portfolioID string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Delete relations from portfolio_skills table only, do not delete the skills themselves
	deleteRelationsQuery := `DELETE FROM portfolio_skills WHERE portfolio_id = $1`
	if _, err := tx.Exec(deleteRelationsQuery, portfolioID); err != nil {
		tx.Rollback()
		log.Printf("Error deleting portfolio-skill relations: %v", err)
		return err
	}

	// Delete relations from portfolio_experience table
	deleteExperienceRelationsQuery := `DELETE FROM portfolio_experience WHERE portfolio_id = $1`
	if _, err := tx.Exec(deleteExperienceRelationsQuery, portfolioID); err != nil {
		tx.Rollback()
		log.Printf("Error deleting portfolio-experience relations: %v", err)
		return err
	}

	// Delete the portfolio from portfolio table
	deletePortfolioQuery := `DELETE FROM portfolio WHERE id = $1`
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

// GetSkillsByPortfolioID retrieves the skills associated with a given portfolio ID
func GetSkillsByPortfolioID(db *sql.DB, portfolioID string) ([]Skills, error) {
	query := `SELECT skills.id, skills.name, skills.image FROM skills 
	          INNER JOIN portfolio_skills ON skills.id = portfolio_skills.skill_id 
	          WHERE portfolio_skills.portfolio_id = $1`
	rows, err := db.Query(query, portfolioID)
	if err != nil {
		log.Printf("Error querying skills by portfolio ID: %v", err)
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
		log.Printf("Error during skills rows iteration: %v", err)
		return nil, err
	}

	return skills, nil
}

// get experience by portfolio id
func GetExperienceByPortfolioID(db *sql.DB, portfolioID string) (*Experience, error) {
	query := `SELECT experiance.id, experiance.company_name, experiance.position, experiance.image, experiance.start_date, experiance.end_date, experiance.location FROM experiance INNER JOIN portfolio_experience ON experiance.id = portfolio_experience.experiance_id WHERE portfolio_experience.portfolio_id = $1`

	rows, err := db.Query(query, portfolioID)
	if err != nil {
		log.Printf("Error querying experience by portfolio ID: %v", err)
		return nil, err
	}
	defer rows.Close()

	var experience Experience
	for rows.Next() {
		if err := rows.Scan(&experience.ID, &experience.CompanyName, &experience.Position, &experience.Image, &experience.StartDate, &experience.EndDate, &experience.Location); err != nil {
			log.Printf("Error scanning experience: %v", err)
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during experience rows iteration: %v", err)
		return nil, err
	}

	return &experience, nil
}
