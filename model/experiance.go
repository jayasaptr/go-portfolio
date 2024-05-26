package model

import (
	"database/sql"
	"log"
	"time"
)

type Experience struct {
	ID string `json:"id"`
	CompanyName string `json:"company_name"`
	Position string `json:"position"`
	Image string `json:"image"`
	StartDate time.Time `json:"start_date"`
	EndDate time.Time `json:"end_date"`
	Location string `json:"location"`
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
	if _, err :=  tx.Exec(deleteQuery, experianceID); err != nil{
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

func GetExperience(db *sql.DB, offset int, limit int)([]*Experience, error){
	query := `SELECT id, company_name, position, image, start_date, end_date, location FROM experiance LIMIT $1 OFFSET $2`

	rows, err := db.Query(query, limit, offset)

	if  err != nil {
		log.Printf("Error querying experience: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var experiances []*Experience
	for rows.Next(){
		var experience Experience
		if err := rows.Scan(&experience.ID, &experience.CompanyName, &experience.Position,&experience.Image, &experience.StartDate, &experience.EndDate, &experience.Location); err != nil {
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
	err := row.Scan(&experience.ID, &experience.CompanyName, &experience.Image, &experience.Position,&experience.StartDate, &experience.StartDate, &experience.Location)
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