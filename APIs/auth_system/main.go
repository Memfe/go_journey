package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://memfe:annansie@localhost:5432/myapp?sslmode=disable")
	if err != nil {
		log.Fatalln("Failed to open db", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	sessions := &SessionStore{
		sessions: make(map[string]int),
	}
	store := &UserStore{db: db}
	todos := &TodoStore{db: db}
	authHandler := &AuthHandler{store: store, sessionStore: sessions, todos: todos}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", authHandler.Signup)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /me", authHandler.Me)
	mux.HandleFunc("POST /create", authHandler.CreateTodo)
	mux.HandleFunc("GET /get", authHandler.GetTodos)

	log.Println("Listening on Port 8000")
	http.ListenAndServe(":8000", mux)
}
