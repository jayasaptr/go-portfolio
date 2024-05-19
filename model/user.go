package model

import (
	"database/sql"
	"errors"
)

type User struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email" gorm:"unique"`
	Password string  `json:"password,omitempty"`
	Image    string  `json:"image,omitempty"`
	Token    *string `json:"token,omitempty"` // Token can be null
}

var (
	ErrDBNil = errors.New("koneksi tidak tersedia")
)

func InsertUser(db *sql.DB, user User) error {
	if db == nil {
		return ErrDBNil
	}

	query := `INSERT INTO users (id, name, email, password, image) VALUES ($1, $2, $3, $4, $5);`
	_, err := db.Exec(query, user.ID, user.Name, user.Email, user.Password, user.Image)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(db *sql.DB, user User) error {
	if db == nil {
		return ErrDBNil
	}

	query := `UPDATE users SET name=$2, email=$3, password=$4, image=$5, token=$6 WHERE id=$1;`
	_, err := db.Exec(query, user.ID, user.Name, user.Email, user.Password, user.Image, user.Token)
	if err != nil {
		return err
	}

	return nil
}

func GetUserID(db *sql.DB, userID string) (*User, error) {
	if db == nil {
		return nil, ErrDBNil
	}

	query := `SELECT id, name, email, image, token FROM users WHERE id = $1;`
	row := db.QueryRow(query, userID)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Image, &user.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no user found")
		}
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(db *sql.DB, userEmail string) (*User, error) {
	if db == nil {
		return nil, ErrDBNil
	}

	query := `SELECT id, name, email, image, password FROM users WHERE email = $1;`
	row := db.QueryRow(query, userEmail)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Image, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no user found with the specified email")
		}
		return nil, err
	}

	return &user, nil
}

func DeleteUser(db *sql.DB, userID string) error {
	if db == nil {
		return ErrDBNil
	}

	query := `DELETE FROM users WHERE id = $1;`
	_, err := db.Exec(query, userID)
	if err != nil {
		return err
	}

	return nil
}

