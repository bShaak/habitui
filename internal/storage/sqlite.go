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
			goal INTEGER NOT NULL DEFAULT 1,
			start_date TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			archived_at TEXT NULL
		);
	`)
	if err != nil {
		return err
	}

	// Create completions table
	_, err = s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS completions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			habit_id INTEGER NOT NULL,
			completed_at TEXT NOT NULL,
			FOREIGN KEY (habit_id) REFERENCES habits (id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return err
	}

	// Create indexes for faster completions queries
	_, err = s.DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_completions_habit_id_completed_at ON completions (habit_id, completed_at);
	`)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_completions_completed_at ON completions (completed_at);
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
	t, err := time.Parse(time.RFC3339, h.StartDate)
	if err != nil {
		return nil, err
	}
	if t.IsZero() {
		h.StartDate = now.Format(time.RFC3339)
	}
	h.CreatedAt = now.Format(time.RFC3339)
	h.UpdatedAt = now.Format(time.RFC3339)

	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO habits(name, description, frequency, goal, start_date, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)`,
		h.Name, h.Description, h.Frequency, h.Goal, h.StartDate, h.CreatedAt, h.UpdatedAt)
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
	h.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.ExecContext(ctx, `
		UPDATE habits
		SET name = ?, description = ?, frequency = ?, goal = ?, start_date = ?, updated_at = ?
		WHERE id = ?`,
		h.Name, h.Description, h.Frequency, h.Goal, h.StartDate, h.UpdatedAt, h.ID)
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
		SELECT id, name, description, frequency, goal, start_date, created_at, updated_at
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
		if err := rows.Scan(&h.ID, &h.Name, &h.Description, &h.Frequency, &h.Goal, &startDateStr, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		h.StartDate = startDateStr
		h.CreatedAt = createdAtStr
		h.UpdatedAt = updatedAtStr
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) CreateCompletion(ctx context.Context, c *models.Completion) (*models.Completion, error) {
	if c == nil {
		return nil, errors.New("completion is nil")
	}
	c.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO completions(habit_id, completed_at)
		VALUES(?, ?)`,
		c.HabitID, c.CompletedAt)
	return c, err
}

func (s *SQLiteStore) DeleteCompletion(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("invalid id")
	}
	_, err := s.DB.ExecContext(ctx, `DELETE FROM completions WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) ListCompletions(ctx context.Context) ([]models.Completion, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, habit_id, completed_at
		FROM completions
		ORDER BY completed_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var completions []models.Completion
	for rows.Next() {
		var c models.Completion
		var completedAtStr string
		if err := rows.Scan(&c.ID, &c.HabitID, &completedAtStr); err != nil {
			return nil, err
		}
		c.CompletedAt = completedAtStr
		completions = append(completions, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return completions, nil
}

func (s *SQLiteStore) GetCompletionsByHabitIdAndDate(ctx context.Context, habitId int64, date time.Time) ([]models.Completion, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, habit_id, completed_at
		FROM completions
		WHERE habit_id = ? AND completed_at = ?`,
		habitId, date.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var completions []models.Completion
	for rows.Next() {
		var c models.Completion
		var completedAtStr string
		if err := rows.Scan(&c.ID, &c.HabitID, &completedAtStr); err != nil {
			return nil, err
		}
		c.CompletedAt = completedAtStr
		completions = append(completions, c)
	}
	return completions, rows.Err()
}

func (s *SQLiteStore) GetCompletionsByHabitId(ctx context.Context, habitId int64) ([]models.Completion, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, habit_id, completed_at
		FROM completions
		WHERE habit_id = ?`,
		habitId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var completions []models.Completion
	for rows.Next() {
		var c models.Completion
		var completedAtStr string
		if err := rows.Scan(&c.ID, &c.HabitID, &completedAtStr); err != nil {
			return nil, err
		}
		c.CompletedAt = completedAtStr
		completions = append(completions, c)
	}
	return completions, rows.Err()
}

func (s *SQLiteStore) GetCompletionsByDate(ctx context.Context, date time.Time) ([]models.Completion, error) {
	// Get the start of the day (00:00:00) and end of the day (23:59:59.999999999)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, habit_id, completed_at
		FROM completions
		WHERE completed_at >= ? AND completed_at <= ?`,
		startOfDay.Format(time.RFC3339),
		endOfDay.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var completions []models.Completion
	for rows.Next() {
		var c models.Completion
		var completedAtStr string
		if err := rows.Scan(&c.ID, &c.HabitID, &completedAtStr); err != nil {
			return nil, err
		}
		c.CompletedAt = completedAtStr
		completions = append(completions, c)
	}
	return completions, rows.Err()
}