package main

import (
	"log"
	"my-project/connection"
	"my-project/titles"
	"net/http"
)

func main() {
	db := connection.ConnectDB()
	defer db.Close()
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/articles", titles.IndexHandler(db))
	http.HandleFunc("/article", titles.ArticleHandler(db))
	http.HandleFunc("/admin", titles.InputHandler(db))
	if err := http.ListenAndServe(":8888", nil); err != nil {
		log.Fatal("Error to start HTTP SERVER:", err)
	}
}
