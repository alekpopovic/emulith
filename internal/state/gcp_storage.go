package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type GCPBucket struct {
	Project, Name, Location, StorageClass, ETag string
	Labels                                      map[string]string
	CreatedAt, UpdatedAt                        time.Time
	Metageneration                              int64
}

func (s *Store) CreateGCPBucket(ctx context.Context, b GCPBucket) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, _ := json.Marshal(b.Labels)
	_, e := s.db.ExecContext(ctx, `INSERT INTO gcp_buckets(project,name,location,storage_class,labels,etag,created_at,updated_at,metageneration) VALUES(?,?,?,?,?,?,?,?,?)`, b.Project, b.Name, b.Location, b.StorageClass, string(v), b.ETag, b.CreatedAt, b.UpdatedAt, b.Metageneration)
	if e != nil {
		return ErrConflict
	}
	return nil
}
func (s *Store) GetGCPBucket(ctx context.Context, p, n string) (GCPBucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var b GCPBucket
	var l string
	e := s.db.QueryRowContext(ctx, `SELECT project,name,location,storage_class,labels,etag,created_at,updated_at,metageneration FROM gcp_buckets WHERE project=? AND name=?`, p, n).Scan(&b.Project, &b.Name, &b.Location, &b.StorageClass, &l, &b.ETag, &b.CreatedAt, &b.UpdatedAt, &b.Metageneration)
	if errors.Is(e, sql.ErrNoRows) {
		return b, ErrNotFound
	}
	json.Unmarshal([]byte(l), &b.Labels)
	return b, e
}
func (s *Store) ListGCPBuckets(ctx context.Context, p string) ([]GCPBucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, e := s.db.QueryContext(ctx, `SELECT project,name,location,storage_class,labels,etag,created_at,updated_at,metageneration FROM gcp_buckets WHERE project=? ORDER BY name`, p)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []GCPBucket
	for rows.Next() {
		var b GCPBucket
		var l string
		if e = rows.Scan(&b.Project, &b.Name, &b.Location, &b.StorageClass, &l, &b.ETag, &b.CreatedAt, &b.UpdatedAt, &b.Metageneration); e != nil {
			return nil, e
		}
		json.Unmarshal([]byte(l), &b.Labels)
		out = append(out, b)
	}
	return out, rows.Err()
}
func (s *Store) UpdateGCPBucket(ctx context.Context, b GCPBucket) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, _ := json.Marshal(b.Labels)
	_, e := s.db.ExecContext(ctx, `UPDATE gcp_buckets SET labels=?,etag=?,updated_at=?,metageneration=? WHERE project=? AND name=?`, string(v), b.ETag, b.UpdatedAt, b.Metageneration, b.Project, b.Name)
	return e
}
func (s *Store) DeleteGCPBucket(ctx context.Context, p, n string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, e := s.db.ExecContext(ctx, `DELETE FROM gcp_buckets WHERE project=? AND name=?`, p, n)
	if e != nil {
		return e
	}
	x, _ := r.RowsAffected()
	if x == 0 {
		return ErrNotFound
	}
	return nil
}
