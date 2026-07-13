package state

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type DynamoTable struct {
	Name, TableID, ARN, Status, BillingMode, PartitionKey, PartitionType, SortKey, SortType string
	CreatedAt                                                                               time.Time
}
type DynamoItem struct{ PrimaryKey, PartitionKey, SortKey, Payload []byte }
type DynamoWrite struct {
	Table                            string
	Key, Partition, SortKey, Payload []byte
	Delete                           bool
}

var ErrDynamoExists = errors.New("DynamoDB table exists")
var ErrDynamoNotFound = errors.New("DynamoDB table not found")

func (s *Store) CreateDynamoTable(ctx context.Context, t DynamoTable) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, `INSERT INTO dynamodb_tables(name,table_id,arn,status,created_at,billing_mode,partition_key,partition_type,sort_key,sort_type) VALUES(?,?,?,?,?,?,?,?,NULLIF(?,''),NULLIF(?,''))`, t.Name, t.TableID, t.ARN, t.Status, t.CreatedAt, t.BillingMode, t.PartitionKey, t.PartitionType, t.SortKey, t.SortType)
	if e != nil {
		if strings.Contains(strings.ToLower(e.Error()), "unique constraint") {
			return ErrDynamoExists
		}
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO dynamodb_attributes(table_name,name,attribute_type) VALUES(?,?,?)`, t.Name, t.PartitionKey, t.PartitionType)
	if e == nil && t.SortKey != "" {
		_, e = tx.ExecContext(ctx, `INSERT INTO dynamodb_attributes(table_name,name,attribute_type) VALUES(?,?,?)`, t.Name, t.SortKey, t.SortType)
	}
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (s *Store) GetDynamoTable(ctx context.Context, name string) (DynamoTable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return scanDynamo(s.db.QueryRowContext(ctx, `SELECT name,table_id,arn,status,created_at,billing_mode,partition_key,partition_type,COALESCE(sort_key,''),COALESCE(sort_type,'') FROM dynamodb_tables WHERE name=?`, name))
}

type rowScanner interface{ Scan(...any) error }

func scanDynamo(r rowScanner) (DynamoTable, error) {
	var t DynamoTable
	e := r.Scan(&t.Name, &t.TableID, &t.ARN, &t.Status, &t.CreatedAt, &t.BillingMode, &t.PartitionKey, &t.PartitionType, &t.SortKey, &t.SortType)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrDynamoNotFound
	}
	return t, e
}
func (s *Store) ListDynamoTables(ctx context.Context, start string, limit int) ([]DynamoTable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, e := s.db.QueryContext(ctx, `SELECT name,table_id,arn,status,created_at,billing_mode,partition_key,partition_type,COALESCE(sort_key,''),COALESCE(sort_type,'') FROM dynamodb_tables WHERE name>? ORDER BY name LIMIT ?`, start, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []DynamoTable
	for rows.Next() {
		t, e := scanDynamo(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) DeleteDynamoTable(ctx context.Context, name string) (DynamoTable, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return DynamoTable{}, e
	}
	defer tx.Rollback()
	t, e := scanDynamo(tx.QueryRowContext(ctx, `SELECT name,table_id,arn,status,created_at,billing_mode,partition_key,partition_type,COALESCE(sort_key,''),COALESCE(sort_type,'') FROM dynamodb_tables WHERE name=?`, name))
	if e != nil {
		return t, e
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM dynamodb_tables WHERE name=?`, name); e != nil {
		return t, e
	}
	return t, tx.Commit()
}

