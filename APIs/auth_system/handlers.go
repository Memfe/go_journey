package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type AuthHandler struct {
	store        *UserStore
	sessionStore *SessionStore
	todos        *TodoStore
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if len(req.Password) < 6 {
		http.Error(w, "weak password", http.StatusBadRequest)
		return
	}
	user, err := a.store.CreateUser(req.Email, req.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to parse json", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	user, hash, err := a.store.GetByEmail(req.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = ComparePassword(hash, req.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	sessionID := GenerateSessionId()
	a.sessionStore.sessions[sessionID] = user.ID

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "login successful",
		"user":    user.Email,
	})
}

func (a *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := a.sessionStore.sessions[cookie.Value]
	if !ok {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{
		"user_id": userID,
	})
}

func (a *AuthHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_, ok := a.sessionStore.sessions[cookie.Value]
	if !ok {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	todos, err := a.todos.GetAll()
	if err != nil {
		http.Error(w, "garbage pleace", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(todos)
}

func (a *AuthHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_, ok := a.sessionStore.sessions[cookie.Value]
	if !ok {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	var todo Todo
	err = json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "failed to parse todo", http.StatusBadRequest)
		return
	}

	todo, err = a.todos.Add(todo)
	if err != nil {
		http.Error(w, "failed to add todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}
