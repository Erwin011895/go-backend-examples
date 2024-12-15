package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

// database connection
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "87654321"
	dbname   = "postgres"
)

var db *sql.DB

func main() {
	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// open database
	var err error
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Database connected!")

	http.HandleFunc("POST /insert", insertNoteHandler)
	http.HandleFunc("GET /select", getNotesHandler)

	webport := "8081"
	log.Println("Starting server on :" + webport)
	if err := http.ListenAndServe(":"+webport, nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}

type Note struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"` // nullable
}

func insertNoteHandler(w http.ResponseWriter, r *http.Request) {
	var note struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		log.Println(err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// create query
	query := `INSERT INTO notes(title, description, created_at, updated_at, deleted_at) VALUES($1, $2, now(), now(), null)`
	_, err := db.Exec(query, note.Title, note.Description)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(note); err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func getNotesHandler(w http.ResponseWriter, r *http.Request) {
	var notes []Note

	// query to get all notes
	query := `SELECT id, title, description, created_at, updated_at, deleted_at FROM notes`

	// execute query
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// iterate result
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Description, &note.CreatedAt, &note.UpdatedAt, &note.DeletedAt)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	// encode to json
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