func (s *Store) GetDynamoItem(ctx context.Context, table string, key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var b []byte
	e := s.db.QueryRowContext(ctx, `SELECT payload FROM dynamodb_items WHERE table_name=? AND primary_key=?`, table, key).Scan(&b)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return b, e
}
func (s *Store) PutDynamoItem(ctx context.Context, table string, key, partition, sortKey, payload []byte) ([]byte, error) {
	return s.ConditionalPutDynamoItem(ctx, table, key, partition, sortKey, payload, nil)
}
func (s *Store) ConditionalPutDynamoItem(ctx context.Context, table string, key, partition, sortKey, payload []byte, check func([]byte) error) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var old []byte
	e = tx.QueryRowContext(ctx, `SELECT payload FROM dynamodb_items WHERE table_name=? AND primary_key=?`, table, key).Scan(&old)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return nil, e
	}
	if check != nil {
		if e = check(bytes.Clone(old)); e != nil {
			return nil, e
		}
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO dynamodb_items(table_name,primary_key,partition_key,sort_key,payload,item_size,updated_at) VALUES(?,?,?,?,?,?,?) ON CONFLICT(table_name,primary_key) DO UPDATE SET payload=excluded.payload,item_size=excluded.item_size,updated_at=excluded.updated_at`, table, key, partition, nullBytes(sortKey), payload, len(payload), time.Now().UTC())
	if e != nil {
		return nil, e
	}
	return old, tx.Commit()
}
func (s *Store) DeleteDynamoItem(ctx context.Context, table string, key []byte) ([]byte, error) {
	return s.ConditionalDeleteDynamoItem(ctx, table, key, nil)
}
func (s *Store) ConditionalDeleteDynamoItem(ctx context.Context, table string, key []byte, check func([]byte) error) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var old []byte
	e = tx.QueryRowContext(ctx, `SELECT payload FROM dynamodb_items WHERE table_name=? AND primary_key=?`, table, key).Scan(&old)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return nil, e
	}
	if check != nil {
		if e = check(bytes.Clone(old)); e != nil {
			return nil, e
		}
	}
	if errors.Is(e, sql.ErrNoRows) || old == nil {
		return nil, nil
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM dynamodb_items WHERE table_name=? AND primary_key=?`, table, key); e != nil {
		return nil, e
	}
	return old, tx.Commit()
}
func (s *Store) UpdateDynamoItem(ctx context.Context, table string, key, partition, sortKey []byte, mutate func([]byte) ([]byte, error)) ([]byte, []byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return nil, nil, e
	}
	defer tx.Rollback()
	var old []byte
	e = tx.QueryRowContext(ctx, `SELECT payload FROM dynamodb_items WHERE table_name=? AND primary_key=?`, table, key).Scan(&old)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return nil, nil, e
	}
	next, e := mutate(bytes.Clone(old))
	if e != nil {
		return nil, nil, e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO dynamodb_items(table_name,primary_key,partition_key,sort_key,payload,item_size,updated_at) VALUES(?,?,?,?,?,?,?) ON CONFLICT(table_name,primary_key) DO UPDATE SET payload=excluded.payload,item_size=excluded.item_size,updated_at=excluded.updated_at`, table, key, partition, nullBytes(sortKey), next, len(next), time.Now().UTC())
	if e != nil {
		return nil, nil, e
	}
	return old, next, tx.Commit()
}
func nullBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
func (s *Store) ListAllDynamoItems(ctx context.Context, table string, maximum int) ([]DynamoItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, e := s.db.QueryContext(ctx, `SELECT primary_key,partition_key,COALESCE(sort_key,X''),payload FROM dynamodb_items WHERE table_name=? ORDER BY primary_key LIMIT ?`, table, maximum+1)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []DynamoItem
	for rows.Next() {
		var x DynamoItem
		if e = rows.Scan(&x.PrimaryKey, &x.PartitionKey, &x.SortKey, &x.Payload); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	if len(out) > maximum {
		return nil, errors.New("DynamoDB local evaluation limit exceeded")
	}
	return out, nil
}
func (s *Store) BatchWriteDynamoItems(ctx context.Context, ops []DynamoWrite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	checked := map[string]bool{}
	for _, op := range ops {
		if !checked[op.Table] {
			var one int
			if e = tx.QueryRowContext(ctx, `SELECT 1 FROM dynamodb_tables WHERE name=?`, op.Table).Scan(&one); errors.Is(e, sql.ErrNoRows) {
				return ErrDynamoNotFound
			} else if e != nil {
				return e
			}
			checked[op.Table] = true
		}
		if op.Delete {
			if _, e = tx.ExecContext(ctx, `DELETE FROM dynamodb_items WHERE table_name=? AND primary_key=?`, op.Table, op.Key); e != nil {
				return e
			}
		} else {
			if _, e = tx.ExecContext(ctx, `INSERT INTO dynamodb_items(table_name,primary_key,partition_key,sort_key,payload,item_size,updated_at) VALUES(?,?,?,?,?,?,?) ON CONFLICT(table_name,primary_key) DO UPDATE SET payload=excluded.payload,item_size=excluded.item_size,updated_at=excluded.updated_at`, op.Table, op.Key, op.Partition, nullBytes(op.SortKey), op.Payload, len(op.Payload), time.Now().UTC()); e != nil {
				return e
			}
		}
	}
	return tx.Commit()
}
