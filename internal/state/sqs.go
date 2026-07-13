package state

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type SQSQueue struct {
	Name, URLPath     string
	VisibilityTimeout int
	CreatedAt         time.Time
}
type SQSMessage struct {
	ID, QueueName, Body, MD5, ReceiptHandle string
	VisibleAt, CreatedAt                    time.Time
}

func (s *Store) CreateSQSQueue(ctx context.Context, q SQSQueue) (SQSQueue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var existing SQSQueue
	err := s.db.QueryRowContext(ctx, `SELECT name,url_path,visibility_timeout_seconds,created_at FROM sqs_queues WHERE name=?`, q.Name).Scan(&existing.Name, &existing.URLPath, &existing.VisibilityTimeout, &existing.CreatedAt)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return q, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO sqs_queues(name,url_path,visibility_timeout_seconds,created_at) VALUES(?,?,?,?)`, q.Name, q.URLPath, q.VisibilityTimeout, q.CreatedAt)
	return q, err
}
func (s *Store) GetSQSQueue(ctx context.Context, name string) (SQSQueue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var q SQSQueue
	err := s.db.QueryRowContext(ctx, `SELECT name,url_path,visibility_timeout_seconds,created_at FROM sqs_queues WHERE name=?`, name).Scan(&q.Name, &q.URLPath, &q.VisibilityTimeout, &q.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return q, ErrNotFound
	}
	return q, err
}
func (s *Store) ListSQSQueues(ctx context.Context, prefix string) ([]SQSQueue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.QueryContext(ctx, `SELECT name,url_path,visibility_timeout_seconds,created_at FROM sqs_queues WHERE name LIKE ? ESCAPE '\' ORDER BY name`, escapeLike(prefix)+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SQSQueue
	for rows.Next() {
		var q SQSQueue
		if err := rows.Scan(&q.Name, &q.URLPath, &q.VisibilityTimeout, &q.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}
func (s *Store) SendSQSMessage(ctx context.Context, queue, body string, now time.Time) (SQSMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sum := md5.Sum([]byte(body))
	m := SQSMessage{ID: randomToken(), QueueName: queue, Body: body, MD5: hex.EncodeToString(sum[:]), VisibleAt: now, CreatedAt: now}
	_, err := s.db.ExecContext(ctx, `INSERT INTO sqs_messages(id,queue_name,body,md5,visible_at,created_at) VALUES(?,?,?,?,?,?)`, m.ID, m.QueueName, m.Body, m.MD5, m.VisibleAt, m.CreatedAt)
	return m, err
}
func (s *Store) ReceiveSQSMessages(ctx context.Context, queue string, max, visibility int, now time.Time) ([]SQSMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id,queue_name,body,md5,COALESCE(receipt_handle,''),visible_at,created_at FROM sqs_messages WHERE queue_name=? AND visible_at<=? ORDER BY created_at,id LIMIT ?`, queue, now, max)
	if err != nil {
		return nil, err
	}
	var out []SQSMessage
	for rows.Next() {
		var m SQSMessage
		if err := rows.Scan(&m.ID, &m.QueueName, &m.Body, &m.MD5, &m.ReceiptHandle, &m.VisibleAt, &m.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range out {
		out[i].ReceiptHandle = randomToken()
		out[i].VisibleAt = now.Add(time.Duration(visibility) * time.Second)
		if _, err := tx.ExecContext(ctx, `UPDATE sqs_messages SET receipt_handle=?,visible_at=? WHERE id=?`, out[i].ReceiptHandle, out[i].VisibleAt, out[i].ID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *Store) DeleteSQSMessage(ctx context.Context, queue, receipt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := s.db.ExecContext(ctx, `DELETE FROM sqs_messages WHERE queue_name=? AND receipt_handle=?`, queue, receipt)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *Store) PurgeSQSQueue(ctx context.Context, queue string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.ExecContext(ctx, `DELETE FROM sqs_messages WHERE queue_name=?`, queue)
	return err
}
func (s *Store) SQSMessageCounts(ctx context.Context, queue string, now time.Time) (visible, hidden int, error error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN visible_at<=? THEN 1 ELSE 0 END),0),COALESCE(SUM(CASE WHEN visible_at>? THEN 1 ELSE 0 END),0) FROM sqs_messages WHERE queue_name=?`, now, now, queue).Scan(&visible, &hidden)
	return visible, hidden, err
}
func randomToken() string {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b[:])
}
