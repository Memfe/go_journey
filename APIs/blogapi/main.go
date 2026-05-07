package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int       `json:"userId"`
}

type Comment struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	UserID  int    `json:"userId"`
	PostID  int    `json:"postId"`
}

func main() {
	fmt.Println("Hello Myo")
	http.HandleFunc("/", healthHandler)

	log.Println("Listening on port 8000")
	http.ListenAndServe(":8000", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("the blog api in the mix"))
}
