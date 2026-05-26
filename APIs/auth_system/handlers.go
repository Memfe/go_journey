package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type AuthHandler struct {
	store    *UserStore
	todos    *TodoStore
	sessions *SessionStore
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"-"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
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

	v := New()

	v.Check(len(req.Password) < 8, "Password", "Password must be 8 or more characters")
	v.Check(IsEmail(req.Email), "Email", "Check the email pattern")
	v.Check(req.Name != "", "name", "name must not be empty")
	v.Check(req.Bio != "", "bio", "bio must not be empty")

	if !v.Valid() {
		v.ErrorHelper(w)
		return
	}

	user, err := a.store.CreateUser(req.Email, req.Password, req.Name, req.Bio)
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

	user, err := a.store.GetByEmail(req.Email)
	if err == sql.ErrNoRows {
		log.Println(err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = ComparePassword(user.Password, req.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	sessionID, err := a.sessions.CreateSession(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to create session", 500)
		return
	}

	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "session_id",
	// 	Value:    sessionID,
	// 	HttpOnly: true,
	// 	Path:     "/",
	// })

	SessionCookie(w, sessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "login successful",
		"user":    user.Email,
	})
}

func (a *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromContext(r)

	user, err := a.store.GetByID(userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (a *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromContext(r)

	var user UserUpdate
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	updatedUser, err := a.store.UpdateUser(userID, user)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

func (a *AuthHandler) GetTodos(w http.ResponseWriter, r *http.Request) {

	userID, _ := GetUserIDFromContext(r)

	todos, err := a.todos.GetAll(userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "No todos please", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(todos)
}

func (a *AuthHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {

	userID, _ := GetUserIDFromContext(r)

	defer r.Body.Close()
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "failed to parse todo", http.StatusBadRequest)
		return
	}

	v := New()
	v.Check(todo.Task != "", "task", "todo must not be empty")
	v.Check(todo.Status != "", "status", "status required")
	if !v.Valid() {
		v.ErrorHelper(w)
		return
	}

	todo, err = a.todos.Add(todo, userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to add todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "no session found", http.StatusUnauthorized)
		return
	}

	err = a.sessions.Delete(cookie.Value)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}
	ClearSessionCookie(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "logout successful",
	})
}
