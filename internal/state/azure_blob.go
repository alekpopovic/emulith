package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type AzureContainer struct {
	Account, Name, ETag     string
	LastModified, CreatedAt time.Time
	Metadata                map[string]string
}
type AzureBlob struct {
	Account, Container, Name, ETag, BodyPath, ContentType, ContentEncoding, ContentLanguage, CacheControl, ContentDisposition, ContentMD5 string
	Size                                                                                                                                  int64
	LastModified, CreatedAt                                                                                                               time.Time
	Metadata                                                                                                                              map[string]string
}

func metaJSON(m map[string]string) string {
	if m == nil {
		m = map[string]string{}
	}
	b, _ := json.Marshal(m)
	return string(b)
}
func parseMeta(s string) map[string]string {
	var m map[string]string
	if json.Unmarshal([]byte(s), &m) != nil || m == nil {
		m = map[string]string{}
	}
	return m
}
func (s *Store) CreateAzureContainer(ctx context.Context, c AzureContainer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `INSERT INTO azure_containers(account,name,etag,last_modified,metadata,created_at) VALUES(?,?,?,?,?,?)`, c.Account, c.Name, c.ETag, c.LastModified, metaJSON(c.Metadata), c.CreatedAt)
	if e != nil {
		return ErrConflict
	}
	return nil
}
func (s *Store) GetAzureContainer(ctx context.Context, a, n string) (AzureContainer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var c AzureContainer
	var m string
	e := s.db.QueryRowContext(ctx, `SELECT account,name,etag,last_modified,metadata,created_at FROM azure_containers WHERE account=? AND name=?`, a, n).Scan(&c.Account, &c.Name, &c.ETag, &c.LastModified, &m, &c.CreatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	c.Metadata = parseMeta(m)
	return c, e
}
func (s *Store) UpdateAzureContainer(ctx context.Context, c AzureContainer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `UPDATE azure_containers SET etag=?,last_modified=?,metadata=? WHERE account=? AND name=?`, c.ETag, c.LastModified, metaJSON(c.Metadata), c.Account, c.Name)
	return e
}
func (s *Store) DeleteAzureContainer(ctx context.Context, a, n string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, e := s.db.ExecContext(ctx, `DELETE FROM azure_containers WHERE account=? AND name=?`, a, n)
	return e
}
func (s *Store) ListAzureContainers(ctx context.Context, a, prefix string) ([]AzureContainer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, e := s.db.QueryContext(ctx, `SELECT account,name,etag,last_modified,metadata,created_at FROM azure_containers WHERE account=? AND name LIKE ? ORDER BY name`, a, prefix+"%")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []AzureContainer
	for rows.Next() {
		var c AzureContainer
		var m string
		if e = rows.Scan(&c.Account, &c.Name, &c.ETag, &c.LastModified, &m, &c.CreatedAt); e != nil {
			return nil, e
		}
		c.Metadata = parseMeta(m)
		out = append(out, c)
	}
	return out, rows.Err()
}
func (s *Store) PutAzureBlob(ctx context.Context, b AzureBlob) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var old string
	e := s.db.QueryRowContext(ctx, `SELECT body_path FROM azure_blobs WHERE account=? AND container=? AND name=?`, b.Account, b.Container, b.Name).Scan(&old)
	if errors.Is(e, sql.ErrNoRows) {
		old = ""
	} else if e != nil {
		return "", e
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO azure_blobs(account,container,name,etag,last_modified,body_path,size,content_type,content_encoding,content_language,cache_control,content_disposition,content_md5,metadata,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(account,container,name) DO UPDATE SET etag=excluded.etag,last_modified=excluded.last_modified,body_path=excluded.body_path,size=excluded.size,content_type=excluded.content_type,content_encoding=excluded.content_encoding,content_language=excluded.content_language,cache_control=excluded.cache_control,content_disposition=excluded.content_disposition,content_md5=excluded.content_md5,metadata=excluded.metadata`, b.Account, b.Container, b.Name, b.ETag, b.LastModified, b.BodyPath, b.Size, b.ContentType, b.ContentEncoding, b.ContentLanguage, b.CacheControl, b.ContentDisposition, b.ContentMD5, metaJSON(b.Metadata), b.CreatedAt)
	return old, e
}
func (s *Store) GetAzureBlob(ctx context.Context, a, c, n string) (AzureBlob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var b AzureBlob
	var m string
	e := s.db.QueryRowContext(ctx, `SELECT account,container,name,etag,last_modified,body_path,size,content_type,content_encoding,content_language,cache_control,content_disposition,content_md5,metadata,created_at FROM azure_blobs WHERE account=? AND container=? AND name=?`, a, c, n).Scan(&b.Account, &b.Container, &b.Name, &b.ETag, &b.LastModified, &b.BodyPath, &b.Size, &b.ContentType, &b.ContentEncoding, &b.ContentLanguage, &b.CacheControl, &b.ContentDisposition, &b.ContentMD5, &m, &b.CreatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return b, ErrNotFound
	}
	b.Metadata = parseMeta(m)
	return b, e
}
func (s *Store) DeleteAzureBlob(ctx context.Context, a, c, n string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var p string
	e := s.db.QueryRowContext(ctx, `SELECT body_path FROM azure_blobs WHERE account=? AND container=? AND name=?`, a, c, n).Scan(&p)
	if errors.Is(e, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if e != nil {
		return "", e
	}
	_, e = s.db.ExecContext(ctx, `DELETE FROM azure_blobs WHERE account=? AND container=? AND name=?`, a, c, n)
	return p, e
}
