package state

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("state record not found")
var ErrConflict = errors.New("state record already exists")

type S3Bucket struct {
	Name, Region string
	CreatedAt    time.Time
}
type S3Object struct {
	Bucket, Key, ETag, ContentType, BodyPath string
	Size                                     int64
	LastModified                             time.Time
}

func (s *Store) CreateS3Bucket(ctx context.Context, bucket S3Bucket) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := s.db.ExecContext(ctx, `INSERT INTO s3_buckets(name,region,created_at) VALUES(?,?,?)`, bucket.Name, bucket.Region, bucket.CreatedAt)
	if err != nil {
		return ErrConflict
	}
	return nil
}
func (s *Store) ListS3Buckets(ctx context.Context) ([]S3Bucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.QueryContext(ctx, `SELECT name,region,created_at FROM s3_buckets ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []S3Bucket
	for rows.Next() {
		var b S3Bucket
		if err := rows.Scan(&b.Name, &b.Region, &b.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}
func (s *Store) S3BucketExists(ctx context.Context, name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM s3_buckets WHERE name=?`, name).Scan(&n)
	return n == 1, err
}
func (s *Store) PutS3Object(ctx context.Context, object S3Object) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var old string
	err := s.db.QueryRowContext(ctx, `SELECT body_path FROM s3_objects WHERE bucket=? AND key=?`, object.Bucket, object.Key).Scan(&old)
	if errors.Is(err, sql.ErrNoRows) {
		old = ""
	} else if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO s3_objects(bucket,key,etag,size,content_type,last_modified,body_path) VALUES(?,?,?,?,?,?,?) ON CONFLICT(bucket,key) DO UPDATE SET etag=excluded.etag,size=excluded.size,content_type=excluded.content_type,last_modified=excluded.last_modified,body_path=excluded.body_path`, object.Bucket, object.Key, object.ETag, object.Size, object.ContentType, object.LastModified, object.BodyPath)
	return old, err
}
func (s *Store) GetS3Object(ctx context.Context, bucket, key string) (S3Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var o S3Object
	err := s.db.QueryRowContext(ctx, `SELECT bucket,key,etag,size,COALESCE(content_type,''),last_modified,body_path FROM s3_objects WHERE bucket=? AND key=?`, bucket, key).Scan(&o.Bucket, &o.Key, &o.ETag, &o.Size, &o.ContentType, &o.LastModified, &o.BodyPath)
	if errors.Is(err, sql.ErrNoRows) {
		return o, ErrNotFound
	}
	return o, err
}
func (s *Store) DeleteS3Object(ctx context.Context, bucket, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var path string
	err := s.db.QueryRowContext(ctx, `SELECT body_path FROM s3_objects WHERE bucket=? AND key=?`, bucket, key).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM s3_objects WHERE bucket=? AND key=?`, bucket, key)
	return path, err
}
func (s *Store) ListS3Objects(ctx context.Context, bucket, prefix string, limit int) ([]S3Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.QueryContext(ctx, `SELECT bucket,key,etag,size,COALESCE(content_type,''),last_modified,body_path FROM s3_objects WHERE bucket=? AND key LIKE ? ESCAPE '\' ORDER BY key LIMIT ?`, bucket, escapeLike(prefix)+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []S3Object
	for rows.Next() {
		var o S3Object
		if err := rows.Scan(&o.Bucket, &o.Key, &o.ETag, &o.Size, &o.ContentType, &o.LastModified, &o.BodyPath); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
func escapeLike(s string) string {
	r := ""
	for _, c := range s {
		if c == '%' || c == '_' || c == '\\' {
			r += "\\"
		}
		r += string(c)
	}
	return r
}
