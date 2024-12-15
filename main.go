package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var db *sql.DB

type Config struct {
	App struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"app"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Dbname   string `mapstructure:"dbname"`
		Sslmode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`
}

func main() {
	var c Config
	configFile := "config.yaml"
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err = viper.Unmarshal(&c)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Dbname)

	// open database
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

	log.Printf("Starting server on :%d", c.App.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.App.Port), nil); err != nil {
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
