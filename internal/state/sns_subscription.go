package state

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SNSSubscription struct {
	ID, TopicARN, Protocol, Endpoint string
	RawDelivery                      bool
	CreatedAt                        time.Time
}

var ErrSNSSubscriptionNotFound = errors.New("SNS subscription not found")

func (s *Store) CreateSNSSubscription(ctx context.Context, v SNSSubscription) (SNSSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, e := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO sns_subscriptions(id,topic_arn,protocol,endpoint,raw_delivery,created_at) VALUES(?,?,?,?,?,?)`, v.ID, v.TopicARN, v.Protocol, v.Endpoint, v.RawDelivery, v.CreatedAt)
	if e != nil {
		return v, e
	}
	return s.getSub(ctx, v.ID)
}
func (s *Store) getSub(ctx context.Context, id string) (SNSSubscription, error) {
	var v SNSSubscription
	var raw int
	e := s.db.QueryRowContext(ctx, `SELECT id,topic_arn,protocol,endpoint,raw_delivery,created_at FROM sns_subscriptions WHERE id=?`, id).Scan(&v.ID, &v.TopicARN, &v.Protocol, &v.Endpoint, &raw, &v.CreatedAt)
	v.RawDelivery = raw != 0
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrSNSSubscriptionNotFound
	}
	return v, e
}
func (s *Store) GetSNSSubscription(ctx context.Context, id string) (SNSSubscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getSub(ctx, id)
}
func (s *Store) ListSNSSubscriptions(ctx context.Context, topic string) ([]SNSSubscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q := `SELECT id,topic_arn,protocol,endpoint,raw_delivery,created_at FROM sns_subscriptions`
	args := []any{}
	if topic != "" {
		q += ` WHERE topic_arn=?`
		args = append(args, topic)
	}
	q += ` ORDER BY id`
	rows, e := s.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []SNSSubscription
	for rows.Next() {
		var v SNSSubscription
		var raw int
		if e = rows.Scan(&v.ID, &v.TopicARN, &v.Protocol, &v.Endpoint, &raw, &v.CreatedAt); e != nil {
			return nil, e
		}
		v.RawDelivery = raw != 0
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) DeleteSNSSubscription(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, e := s.db.ExecContext(ctx, `DELETE FROM sns_subscriptions WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrSNSSubscriptionNotFound
	}
	return nil
}
func (s *Store) SetSNSRawDelivery(ctx context.Context, id string, raw bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, e := s.db.ExecContext(ctx, `UPDATE sns_subscriptions SET raw_delivery=? WHERE id=?`, raw, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrSNSSubscriptionNotFound
	}
	return nil
}
