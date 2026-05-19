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
		db: db,
	}
	store := &UserStore{db: db}
	todos := &TodoStore{db: db}
	authHandler := &AuthHandler{store: store, sessions: sessions, todos: todos}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", authHandler.Signup)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.Handle("GET /me", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.Handle("POST /create", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.CreateTodo)))
	mux.Handle("GET /get", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.GetTodos)))

	log.Println("Listening on Port 8000")
	http.ListenAndServe(":8000", mux)
}
