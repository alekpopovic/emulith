package state

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type GCPObject struct {
	Project, Bucket, Name, BodyPath, ContentType, ETag string
	Generation, Metageneration, Size                   int64
	CreatedAt, UpdatedAt                               time.Time
}

func (s *Store) PutGCPObject(ctx context.Context, o GCPObject) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `INSERT INTO gcp_objects(project,bucket,name,body_path,content_type,etag,generation,metageneration,size,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(project,bucket,name) DO UPDATE SET body_path=excluded.body_path,content_type=excluded.content_type,etag=excluded.etag,generation=excluded.generation,metageneration=excluded.metageneration,size=excluded.size,updated_at=excluded.updated_at`, o.Project, o.Bucket, o.Name, o.BodyPath, o.ContentType, o.ETag, o.Generation, o.Metageneration, o.Size, o.CreatedAt, o.UpdatedAt)
	return e
}
func (s *Store) GetGCPObject(ctx context.Context, p, b, n string) (GCPObject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var o GCPObject
	e := s.db.QueryRowContext(ctx, `SELECT project,bucket,name,body_path,content_type,etag,generation,metageneration,size,created_at,updated_at FROM gcp_objects WHERE project=? AND bucket=? AND name=?`, p, b, n).Scan(&o.Project, &o.Bucket, &o.Name, &o.BodyPath, &o.ContentType, &o.ETag, &o.Generation, &o.Metageneration, &o.Size, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return o, ErrNotFound
	}
	return o, e
}
func (s *Store) DeleteGCPObject(ctx context.Context, p, b, n string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `DELETE FROM gcp_objects WHERE project=? AND bucket=? AND name=?`, p, b, n)
	return e
}
