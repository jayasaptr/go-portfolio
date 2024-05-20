package model

import (
	"database/sql"
	"log"
)

type Portfolio struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Image   string   `json:"image"`
	Content string   `json:"content"`
	Skills  []Skills `json:"skills"`
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

// Function to insert a new portfolio into the database
func InsertPortfolio(db *sql.DB, portfolio *Portfolio) error {
	query := `INSERT INTO portfolio (id, title, image, content) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, portfolio.ID, portfolio.Title, portfolio.Image, portfolio.Content)
	if err != nil {
		log.Printf("Error inserting portfolio: %v", err)
	}
	return err
}

// Function to retrieve a portfolio along with its associated skills
func GetPortfoliosPaginated(db *sql.DB, offset int, limit int) ([]*Portfolio, error) {
	query := `SELECT id, title, image, content FROM portfolio LIMIT $1 OFFSET $2`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		log.Printf("Error querying portfolios: %v", err)
		return nil, err
	}
	defer rows.Close()

	var portfolios []*Portfolio
	for rows.Next() {
		var portfolio Portfolio
		if err := rows.Scan(&portfolio.ID, &portfolio.Title, &portfolio.Image, &portfolio.Content); err != nil {
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
	query := `UPDATE portfolio SET title = $2, image = $3, content = $4 WHERE id = $1`
	_, err := db.Exec(query, portfolio.ID, portfolio.Title, portfolio.Image, portfolio.Content)
	if err != nil {
		log.Printf("Error updating portfolio: %v", err)
		return err
	}
	return nil
}

// GetPortfolioByID retrieves a portfolio by its ID along with its associated skills
func GetPortfolioByID(db *sql.DB, portfolioID string) (*Portfolio, error) {
	portfolioQuery := `SELECT id, title, image, content FROM portfolio WHERE id = $1`
	row := db.QueryRow(portfolioQuery, portfolioID)

	var portfolio Portfolio
	err := row.Scan(&portfolio.ID, &portfolio.Title, &portfolio.Image, &portfolio.Content)
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
	portfolio.Skills = skills

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
