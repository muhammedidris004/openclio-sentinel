package mem0style

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore persists Mem0-style facts in SQLite.
type SQLiteStore struct {
	db          *sql.DB
	minSalience float64
}

// NewSQLiteStore creates a SQLite-backed Mem0-style store.
func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("sqlite dsn is required")
	}

	if dsn != ":memory:" {
		dir := filepath.Dir(dsn)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("create sqlite dir: %w", err)
			}
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &SQLiteStore{
		db:          db,
		minSalience: defaultMinSalience,
	}
	if err := s.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying DB connection.
func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// SetMinSalience updates the minimum salience floor used by Upsert.
func (s *SQLiteStore) SetMinSalience(v float64) {
	if s == nil {
		return
	}
	if v <= 0 {
		s.minSalience = defaultMinSalience
		return
	}
	s.minSalience = clampUnit(v)
}

// Upsert inserts or updates one consolidated fact by normalized claim+category.
func (s *SQLiteStore) Upsert(input FactInput) (*Fact, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("nil sqlite store")
	}
	claim := normalizeText(input.Claim)
	if claim == "" {
		return nil, nil
	}
	category := normalizeText(input.Category)
	if category == "" {
		category = "unknown"
	}
	salience := clampUnit(input.Salience)
	if salience < s.minSalience {
		salience = s.minSalience
	}
	now := time.Now().UTC()
	key := factKey(claim, category)

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var (
		id           int64
		existingSal  float64
		existingNull sql.NullString
		exists       bool
	)
	row := tx.QueryRow(`
		SELECT id, salience, valid_until
		FROM mem0_facts
		WHERE fact_key = ?
	`, key)
	switch scanErr := row.Scan(&id, &existingSal, &existingNull); scanErr {
	case nil:
		exists = true
	case sql.ErrNoRows:
		exists = false
	default:
		return nil, fmt.Errorf("query existing fact: %w", scanErr)
	}

	if exists {
		if salience < existingSal {
			salience = existingSal
		}
		validUntil := existingNull
		if input.ValidUntil != nil {
			validUntil = sql.NullString{String: input.ValidUntil.UTC().Format(time.RFC3339Nano), Valid: true}
		}
		if _, err := tx.Exec(`
			UPDATE mem0_facts
			SET salience = ?, valid_until = ?, updated_at = ?
			WHERE id = ?
		`, salience, validUntil, now.Format(time.RFC3339Nano), id); err != nil {
			return nil, fmt.Errorf("update fact: %w", err)
		}
	} else {
		var validUntil any
		if input.ValidUntil != nil {
			validUntil = input.ValidUntil.UTC().Format(time.RFC3339Nano)
		}
		res, err := tx.Exec(`
			INSERT INTO mem0_facts(
				fact_key, claim, category, salience, valid_until, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, key, claim, category, salience, validUntil, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
		if err != nil {
			return nil, fmt.Errorf("insert fact: %w", err)
		}
		id, err = res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("fact last insert id: %w", err)
		}
	}

	fact, err := s.getByIDTx(tx, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return fact, nil
}

// Get returns one fact by normalized claim+category key.
func (s *SQLiteStore) Get(claim, category string) (*Fact, bool, error) {
	if s == nil || s.db == nil {
		return nil, false, fmt.Errorf("nil sqlite store")
	}
	claim = normalizeText(claim)
	if claim == "" {
		return nil, false, nil
	}
	category = normalizeText(category)
	if category == "" {
		category = "unknown"
	}
	row := s.db.QueryRow(`
		SELECT id, claim, category, salience, valid_until, created_at, updated_at
		FROM mem0_facts
		WHERE fact_key = ?
	`, factKey(claim, category))
	f, err := scanFactRow(row)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get fact: %w", err)
	}
	return f, true, nil
}

// List returns top facts ordered by salience and recency.
func (s *SQLiteStore) List(limit int) ([]*Fact, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("nil sqlite store")
	}
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, claim, category, salience, valid_until, created_at, updated_at
		FROM mem0_facts
		ORDER BY salience DESC, updated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list facts: %w", err)
	}
	defer rows.Close()
	return scanFacts(rows)
}

// Search returns facts by text match ordered by salience and recency.
func (s *SQLiteStore) Search(query string, limit int) ([]*Fact, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("nil sqlite store")
	}
	query = normalizeText(query)
	if limit <= 0 {
		limit = 10
	}
	if query == "" {
		return s.List(limit)
	}
	rows, err := s.db.Query(`
		SELECT id, claim, category, salience, valid_until, created_at, updated_at
		FROM mem0_facts
		WHERE claim LIKE ?
		ORDER BY salience DESC, updated_at DESC
		LIMIT ?
	`, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("search facts: %w", err)
	}
	defer rows.Close()
	return scanFacts(rows)
}

// CountExpired returns the number of facts expired at `now`.
func (s *SQLiteStore) CountExpired(now time.Time) (int, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("nil sqlite store")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var count int
	if err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM mem0_facts
		WHERE valid_until IS NOT NULL
		  AND valid_until < ?
	`, now.Format(time.RFC3339Nano)).Scan(&count); err != nil {
		return 0, fmt.Errorf("count expired facts: %w", err)
	}
	return count, nil
}

// DeleteAll removes all facts.
func (s *SQLiteStore) DeleteAll() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("nil sqlite store")
	}
	if _, err := s.db.Exec(`DELETE FROM mem0_facts`); err != nil {
		return fmt.Errorf("delete all mem0 facts: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ensureSchema() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("nil sqlite store")
	}
	schema := `
CREATE TABLE IF NOT EXISTS mem0_facts (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	fact_key    TEXT NOT NULL UNIQUE,
	claim       TEXT NOT NULL,
	category    TEXT NOT NULL,
	salience    REAL NOT NULL,
	valid_until TEXT,
	created_at  TEXT NOT NULL,
	updated_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_mem0_facts_salience_updated
	ON mem0_facts(salience DESC, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_mem0_facts_claim
	ON mem0_facts(claim);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("ensure mem0 schema: %w", err)
	}
	return nil
}

func (s *SQLiteStore) getByIDTx(tx *sql.Tx, id int64) (*Fact, error) {
	row := tx.QueryRow(`
		SELECT id, claim, category, salience, valid_until, created_at, updated_at
		FROM mem0_facts
		WHERE id = ?
	`, id)
	f, err := scanFactRow(row)
	if err != nil {
		return nil, fmt.Errorf("get fact by id: %w", err)
	}
	return f, nil
}

type factRowScanner interface {
	Scan(dest ...any) error
}

func scanFactRow(row factRowScanner) (*Fact, error) {
	var (
		f          Fact
		validUntil sql.NullString
		createdRaw string
		updatedRaw string
	)
	if err := row.Scan(
		&f.ID,
		&f.Claim,
		&f.Category,
		&f.Salience,
		&validUntil,
		&createdRaw,
		&updatedRaw,
	); err != nil {
		return nil, err
	}
	f.Salience = clampUnit(f.Salience)
	f.CreatedAt = parseTime(createdRaw)
	f.UpdatedAt = parseTime(updatedRaw)
	if validUntil.Valid {
		v := parseTime(validUntil.String)
		if !v.IsZero() {
			f.ValidUntil = &v
		}
	}
	return &f, nil
}

func scanFacts(rows *sql.Rows) ([]*Fact, error) {
	out := make([]*Fact, 0)
	for rows.Next() {
		f, err := scanFactRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
