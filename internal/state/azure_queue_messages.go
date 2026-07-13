package state

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"time"
)

type AzureQueueMessage struct {
	Account, Queue, ID, Body, PopReceipt        string
	InsertedAt, ExpiresAt, VisibleAt, UpdatedAt time.Time
	DequeueCount                                int
}

func (s *Store) PutAzureMessage(ctx context.Context, a, q, body string, ttl, delay time.Duration) (AzureQueueMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	id := uuid.NewString()
	exp := now.Add(ttl)
	if ttl <= 0 {
		exp = now.Add(7 * 24 * time.Hour)
	}
	if ttl < 0 {
		exp = time.Unix(0, 0).UTC()
	}
	m := AzureQueueMessage{Account: a, Queue: q, ID: id, Body: body, InsertedAt: now, ExpiresAt: exp, VisibleAt: now.Add(delay), DequeueCount: 0, UpdatedAt: now}
	m.PopReceipt = uuid.NewString()
	_, e := s.db.ExecContext(ctx, `INSERT INTO azure_queue_messages(account,queue,id,body,inserted_at,expires_at,visible_at,dequeue_count,pop_receipt,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, m.Account, m.Queue, m.ID, m.Body, m.InsertedAt, m.ExpiresAt, m.VisibleAt, m.DequeueCount, m.PopReceipt, m.UpdatedAt)
	return m, e
}
func (s *Store) QueueMessages(ctx context.Context, a, q string, peek bool, n int, visibility time.Duration) ([]AzureQueueMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	_, _ = s.db.ExecContext(ctx, `DELETE FROM azure_queue_messages WHERE account=? AND queue=? AND expires_at>0 AND expires_at<=?`, a, q, now)
	rows, e := s.db.QueryContext(ctx, `SELECT account,queue,id,body,inserted_at,expires_at,visible_at,dequeue_count,pop_receipt,updated_at FROM azure_queue_messages WHERE account=? AND queue=? AND visible_at<=? ORDER BY inserted_at,id LIMIT ?`, a, q, now, n)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []AzureQueueMessage
	for rows.Next() {
		var m AzureQueueMessage
		if e = rows.Scan(&m.Account, &m.Queue, &m.ID, &m.Body, &m.InsertedAt, &m.ExpiresAt, &m.VisibleAt, &m.DequeueCount, &m.PopReceipt, &m.UpdatedAt); e != nil {
			return nil, e
		}
		if !peek {
			m.DequeueCount++
			m.VisibleAt = now.Add(visibility)
			m.PopReceipt = uuid.NewString()
			m.UpdatedAt = now
			if _, e = s.db.ExecContext(ctx, `UPDATE azure_queue_messages SET visible_at=?,dequeue_count=?,pop_receipt=?,updated_at=? WHERE account=? AND queue=? AND id=?`, m.VisibleAt, m.DequeueCount, m.PopReceipt, m.UpdatedAt, a, q, m.ID); e != nil {
				return nil, e
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *Store) UpdateAzureMessage(ctx context.Context, a, q, id, receipt, body string, visibility time.Duration) (AzureQueueMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var m AzureQueueMessage
	e := s.db.QueryRowContext(ctx, `SELECT account,queue,id,body,inserted_at,expires_at,visible_at,dequeue_count,pop_receipt,updated_at FROM azure_queue_messages WHERE account=? AND queue=? AND id=? AND pop_receipt=?`, a, q, id, receipt).Scan(&m.Account, &m.Queue, &m.ID, &m.Body, &m.InsertedAt, &m.ExpiresAt, &m.VisibleAt, &m.DequeueCount, &m.PopReceipt, &m.UpdatedAt)
	if e != nil {
		return m, e
	}
	if body != "" {
		m.Body = body
	}
	m.VisibleAt = time.Now().UTC().Add(visibility)
	m.PopReceipt = uuid.NewString()
	m.UpdatedAt = time.Now().UTC()
	_, e = s.db.ExecContext(ctx, `UPDATE azure_queue_messages SET body=?,visible_at=?,pop_receipt=?,updated_at=? WHERE account=? AND queue=? AND id=?`, m.Body, m.VisibleAt, m.PopReceipt, m.UpdatedAt, a, q, id)
	return m, e
}
func (s *Store) DeleteAzureMessage(ctx context.Context, a, q, id, receipt string) error {
	r, e := s.db.ExecContext(ctx, `DELETE FROM azure_queue_messages WHERE account=? AND queue=? AND id=? AND pop_receipt=?`, a, q, id, receipt)
	if e == nil {
		n, _ := r.RowsAffected()
		if n == 0 {
			return sql.ErrNoRows
		}
	}
	return e
}
func (s *Store) ClearAzureMessages(ctx context.Context, a, q string) error {
	_, e := s.db.ExecContext(ctx, `DELETE FROM azure_queue_messages WHERE account=? AND queue=?`, a, q)
	return e
}

var _ = errors.New
