package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

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

	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			err := authHandler.sessions.CleanupSession()
			if err != nil {
				log.Println(err)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", authHandler.Signup)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /logout", authHandler.Logout)

	mux.Handle("GET /profile", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.Profile)))
	mux.Handle("POST /updateprofile", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.UpdateProfile)))
	mux.Handle("POST /create", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.CreateTodo)))
	mux.Handle("GET /get", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.GetTodos)))

	log.Println("Listening on Port 8000")
	http.ListenAndServe(":8000", mux)
}
