package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bShaak/habitui/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore provides persistence for habits using an embedded SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

func getDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "habit.db"
	}
	dir := filepath.Join(homeDir, ".habitui")
	return filepath.Join(dir, "habit.db")
}

// OpenSQLite opens (or creates) the default SQLite database and runs migrations.
func OpenSQLite() (*SQLiteStore, error) {
	return OpenSQLiteAt(getDBPath())
}

// OpenSQLiteAt opens (or creates) a SQLite database at dbPath and runs migrations.
func OpenSQLiteAt(dbPath string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func isDuplicateColumnErr(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate column name")
}

func (s *SQLiteStore) migrate() error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY
		);
	`); err != nil {
		return err
	}

	var version int
	if err := s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return err
	}

	migrations := []func() error{
		s.migrateV1,
		s.migrateV2,
		s.migrateV3,
	}
	for i, fn := range migrations {
		v := i + 1
		if version >= v {
			continue
		}
		if err := fn(); err != nil {
			return err
		}
		if _, err := s.db.Exec(`INSERT INTO schema_migrations(version) VALUES(?)`, v); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) migrateV1() error {
	if _, err := s.db.Exec(`
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
	`); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS completions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			habit_id INTEGER NOT NULL,
			completed_at TEXT NOT NULL,
			FOREIGN KEY (habit_id) REFERENCES habits (id) ON DELETE CASCADE
		);
	`); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_completions_habit_id_completed_at ON completions (habit_id, completed_at);
	`); err != nil {
		return err
	}

	_, err := s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_completions_completed_at ON completions (completed_at);
	`)
	return err
}

func (s *SQLiteStore) migrateV2() error {
	_, err := s.db.Exec(`ALTER TABLE habits ADD COLUMN color TEXT NOT NULL DEFAULT 'purple'`)
	if err != nil && !isDuplicateColumnErr(err) {
		return err
	}
	return nil
}

func (s *SQLiteStore) migrateV3() error {
	_, err := s.db.Exec(`ALTER TABLE habits ADD COLUMN icon TEXT NOT NULL DEFAULT ''`)
	if err != nil && !isDuplicateColumnErr(err) {
		return err
	}
	return nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error { return s.db.Close() }

func normalizeHabitDefaults(h *models.Habit) {
	if h.Frequency == "" {
		h.Frequency = "daily"
	}
	if h.Goal < 1 {
		h.Goal = 1
	}
	if h.Color == "" {
		h.Color = "purple"
	}
}

// CreateHabit inserts a new habit and returns the populated model with ID.
func (s *SQLiteStore) CreateHabit(ctx context.Context, h *models.Habit) (*models.Habit, error) {
	if h == nil {
		return nil, errors.New("habit is nil")
	}
	normalizeHabitDefaults(h)
	now := time.Now()
	if h.StartDate == "" {
		h.StartDate = now.Format(time.RFC3339)
	} else {
		t, err := time.Parse(time.RFC3339, h.StartDate)
		if err != nil {
			return nil, err
		}
		if t.IsZero() {
			h.StartDate = now.Format(time.RFC3339)
		}
	}
	h.CreatedAt = now.Format(time.RFC3339)
	h.UpdatedAt = now.Format(time.RFC3339)

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO habits(name, description, frequency, goal, color, icon, start_date, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		h.Name, h.Description, h.Frequency, h.Goal, h.Color, h.Icon, h.StartDate, h.CreatedAt, h.UpdatedAt)
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
	normalizeHabitDefaults(h)
	h.UpdatedAt = time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		UPDATE habits
		SET name = ?, description = ?, frequency = ?, goal = ?, color = ?, icon = ?, start_date = ?, updated_at = ?
		WHERE id = ?`,
		h.Name, h.Description, h.Frequency, h.Goal, h.Color, h.Icon, h.StartDate, h.UpdatedAt, h.ID)
	return err
}

// DeleteHabit removes a habit by ID and its completions.
func (s *SQLiteStore) DeleteHabit(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("invalid id")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM completions WHERE habit_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM habits WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// ListHabits returns all habits ordered by created_at.
func (s *SQLiteStore) ListHabits(ctx context.Context) ([]models.Habit, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, frequency, goal, color, icon, start_date, created_at, updated_at
		FROM habits
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Habit
	for rows.Next() {
		var h models.Habit
		if err := rows.Scan(
			&h.ID, &h.Name, &h.Description, &h.Frequency, &h.Goal, &h.Color, &h.Icon,
			&h.StartDate, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) CreateCompletion(ctx context.Context, c *models.Completion) (*models.Completion, error) {
	if c == nil {
		return nil, errors.New("completion is nil")
	}
	if c.CompletedAt == "" {
		c.CompletedAt = time.Now().Format(time.RFC3339)
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO completions(habit_id, completed_at)
		VALUES(?, ?)`,
		c.HabitID, c.CompletedAt)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	c.ID = id
	return c, nil
}

