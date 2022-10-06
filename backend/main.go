package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

type createTweet struct {
	Content string `json:"content"`
}

type Tweet struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
}

const tweetPrefix = "/api/tweet"

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE if not exists tweets (id serial primary key, content text not null);")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE if not exists words (id serial primary key, word text not null unique);")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE if not exists word_to_tweet (id serial primary key, word_id int not null, tweet_id int not null);")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pang")
	})

	http.HandleFunc(tweetPrefix, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			{
				createNewTweet(w, r, db)
				return
			}
		case http.MethodGet:
			{
				getTweet(w, r, db)
				return
			}

		}

		http.Error(w, "Bad request - Go away!", 405)

	})

	fmt.Print("Starting server\n")
	log.Fatal(http.ListenAndServe(":8081", nil))

}

func createNewTweet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var content createTweet
	err := json.NewDecoder(r.Body).Decode(&content)

	splitted := strings.Split(content.Content, " ")

	for _, word := range splitted {
		insertWord, _ := db.Prepare("INSERT INTO words (word) values ($1) ON CONFLICT DO NOTHING")
		_, err = insertWord.Exec(word)
	}

	stmt, err := db.Prepare("INSERT INTO tweets (content) values ($1) RETURNING id;")
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	var id int
	err = stmt.QueryRow(content.Content).Scan(&id)
	if err != nil {
		http.Error(w, "internal server error", 500)
	}

	for _, word := range splitted {
		insertWord, _ := db.Prepare("INSERT INTO word_to_tweet (word_id, tweet_id) values ((select id from words where word = $1), $2);")
		_, err = insertWord.Exec(word, id)
	}

	w.Write([]byte(strconv.Itoa(id)))
	return
}

func getTweet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")

	stmt, err := db.Prepare("SELECT id, content FROM tweets WHERE id = $1;")
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	var t Tweet
	err = stmt.QueryRow(id).Scan(&t.Id, &t.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "no such element", 404)
		}
		http.Error(w, "internal server error", 500)
	}
	w.Header().Set("Content-Type", "application/json")
	output, err := json.Marshal(t)
	w.Write(output)
}
