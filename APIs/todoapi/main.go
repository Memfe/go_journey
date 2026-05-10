package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Handler struct {
	store *TodoStore
}

type PatchTodo struct {
	Task   *string `json:"task"`
	Status *string `json:"status"`
}

type Todo struct {
	ID     int    `json:"id"`
	Task   string `json:"task"`
	Status string `json:"status"`
	//User_id  int    `json:user_id"`
}
type TodoStore struct {
	db *sql.DB
}

type AuthHandler struct {
	users *UserStore
}

func (s *TodoStore) Add(todo Todo) (Todo, error) {
	query := "INSERT INTO todos(task, status) VALUES($1,$2)RETURNING id"
	err := s.db.QueryRow(query, todo.Task, todo.Status).Scan(&todo.ID)
	if err != nil {
		return Todo{}, err
	}
	return todo, nil
}

func (s *TodoStore) Get() ([]Todo, error) {
	query := "SELECT id, task, status FROM todos"
	row, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var todos []Todo
	for row.Next() {
		var t Todo
		err := row.Scan(&t.ID, &t.Task, &t.Status)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (s *TodoStore) delete(id int) error {
	query := "DELETE FROM todos WHERE id = $1"
	result, err := s.db.Exec(query, id)

	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *TodoStore) getById(id int) (Todo, error) {
	query := "SELECT id, task, status FROM todos WHERE id = $1"
	var todo Todo
	err := s.db.QueryRow(query, id).Scan(&todo.ID, &todo.Task, &todo.Status)
	if err != nil {
		return Todo{}, err
	}
	return todo, nil
}

func (s *TodoStore) put(id int, todo Todo) error {
	query := "UPDATE todos SET task=$1, status=$2 WHERE id=$3"
	result, err := s.db.Exec(query, todo.Task, todo.Status, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *TodoStore) patch(id int, patchTodo PatchTodo) (Todo, error) {
	todo, err := s.getById(id)
	if err != nil {
		return Todo{}, err
	}

	if patchTodo.Task != nil {
		todo.Task = *patchTodo.Task
	}
	if patchTodo.Status != nil {
		todo.Status = *patchTodo.Status
	}
	err = s.put(id, todo)
	if err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func main() {
	db, err := sql.Open("pgx", "postgres://memfe:annansie@localhost:5432/myapp?sslmode=disable")
	if err != nil {
		log.Fatalln("Failed to open db", err)
	}
	store := &TodoStore{db: db}
	handler := &Handler{store: store}

	//http.HandleFunc("/", healthHandler)
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /todos/{id}", handler.Patch)
	mux.HandleFunc("PUT /todos/{id}", handler.Put)
	mux.HandleFunc("GET /todos/{id}", handler.GetById)
	mux.HandleFunc("DELETE /todos/{id}", handler.Delete)
	mux.HandleFunc("POST /todos", handler.createTodo)
	mux.HandleFunc("GET /todos", handler.getTodos)

	authStore := &UserStore{db: db}
	authHandler := &AuthHandler{users: authStore}

	mux.HandleFunc("POST /todos/signup", authHandler.Signup)
	mux.HandleFunc("POST /todos/login", authHandler.Login)

	wrapper := Chain(mux, Logging, Cors)

	log.Println("Listening on port 8000")

	http.ListenAndServe(":8000", wrapper)
}

// func healthHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "This is the health checker")
// }

func (h *Handler) createTodo(w http.ResponseWriter, r *http.Request) {
	todo := Todo{}
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	todo, err = h.store.Add(todo)
	if err != nil {
		http.Error(w, "failed to create todo", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (h *Handler) getTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(todos)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")

	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.store.delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{fmt.Sprintf("todo of %d", id): "deleted"})
}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")

	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	todo, err := h.store.getById(id)
	if err != nil {
		http.Error(w, "id does not exist", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var updatedTodo Todo
	json.NewDecoder(r.Body).Decode(&updatedTodo)

	updatedTodo.ID = id
	err = h.store.put(id, updatedTodo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTodo)
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")

	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, "Enter a proper ID", http.StatusBadRequest)
		return
	}
	var patchedTodo PatchTodo
	json.NewDecoder(r.Body).Decode(&patchedTodo)
	updated, err := h.store.patch(id, patchedTodo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (u *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := u.users.CreateUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "Application/json")
	json.NewEncoder(w).Encode(user)
}

func (u *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, hash, err := u.users.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid email", http.StatusUnauthorized)
		return
	}

	err = ComparePasswords(hash, req.Password)
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "login successful",
		"user":    user.Email,
	})
}
