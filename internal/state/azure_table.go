package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type AzureTable struct {
	Account, Name string
	CreatedAt     time.Time
}
type AzureEntity struct {
	Account, Table, PartitionKey, RowKey, ETag string
	Timestamp                                  time.Time
	Properties                                 map[string]json.RawMessage
}

func (s *Store) CreateAzureTable(ctx context.Context, t AzureTable) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `INSERT INTO azure_tables(account,name,created_at) VALUES(?,?,?)`, t.Account, t.Name, t.CreatedAt)
	if e != nil {
		return ErrConflict
	}
	return nil
}
func (s *Store) DeleteAzureTable(ctx context.Context, a, n string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, e := s.db.ExecContext(ctx, `DELETE FROM azure_tables WHERE account=? AND name=?`, a, n)
	if e != nil {
		return e
	}
	x, _ := r.RowsAffected()
	if x == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *Store) GetAzureTable(ctx context.Context, a, n string) (AzureTable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var t AzureTable
	e := s.db.QueryRowContext(ctx, `SELECT account,name,created_at FROM azure_tables WHERE account=? AND name=?`, a, n).Scan(&t.Account, &t.Name, &t.CreatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, e
}
func (s *Store) ListAzureTables(ctx context.Context, a string) ([]AzureTable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, e := s.db.QueryContext(ctx, `SELECT account,name,created_at FROM azure_tables WHERE account=? ORDER BY name`, a)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []AzureTable
	for rows.Next() {
		var t AzureTable
		if e = rows.Scan(&t.Account, &t.Name, &t.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) GetAzureEntity(ctx context.Context, a, t, p, r string) (AzureEntity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var x AzureEntity
	var raw string
	e := s.db.QueryRowContext(ctx, `SELECT account,table_name,partition_key,row_key,etag,timestamp,properties FROM azure_entities WHERE account=? AND table_name=? AND partition_key=? AND row_key=?`, a, t, p, r).Scan(&x.Account, &x.Table, &x.PartitionKey, &x.RowKey, &x.ETag, &x.Timestamp, &raw)
	if errors.Is(e, sql.ErrNoRows) {
		return x, ErrNotFound
	}
	if e == nil {
		e = json.Unmarshal([]byte(raw), &x.Properties)
	}
	return x, e
}
func (s *Store) SaveAzureEntity(ctx context.Context, x AzureEntity) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, e := json.Marshal(x.Properties)
	if e != nil {
		return e
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO azure_entities(account,table_name,partition_key,row_key,etag,timestamp,properties) VALUES(?,?,?,?,?,?,?) ON CONFLICT(account,table_name,partition_key,row_key) DO UPDATE SET etag=excluded.etag,timestamp=excluded.timestamp,properties=excluded.properties`, x.Account, x.Table, x.PartitionKey, x.RowKey, x.ETag, x.Timestamp, string(b))
	return e
}
func (s *Store) DeleteAzureEntity(ctx context.Context, a, t, p, r string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	z, e := s.db.ExecContext(ctx, `DELETE FROM azure_entities WHERE account=? AND table_name=? AND partition_key=? AND row_key=?`, a, t, p, r)
	if e != nil {
		return e
	}
	n, _ := z.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
