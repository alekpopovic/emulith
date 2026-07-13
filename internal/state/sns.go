package state

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SNSTopic struct {
	Name, ARN, DisplayName string
	CreatedAt              time.Time
}

var ErrSNSTopicNotFound = errors.New("SNS topic not found")

func (s *Store) CreateSNSTopic(ctx context.Context, t SNSTopic) (SNSTopic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return t, e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, `INSERT OR IGNORE INTO sns_topics(name,arn,display_name,created_at) VALUES(?,?,?,?)`, t.Name, t.ARN, t.DisplayName, t.CreatedAt)
	if e != nil {
		return t, e
	}
	e = tx.QueryRowContext(ctx, `SELECT name,arn,display_name,created_at FROM sns_topics WHERE name=?`, t.Name).Scan(&t.Name, &t.ARN, &t.DisplayName, &t.CreatedAt)
	if e != nil {
		return t, e
	}
	return t, tx.Commit()
}
func (s *Store) GetSNSTopic(ctx context.Context, arn string) (SNSTopic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var t SNSTopic
	e := s.db.QueryRowContext(ctx, `SELECT name,arn,display_name,created_at FROM sns_topics WHERE arn=?`, arn).Scan(&t.Name, &t.ARN, &t.DisplayName, &t.CreatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrSNSTopicNotFound
	}
	return t, e
}
func (s *Store) ListSNSTopics(ctx context.Context, start string, limit int) ([]SNSTopic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, e := s.db.QueryContext(ctx, `SELECT name,arn,display_name,created_at FROM sns_topics WHERE arn>? ORDER BY arn LIMIT ?`, start, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []SNSTopic
	for rows.Next() {
		var t SNSTopic
		if e = rows.Scan(&t.Name, &t.ARN, &t.DisplayName, &t.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) DeleteSNSTopic(ctx context.Context, arn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, e := s.db.ExecContext(ctx, `DELETE FROM sns_topics WHERE arn=?`, arn)
	return e
}
