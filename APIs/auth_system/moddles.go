package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserUpdate struct {
	Email *string `json:"email"`
	Name  *string `json:"name"`
	Bio   *string `json:"bio"`
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
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (t *TodoStore) Add(todo Todo, userID int) (Todo, error) {
	query := "INSERT INTO todos (task, status, user_id) VALUES($1, $2, $3) RETURNING id"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := t.db.QueryRowContext(ctx, query, todo.Task, todo.Status, userID).Scan(&todo.Id)
	if err != nil {
		return Todo{}, err
	}
	return todo, nil
}

func (t *TodoStore) GetAll(userID int) ([]Todo, error) {
	query := "SELECT id, task FROM todos WHERE user_id = $1"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := t.db.QueryContext(ctx, query, userID)
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

func (u *UserStore) CreateUser(email, password, name, bio string) (User, error) {
	var user User
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	query := "INSERT INTO users(email, password_hash, name, bio) VALUES($1, $2, $3, $4) RETURNING id, email, name, bio"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = u.db.QueryRowContext(ctx, query, email, hash, name, bio).Scan(&user.ID, &user.Email, &user.Name, &user.Bio)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (u *UserStore) GetByEmail(email string) (User, error) {
	var user User

	query := "SELECT id, email, password_hash FROM users WHERE email=$1"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := u.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password)

	return user, err
}

func (u *UserStore) GetByID(id int) (User, error) {
	var user User
	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, email, name, bio, created_at FROM users WHERE id=$1"
	err := u.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.Name, &user.Bio, &user.CreatedAt)
	return user, err
}

func (u *UserStore) UpdateUser(userId int, updatedUser UserUpdate) (User, error) {
	user, err := u.GetByID(userId)
	if err != nil {
		return User{}, err
	}

	if updatedUser.Email != nil {
		user.Email = *updatedUser.Email
	}
	if updatedUser.Name != nil {
		user.Name = *updatedUser.Name
	}
	if updatedUser.Bio != nil {
		user.Bio = *updatedUser.Bio
	}
	query := `UPDATE users SET email+$1, name=$2, bio=$3 WHERE id=$4`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = u.db.QueryRowContext(ctx, query, user.ID).Scan(&user.Email, &user.Name, &user.Bio)
	return user, err
}

func (s *SessionStore) GenerateSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *SessionStore) CreateSession(userId int) (string, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	sessionId := s.GenerateSession()
	query := "INSERT INTO sessions (id, user_id, expires_at) VALUES($1, $2, $3)"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, sessionId, userId, expiresAt)
	if err != nil {
		return "", err
	}
	return sessionId, nil
}

func (s *SessionStore) GetUserID(sessionId string) (int, error) {
	var userId int

	query := "SELECT user_id FROM sessions WHERE id = $1 AND expires_at>NOW()"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, sessionId).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (s *SessionStore) Delete(sessionId string) error {
	query := "DELETE FROM sessions WHERE id = $1"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, sessionId)
	return err
}

func SessionCookie(w http.ResponseWriter, sessionId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400,
		//Secure:   true,
		SameSite: http.SameSiteDefaultMode,
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

func (s *SessionStore) CleanupSession() error {
	query := "DELETE FROM sessions WHERE expires_at<=NOW()"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}
