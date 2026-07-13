package state

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

//go:embed migrations.sql
var migrations string

type Store struct {
	dataDir     string
	objectsRoot string
	tmpRoot     string
	db          *sql.DB
	mu          sync.RWMutex
}

func Open(ctx context.Context, dataDir string) (*Store, error) {
	root, err := safeRoot(dataDir)
	if err != nil {
		return nil, err
	}
	objects := filepath.Join(root, "objects")
	tmp := filepath.Join(root, "tmp")
	for _, dir := range []string{root, filepath.Join(objects, "aws", "s3"), tmp} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create state directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite", filepath.Join(root, "emulith.db"))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &Store{dataDir: root, objectsRoot: objects, tmpRoot: tmp, db: db}
	if err := store.initialize(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func safeRoot(dataDir string) (string, error) {
	if dataDir == "" {
		return "", errors.New("data directory must not be empty")
	}
	root, err := filepath.Abs(filepath.Clean(dataDir))
	if err != nil {
		return "", fmt.Errorf("resolve data directory: %w", err)
	}
	home, _ := os.UserHomeDir()
	volume := filepath.VolumeName(root) + string(filepath.Separator)
	if root == volume || (home != "" && root == filepath.Clean(home)) {
		return "", fmt.Errorf("unsafe data directory %q", root)
	}
	return root, nil
}

func (s *Store) initialize(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, "PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000; PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("configure sqlite: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, migrations); err != nil {
		return fmt.Errorf("migrate sqlite: %w", err)
	}
	return nil
}

func (s *Store) Close() error        { return s.db.Close() }
func (s *Store) DataDir() string     { return s.dataDir }
func (s *Store) ObjectsRoot() string { return s.objectsRoot }

func (s *Store) NewObjectBodyPath(provider, service, namespace, key string) (string, error) {
	if provider == "" || service == "" || namespace == "" {
		return "", errors.New("provider, service, and namespace are required")
	}
	digest := sha256.Sum256([]byte(provider + "\x00" + service + "\x00" + namespace + "\x00" + key))
	hexDigest := hex.EncodeToString(digest[:])
	path := filepath.Join(s.objectsRoot, hexDigest[:2], hexDigest[2:4], hexDigest)
	if !within(s.objectsRoot, path) {
		return "", errors.New("object path escaped object root")
	}
	return path, nil
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && !filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

type PendingBody struct {
	TempPath, FinalPath, Hash, MD5 string
	Size                           int64
}

func (s *Store) StreamObjectBody(provider, service, namespace, key string, body io.Reader) (*PendingBody, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	finalPath, err := s.NewObjectBodyPath(provider, service, namespace, key)
	if err != nil {
		return nil, err
	}
	temp, err := os.CreateTemp(s.tmpRoot, "body-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary body: %w", err)
	}
	hash := sha256.New()
	md5Hash := md5.New()
	size, copyErr := io.Copy(io.MultiWriter(temp, hash, md5Hash), body)
	syncErr := temp.Sync()
	closeErr := temp.Close()
	if copyErr != nil || syncErr != nil || closeErr != nil {
		_ = os.Remove(temp.Name())
		return nil, errors.Join(copyErr, syncErr, closeErr)
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o700); err != nil {
		_ = os.Remove(temp.Name())
		return nil, err
	}
	finalPath += "-" + randomToken()
	if err := os.Rename(temp.Name(), finalPath); err != nil {
		_ = os.Remove(temp.Name())
		return nil, fmt.Errorf("publish object body: %w", err)
	}
	return &PendingBody{TempPath: temp.Name(), FinalPath: finalPath, Hash: hex.EncodeToString(hash.Sum(nil)), MD5: hex.EncodeToString(md5Hash.Sum(nil)), Size: size}, nil
}

func (s *Store) RemoveBody(path string) error {
	if !within(s.objectsRoot, path) {
		return errors.New("refusing to remove path outside object root")
	}
	return os.Remove(path)
}

func (s *Store) Reset(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, table := range []string{"sns_subscriptions", "sns_topics", "dynamodb_items", "dynamodb_attributes", "dynamodb_tables", "sqs_messages", "sqs_queues", "s3_objects", "s3_buckets"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for _, root := range []string{s.objectsRoot, s.tmpRoot} {
		entries, err := os.ReadDir(root)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
				return err
			}
		}
	}
	return os.MkdirAll(filepath.Join(s.objectsRoot, "aws", "s3"), 0o700)
}
