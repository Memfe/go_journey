package main

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserStore struct {
	db *sql.DB
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func ComparePasswords(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func (u *UserStore) CreateUser(email, password string) (User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	var user User
	query := "INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id"
	err = u.db.QueryRow(query, email, hash).Scan(&user.ID)
	if err != nil {
		return User{}, err
	}
	user.Email = email
	return user, nil
}

func (u *UserStore) GetByEmail(email string) (User, string, error) {
	var user User
	var hash string

	query := "SELECT id, email, password_hash FROM users WHERE email=$1"
	err := u.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &hash)

	return user, hash, err
}
