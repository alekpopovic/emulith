package state

import (
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
