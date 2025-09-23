package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	types "github.com/bShaak/habitui/internal/models"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type DBClient struct {
	DB *sql.DB
}

func NewDBClient() *DBClient {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get the database URL and token from environment variables
	url := os.Getenv("DB_URL")
	token := os.Getenv("DB_TOKEN")

	// Construct the full database URL with the token
	fullURL := fmt.Sprintf("%s?authToken=%s", url, token)

	// Open the database connection
	db, err := sql.Open("libsql", fullURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", fullURL, err)
		os.Exit(1)
	}

	// Test the database connection
	err = db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db %s: %s", fullURL, err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to the database!")

	return &DBClient{DB: db}
}

func (client *DBClient) Close() {
	client.DB.Close()
}

func (client *DBClient) GetHabits() ([]types.Habit, error) {
	rows, err := client.DB.Query("SELECT name, completed FROM habits")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []types.Habit
	for rows.Next() {
		var h types.Habit
		if err := rows.Scan(&h.Name, &h.Completed); err != nil {
			return nil, err
		}
		habits = append(habits, h)
	}
	return habits, nil
}