func (s *SQLiteStore) DeleteCompletion(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("invalid id")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM completions WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) ListCompletions(ctx context.Context) ([]models.Completion, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, habit_id, completed_at
		FROM completions
		ORDER BY completed_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCompletions(rows)
}

func dayBounds(date time.Time) (time.Time, time.Time) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
	return start, end
}

func scanCompletions(rows *sql.Rows) ([]models.Completion, error) {
	var completions []models.Completion
	for rows.Next() {
		var c models.Completion
		if err := rows.Scan(&c.ID, &c.HabitID, &c.CompletedAt); err != nil {
			return nil, err
		}
		completions = append(completions, c)
	}
	return completions, rows.Err()
}

// filterCompletionsInRange keeps completions whose parsed timestamp falls in [start, end].
// This avoids fragile lexicographic compares when offsets differ (local vs UTC).
func filterCompletionsInRange(completions []models.Completion, start, end time.Time) []models.Completion {
	var out []models.Completion
	for _, c := range completions {
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		if (completedAt.Equal(start) || completedAt.After(start)) &&
			(completedAt.Equal(end) || completedAt.Before(end)) {
			out = append(out, c)
		}
	}
	return out
}

func (s *SQLiteStore) queryCompletionsInPaddedRange(ctx context.Context, start, end time.Time, habitID *int64) ([]models.Completion, error) {
	var (
		rows *sql.Rows
		err  error
	)
	padStart := start.Add(-24 * time.Hour).Format(time.RFC3339)
	padEnd := end.Add(24 * time.Hour).Format(time.RFC3339)
	if habitID != nil {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, habit_id, completed_at
			FROM completions
			WHERE habit_id = ? AND completed_at >= ? AND completed_at <= ?`,
			*habitID, padStart, padEnd)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, habit_id, completed_at
			FROM completions
			WHERE completed_at >= ? AND completed_at <= ?`,
			padStart, padEnd)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	completions, err := scanCompletions(rows)
	if err != nil {
		return nil, err
	}
	return filterCompletionsInRange(completions, start, end), nil
}

func (s *SQLiteStore) GetCompletionsByHabitIDAndDate(ctx context.Context, habitID int64, date time.Time) ([]models.Completion, error) {
	start, end := dayBounds(date)
	return s.queryCompletionsInPaddedRange(ctx, start, end, &habitID)
}

func (s *SQLiteStore) GetCompletionsByHabitID(ctx context.Context, habitID int64) ([]models.Completion, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, habit_id, completed_at
		FROM completions
		WHERE habit_id = ?`,
		habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCompletions(rows)
}

func (s *SQLiteStore) GetCompletionsByDate(ctx context.Context, date time.Time) ([]models.Completion, error) {
	start, end := dayBounds(date)
	return s.queryCompletionsInPaddedRange(ctx, start, end, nil)
}

func (s *SQLiteStore) GetCompletionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Completion, error) {
	start, _ := dayBounds(startDate)
	_, end := dayBounds(endDate)
	return s.queryCompletionsInPaddedRange(ctx, start, end, nil)
}

var _ Store = (*SQLiteStore)(nil)
