package state

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type LogEvent struct {
	Timestamp int64
	Message   string
	Ingested  int64
	ID        int64
}

func (s *Store) LogCreateGroup(ctx context.Context, n string) error {
	_, e := s.db.ExecContext(ctx, "INSERT INTO cw_log_groups(name,created_at) VALUES(?,?)", n, time.Now().UnixMilli())
	return e
}
func (s *Store) LogDeleteGroup(ctx context.Context, n string) error {
	r, e := s.db.ExecContext(ctx, "DELETE FROM cw_log_groups WHERE name=?", n)
	if e == nil {
		c, _ := r.RowsAffected()
		if c == 0 {
			return sql.ErrNoRows
		}
	}
	return e
}
func (s *Store) LogGroups(ctx context.Context, prefix string, limit int) ([]string, error) {
	q := "SELECT name FROM cw_log_groups WHERE name LIKE ? ORDER BY name LIMIT ?"
	rows, e := s.db.QueryContext(ctx, q, prefix+"%", limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var a []string
	for rows.Next() {
		var n string
		_ = rows.Scan(&n)
		a = append(a, n)
	}
	return a, rows.Err()
}
func (s *Store) LogCreateStream(ctx context.Context, g, n string) error {
	_, e := s.db.ExecContext(ctx, "INSERT INTO cw_log_streams(group_name,name,created_at) VALUES(?,?,?)", g, n, time.Now().UnixMilli())
	return e
}
func (s *Store) LogStreams(ctx context.Context, g string) ([]string, error) {
	rows, e := s.db.QueryContext(ctx, "SELECT name FROM cw_log_streams WHERE group_name=? ORDER BY name", g)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var a []string
	for rows.Next() {
		var n string
		_ = rows.Scan(&n)
		a = append(a, n)
	}
	return a, rows.Err()
}
func (s *Store) LogPut(ctx context.Context, g, st string, ev []LogEvent) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, v := range ev {
		if _, e = tx.ExecContext(ctx, "INSERT INTO cw_log_events(group_name,stream_name,timestamp_ms,message,ingested_ms) VALUES(?,?,?,?,?)", g, st, v.Timestamp, v.Message, v.Ingested); e != nil {
			return e
		}
	}
	return tx.Commit()
}
func (s *Store) LogEvents(ctx context.Context, g, st string, start, end int64, limit int) ([]LogEvent, error) {
	rows, e := s.db.QueryContext(ctx, "SELECT timestamp_ms,message,ingested_ms,id FROM cw_log_events WHERE group_name=? AND stream_name=? AND timestamp_ms>=? AND timestamp_ms<? ORDER BY timestamp_ms, id LIMIT ?", g, st, start, end, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var a []LogEvent
	for rows.Next() {
		var v LogEvent
		if e = rows.Scan(&v.Timestamp, &v.Message, &v.Ingested, &v.ID); e != nil {
			return nil, e
		}
		a = append(a, v)
	}
	return a, rows.Err()
}

var _ = errors.New
