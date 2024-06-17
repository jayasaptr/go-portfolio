package main

import (
	"database/sql"
	"errors"
)

func migrate(db *sql.DB) (sql.Result, error) {
	if db == nil {
		return nil, errors.New("database connection is unavailable")
	}

	return db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		password VARCHAR(255) NOT NULL,
		image TEXT,
		token TEXT
	);

	CREATE TABLE IF NOT EXISTS skills (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		image TEXT
	);

	CREATE TABLE IF NOT EXISTS portfolio (
		id VARCHAR(36) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		subtitle VARCHAR(255) NOT NULL,
		image TEXT,
		content TEXT,
		status VARCHAR(255),
		date_project DATE
	);

	CREATE TABLE IF NOT EXISTS experiance (
		id VARCHAR(36) PRIMARY KEY,
		company_name VARCHAR(255) NOT NULL,
		position VARCHAR(255) NOT NULL,
		image TEXT,
		start_date DATE,
		end_date DATE,
		location VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS experiance_skills (
		experiance_id VARCHAR(36) NOT NULL,
		skill_id VARCHAR(36) NOT NULL,
		FOREIGN KEY (experiance_id) REFERENCES experiance(id),
		FOREIGN KEY (skill_id) REFERENCES skills(id) ON UPDATE CASCADE ON DELETE RESTRICT,
		PRIMARY KEY (experiance_id, skill_id)
	);

	CREATE TABLE IF NOT EXISTS portfolio_skills (
		portfolio_id VARCHAR(36) NOT NULL,
		skill_id VARCHAR(36) NOT NULL,
		FOREIGN KEY (portfolio_id) REFERENCES portfolio(id),
		FOREIGN KEY (skill_id) REFERENCES skills(id) ON UPDATE CASCADE ON DELETE RESTRICT,
		PRIMARY KEY (portfolio_id, skill_id) 
	);

	CREATE TABLE IF NOT EXISTS portfolio_experience (
		portfolio_id VARCHAR(36) NOT NULL,
		experiance_id VARCHAR(36) NOT NULL,
		FOREIGN KEY (portfolio_id) REFERENCES portfolio(id),
		FOREIGN KEY (experiance_id) REFERENCES experiance(id) ON UPDATE CASCADE ON DELETE RESTRICT,
		PRIMARY KEY (portfolio_id, experiance_id) 
	);
	`)
}
