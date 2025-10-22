package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bShaak/habitui/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore provides persistence for habits using an embedded SQLite database.
type SQLiteStore struct {
	DB *sql.DB
}

// OpenSQLite opens (or creates) a SQLite database at the provided path and runs migrations.
func OpenSQLite() (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", "./habit.db")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &SQLiteStore{DB: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) migrate() error {
	// Create habits table
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS habits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			frequency TEXT NOT NULL DEFAULT 'daily',
			start_date TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			archived_at TEXT NULL
		);
	`)
	if err != nil {
		return err
	}

	// Create habit_logs table
	_, err = s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS habit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			habit_id INTEGER NOT NULL,
			timestamp TEXT NOT NULL,
			FOREIGN KEY (habit_id) REFERENCES habits (id) ON DELETE CASCADE
		);
	`)
	return err
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error { return s.DB.Close() }

// CreateHabit inserts a new habit and returns the populated model with ID.
func (s *SQLiteStore) CreateHabit(ctx context.Context, h *models.Habit) (*models.Habit, error) {
	if h == nil {
		return nil, errors.New("habit is nil")
	}
	now := time.Now().UTC()
	if h.StartDate.IsZero() {
		h.StartDate = now
	}
	h.CreatedAt = now
	h.UpdatedAt = now

	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO habits(name, description, frequency, start_date, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)`,
		h.Name, h.Description, h.Frequency, h.StartDate.Format(time.RFC3339), h.CreatedAt.Format(time.RFC3339), h.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	h.ID = id
	return h, nil
}

// UpdateHabit updates editable fields on a habit by ID.
func (s *SQLiteStore) UpdateHabit(ctx context.Context, h *models.Habit) error {
	if h == nil || h.ID == 0 {
		return errors.New("invalid habit")
	}
	h.UpdatedAt = time.Now().UTC()
	_, err := s.DB.ExecContext(ctx, `
		UPDATE habits
		SET name = ?, description = ?, frequency = ?, start_date = ?, updated_at = ?
		WHERE id = ?`,
		h.Name, h.Description, h.Frequency, h.StartDate.Format(time.RFC3339), h.UpdatedAt.Format(time.RFC3339), h.ID)
	return err
}

// DeleteHabit removes a habit by ID.
func (s *SQLiteStore) DeleteHabit(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("invalid id")
	}
	_, err := s.DB.ExecContext(ctx, `DELETE FROM habits WHERE id = ?`, id)
	return err
}

// ListHabits returns all habits ordered by created_at.
func (s *SQLiteStore) ListHabits(ctx context.Context) ([]models.Habit, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, description, frequency, start_date, created_at, updated_at
		FROM habits
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Habit
	for rows.Next() {
		var h models.Habit
		var startDateStr, createdAtStr, updatedAtStr string
		if err := rows.Scan(&h.ID, &h.Name, &h.Description, &h.Frequency, &startDateStr, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		var err1, err2, err3 error
		h.StartDate, err1 = time.Parse(time.RFC3339, startDateStr)
		h.CreatedAt, err2 = time.Parse(time.RFC3339, createdAtStr)
		h.UpdatedAt, err3 = time.Parse(time.RFC3339, updatedAtStr)
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, errors.New("failed to parse timestamp")
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
