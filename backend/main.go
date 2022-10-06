package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

type createTweet struct {
	content string
}

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE tweets (id int primary key auto increment, content text);")

	if err != nil {
		panic(err)
	}

	http.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pang")
	})

	http.HandleFunc("/api/tweet", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			createNewTweet(w, r, db)
			return
		}

		http.Error(w, "Bad request - Go away!", 405)

	})

	fmt.Print("Starting server\n")
	log.Fatal(http.ListenAndServe(":8081", nil))

}

func createNewTweet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var content createTweet
	err := json.NewDecoder(r.Body).Decode(&content)

	stmt, err := db.Prepare("INSERT INTO tweets (content) value (?)")
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	_, err = stmt.Exec(content.content)
	if err != nil {
		http.Error(w, "internal server error", 500)
	}

	return
}
