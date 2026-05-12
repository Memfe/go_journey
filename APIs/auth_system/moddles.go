package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}
type SessionStore struct {
	sessions map[string]int
}

type UserStore struct {
	db *sql.DB
}

type Todo struct {
	Id   int    `json:"id"`
	Task string `json:"task"`
}

type TodoStore struct {
	db *sql.DB
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (t *TodoStore) Add(todo Todo) (Todo, error) {
	query := "INSERT INTO todos (task) VALUES($1) RETURNING id"
	err := t.db.QueryRow(query, todo.Task).Scan(&todo.Id)
	if err != nil {
		return Todo{}, err
	}
	return todo, nil
}

func (t *TodoStore) GetAll() ([]Todo, error) {
	query := "SELECT id, task FROM todos"
	rows, err := t.db.Query(query)
	if err != nil {
		return []Todo{}, err
	}
	defer rows.Close()
	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.Id, &todo.Task)
		if err != nil {
			return []Todo{}, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (u *UserStore) CreateUser(email, password string) (User, error) {
	var user User
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	query := "INSERT INTO user_testing(email, password_hash) VALUES($1, $2) RETURNING id"
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
	query := "SELECT id, email, password_hash FROM user_testing WHERE email=$1"
	err := u.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &hash)

	return user, hash, err
}

func GenerateSessionId() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
