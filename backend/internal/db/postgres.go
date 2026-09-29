// Package db provides PostgreSQL persistence for RecoLab.
// It stores every feedback event submitted via POST /api/events so that
// events survive backend restarts and can be queried for analytics.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Store wraps a *sql.DB and exposes domain-level operations.
type Store struct {
	db *sql.DB
}

// FeedbackEvent mirrors the payload stored in Postgres.
type FeedbackEvent struct {
	ID           int64
	FeedbackType string
	UserID       string
	ItemID       string
	Timestamp    time.Time
	TenantID     string
	CreatedAt    time.Time
}

// Open opens a Postgres connection using dsn and runs the schema migration.
func Open(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	log.Println("PostgreSQL connected and schema up-to-date")
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS feedback_events (
			id            BIGSERIAL PRIMARY KEY,
			feedback_type TEXT        NOT NULL,
			user_id       TEXT        NOT NULL,
			item_id       TEXT        NOT NULL,
			timestamp     TIMESTAMPTZ NOT NULL,
			tenant_id     TEXT        NOT NULL DEFAULT '',
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_fe_user   ON feedback_events (user_id);
		CREATE INDEX IF NOT EXISTS idx_fe_item   ON feedback_events (item_id);
		CREATE INDEX IF NOT EXISTS idx_fe_tenant ON feedback_events (tenant_id);
	`)
	return err
}

// InsertEvent persists a single feedback event.
func (s *Store) InsertEvent(feedbackType, userID, itemID, tenantID string, ts time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO feedback_events (feedback_type, user_id, item_id, timestamp, tenant_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		feedbackType, userID, itemID, ts, tenantID,
	)
	return err
}

// RecentEvents returns the most recent n events for a tenant (latest first).
func (s *Store) RecentEvents(tenantID string, n int) ([]FeedbackEvent, error) {
	rows, err := s.db.Query(
		`SELECT id, feedback_type, user_id, item_id, timestamp, tenant_id, created_at
		   FROM feedback_events
		  WHERE tenant_id = $1
		  ORDER BY created_at DESC
		  LIMIT $2`,
		tenantID, n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []FeedbackEvent
	for rows.Next() {
		var e FeedbackEvent
		if err := rows.Scan(&e.ID, &e.FeedbackType, &e.UserID, &e.ItemID,
			&e.Timestamp, &e.TenantID, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// FeedbackCount returns the total number of events stored for a tenant.
func (s *Store) FeedbackCount(tenantID string) (int64, error) {
	var count int64
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM feedback_events WHERE tenant_id = $1`, tenantID,
	).Scan(&count)
	return count, err
}

// Close closes the underlying database connection pool.
func (s *Store) Close() error {
	return s.db.Close()
}
