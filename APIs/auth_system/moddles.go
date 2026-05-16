package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Session struct {
	Id        string    `json:"id"`
	User_id   int       `json:"user_id"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}
type SessionStore struct {
	db *sql.DB
}

type UserStore struct {
	db *sql.DB
}

type Todo struct {
	Id     int    `json:"id"`
	Task   string `json:"task"`
	Status string `json:"status"`
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

func (t *TodoStore) Add(todo Todo, userID int) (Todo, error) {
	query := "INSERT INTO todos (task, status, user_id) VALUES($1, $2, $3) RETURNING id"
	err := t.db.QueryRow(query, todo.Task, todo.Status, userID).Scan(&todo.Id)
	if err != nil {
		return Todo{}, err
	}
	return todo, nil
}

func (t *TodoStore) GetAll(userID int) ([]Todo, error) {
	query := "SELECT id, task FROM todos WHERE user_id = $1"
	rows, err := t.db.Query(query, userID)
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

func (s *SessionStore) GenereateSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *SessionStore) CeateSession(userId int) (string, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	sessionId := s.GenereateSession()
	query := "INSERT INTO sessions (id, user_id, expires_at) VALUES($1, $2, $3)"

	_, err := s.db.Exec(query, sessionId, userId, expiresAt)
	if err != nil {
		return "", err
	}
	return sessionId, nil
}

func (s *SessionStore) GetUserID(sessionId string) (int, error) {
	var userId int
	var expiresAt time.Time

	query := "SELECT user_id, expires_at FROM sessions WHERE id = $1"
	err := s.db.QueryRow(query, sessionId).Scan(&userId, &expiresAt)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (s *SessionStore) Delete(sessionId string) error {
	query := "DELETE FROM sessions WHERE id = $1"
	_, err := s.db.Exec(query, sessionId)
	return err
}

func SessionCookie(w http.ResponseWriter, sessionId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
