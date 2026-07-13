package state

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type AzureQueue struct {
	Account, Name, ETag     string
	LastModified, CreatedAt time.Time
	Metadata                map[string]string
}

func (s *Store) CreateAzureQueue(ctx context.Context, q AzureQueue) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `INSERT INTO azure_queues(account,name,etag,last_modified,metadata,created_at) VALUES(?,?,?,?,?,?)`, q.Account, q.Name, q.ETag, q.LastModified, metaJSON(q.Metadata), q.CreatedAt)
	if e != nil {
		return ErrConflict
	}
	return nil
}
func (s *Store) GetAzureQueue(ctx context.Context, a, n string) (AzureQueue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var q AzureQueue
	var m string
	e := s.db.QueryRowContext(ctx, `SELECT account,name,etag,last_modified,metadata,created_at FROM azure_queues WHERE account=? AND name=?`, a, n).Scan(&q.Account, &q.Name, &q.ETag, &q.LastModified, &m, &q.CreatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return q, ErrNotFound
	}
	q.Metadata = parseMeta(m)
	return q, e
}
func (s *Store) UpdateAzureQueue(ctx context.Context, q AzureQueue) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `UPDATE azure_queues SET etag=?,last_modified=?,metadata=? WHERE account=? AND name=?`, q.ETag, q.LastModified, metaJSON(q.Metadata), q.Account, q.Name)
	return e
}
func (s *Store) DeleteAzureQueue(ctx context.Context, a, n string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, e := s.db.ExecContext(ctx, `DELETE FROM azure_queues WHERE account=? AND name=?`, a, n)
	if e == nil {
		if n, _ := r.RowsAffected(); n == 0 {
			return ErrNotFound
		}
	}
	return e
}
func (s *Store) ListAzureQueues(ctx context.Context, a, prefix string) ([]AzureQueue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, e := s.db.QueryContext(ctx, `SELECT account,name,etag,last_modified,metadata,created_at FROM azure_queues WHERE account=? AND name LIKE ? ORDER BY name`, a, prefix+"%")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []AzureQueue
	for rows.Next() {
		var q AzureQueue
		var m string
		if e = rows.Scan(&q.Account, &q.Name, &q.ETag, &q.LastModified, &m, &q.CreatedAt); e != nil {
			return nil, e
		}
		q.Metadata = parseMeta(m)
		out = append(out, q)
	}
	return out, rows.Err()
}
