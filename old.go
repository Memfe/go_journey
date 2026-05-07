package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
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
}
type TodoStore struct {
	todos  map[int]Todo
	mu     sync.Mutex
	NextID int
}

func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos:  make(map[int]Todo),
		NextID: 0,
	}
}

func (s *TodoStore) Add(todo Todo) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.NextID += 1
	todo.ID = s.NextID
	s.todos[todo.ID] = todo
	return todo
}

func (s *TodoStore) Get() []Todo {
	var todos []Todo
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.todos {
		todos = append(todos, v)
	}
	return todos
}

func (s *TodoStore) delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.todos[id]
	if !ok {
		return errors.New("Todo does not exist")
	}
	delete(s.todos, id)
	return nil
}

func (s *TodoStore) getById(id int) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	todo, ok := s.todos[id]
	if !ok {
		return Todo{}, errors.New("Todo does not exist")
	}
	return todo, nil
}

func (s *TodoStore) put(id int, todo Todo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.todos[id]
	if !ok {
		return errors.New("Todo does not exist")
	}
	s.todos[id] = todo
	return nil
}

func (s *TodoStore) patch(id int, patchTodo PatchTodo) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	todo, ok := s.todos[id]
	if !ok {
		return Todo{}, errors.New("Todo does not exist")
	}

	if patchTodo.Task != nil {
		todo.Task = *patchTodo.Task
	}
	if patchTodo.Status != nil {
		todo.Status = *patchTodo.Status
	}
	s.todos[id] = todo

	return todo, nil
}

func main() {
	store := NewTodoStore()
	handler := &Handler{store: store}

	http.HandleFunc("/", healthHandler)
	http.HandleFunc("PATCH /todos/{id}", handler.Patch)
	http.HandleFunc("PUT /todos/{id}", handler.Put)
	http.HandleFunc("GET /todos/{id}", handler.GetById)
	http.HandleFunc("DELETE /todos/{id}", handler.Delete)
	http.HandleFunc("POST /todos", handler.createTodo)
	http.HandleFunc("GET /todos", handler.getTodos)
	log.Println("Listening on port 8000")

	http.ListenAndServe(":8000", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is the health checker")
}

func (h *Handler) createTodo(w http.ResponseWriter, r *http.Request) {
	todo := Todo{}
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	todo = h.store.Add(todo)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (h *Handler) getTodos(w http.ResponseWriter, r *http.Request) {
	todos := h.store.Get()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(todos)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
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
