package state

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const SnapshotFormatVersion = 1
const maxSnapshotEntries = 10000
const maxSnapshotFile = 1 << 30
const maxSnapshotTotal = 4 << 30

type SnapshotManifest struct {
	FormatVersion         int            `json:"format_version"`
	EmulithVersion        string         `json:"emulith_version"`
	CreatedAt             time.Time      `json:"created_at"`
	DatabaseSchemaVersion int            `json:"database_schema_version"`
	Files                 []SnapshotFile `json:"files"`
}
type SnapshotFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

func (s *Store) Export(ctx context.Context, out io.Writer, version string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stage, err := os.MkdirTemp(s.dataDir, ".export-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	dbCopy := filepath.Join(stage, "emulith.db")
	escaped := strings.ReplaceAll(dbCopy, "'", "''")
	if _, err = s.db.ExecContext(ctx, "VACUUM INTO '"+escaped+"'"); err != nil {
		return fmt.Errorf("snapshot database: %w", err)
	}
	files := []struct{ archive, local string }{{"metadata/emulith.db", dbCopy}}
	err = filepath.WalkDir(s.objectsRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("snapshot refuses symlinks")
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return errors.New("snapshot refuses special files")
		}
		rel, e := filepath.Rel(s.objectsRoot, path)
		if e != nil {
			return e
		}
		files = append(files, struct{ archive, local string }{"objects/" + filepath.ToSlash(rel), path})
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].archive < files[j].archive })
	manifest := SnapshotManifest{FormatVersion: 1, EmulithVersion: version, CreatedAt: now.UTC(), DatabaseSchemaVersion: 1}
	for _, f := range files {
		info, e := os.Stat(f.local)
		if e != nil {
			return e
		}
		sum, e := fileSHA(f.local)
		if e != nil {
			return e
		}
		manifest.Files = append(manifest.Files, SnapshotFile{Path: f.archive, Size: info.Size(), SHA256: sum})
	}
	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	manifestData, _ := json.Marshal(manifest)
	if err := writeTar(tw, "emulith-snapshot/manifest.json", manifestData, now); err != nil {
		return err
	}
	for _, f := range files {
		if err := writeTarFile(tw, "emulith-snapshot/"+f.archive, f.local, now); err != nil {
			return err
		}
	}
	return errors.Join(tw.Close(), gz.Close())
}

func ValidateSnapshot(in io.Reader) (SnapshotManifest, error) {
	gz, err := gzip.NewReader(in)
	if err != nil {
		return SnapshotManifest{}, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	payload := map[string]SnapshotFile{}
	var manifest SnapshotManifest
	var total int64
	entries := 0
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return manifest, err
		}
		entries++
		if entries > maxSnapshotEntries {
			return manifest, errors.New("snapshot has too many entries")
		}
		name := h.Name
		if h.Typeflag != tar.TypeReg {
			return manifest, errors.New("snapshot contains non-regular entry")
		}
		if filepath.IsAbs(name) || strings.Contains(name, "\\") || strings.Contains("/"+name+"/", "/../") {
			return manifest, errors.New("unsafe snapshot path")
		}
		lower := strings.ToLower(name)
		if seen[lower] {
			return manifest, errors.New("duplicate snapshot path")
		}
		seen[lower] = true
		if h.Size < 0 || h.Size > maxSnapshotFile || total+h.Size > maxSnapshotTotal {
			return manifest, errors.New("snapshot size limit exceeded")
		}
		total += h.Size
		data, err := io.ReadAll(io.LimitReader(tr, h.Size+1))
		if err != nil || int64(len(data)) != h.Size {
			return manifest, errors.New("truncated snapshot")
		}
		rel := strings.TrimPrefix(name, "emulith-snapshot/")
		if rel == "manifest.json" {
			if err := json.Unmarshal(data, &manifest); err != nil {
				return manifest, err
			}
			continue
		}
		sum := sha256.Sum256(data)
		payload[rel] = SnapshotFile{Path: rel, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:])}
	}
	if manifest.FormatVersion != 1 {
		return manifest, fmt.Errorf("unsupported snapshot format %d", manifest.FormatVersion)
	}
	if manifest.DatabaseSchemaVersion != 1 {
		return manifest, fmt.Errorf("unsupported database schema %d", manifest.DatabaseSchemaVersion)
	}
	if len(payload) != len(manifest.Files) {
		return manifest, errors.New("snapshot manifest file count mismatch")
	}
	for _, want := range manifest.Files {
		got, ok := payload[want.Path]
		if !ok || got.Size != want.Size || got.SHA256 != want.SHA256 {
			return manifest, fmt.Errorf("snapshot checksum mismatch for %q", want.Path)
		}
	}
	return manifest, nil
}
func (s *Store) Import(ctx context.Context, in io.Reader, replace bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	archive, err := os.CreateTemp(s.tmpRoot, "import-*.tar.gz")
	if err != nil {
		return err
	}
	archivePath := archive.Name()
	defer os.Remove(archivePath)
	if _, err = io.Copy(archive, io.LimitReader(in, maxSnapshotTotal+1)); err != nil {
		archive.Close()
		return err
	}
	if err = archive.Close(); err != nil {
		return err
	}
	check, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	_, err = ValidateSnapshot(check)
	check.Close()
	if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(s.tmpRoot, "import-stage-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err = extractSnapshot(archivePath, stage); err != nil {
		return err
	}
	var existing int
	if err = s.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM s3_buckets)+(SELECT count(*) FROM sqs_queues)+(SELECT count(*) FROM dynamodb_tables)`).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 && !replace {
		return errors.New("target state is not empty; use replace")
	}
	importDB := filepath.Join(stage, "metadata", "emulith.db")
	candidate, err := sql.Open("sqlite", importDB)
	if err != nil {
		return err
	}
	var schema int
	err = candidate.QueryRowContext(ctx, `SELECT max(version) FROM schema_version`).Scan(&schema)
	candidate.Close()
	if err != nil || schema != 2 {
		return errors.New("import database schema is unsupported")
	}
	backup := s.objectsRoot + ".import-backup"
	_ = os.RemoveAll(backup)
	if err = os.Rename(s.objectsRoot, backup); err != nil {
		return err
	}
	if err = os.Rename(filepath.Join(stage, "objects"), s.objectsRoot); err != nil {
		_ = os.Rename(backup, s.objectsRoot)
		return err
	}
	if err = restoreDatabase(ctx, s.db, importDB); err != nil {
		_ = os.RemoveAll(s.objectsRoot)
		_ = os.Rename(backup, s.objectsRoot)
		return err
	}
	if err = s.rebaseImportedBodies(ctx); err != nil {
		return err
	}
	_ = os.RemoveAll(backup)
	return os.MkdirAll(filepath.Join(s.objectsRoot, "aws", "s3"), 0o700)
}
func extractSnapshot(path, stage string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	gz, e := gzip.NewReader(f)
	if e != nil {
		return e
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, e := tr.Next()
		if errors.Is(e, io.EOF) {
			return nil
		}
		if e != nil {
			return e
		}
		rel := strings.TrimPrefix(h.Name, "emulith-snapshot/")
		if rel == "manifest.json" {
			continue
		}
		target := filepath.Join(stage, filepath.FromSlash(rel))
		if !within(stage, target) {
			return errors.New("unsafe import path")
		}
		if e = os.MkdirAll(filepath.Dir(target), 0o700); e != nil {
			return e
		}
		out, e := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if e != nil {
			return e
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if e = errors.Join(copyErr, closeErr); e != nil {
			return e
		}
	}
}
func restoreDatabase(ctx context.Context, db *sql.DB, path string) error {
	escaped := strings.ReplaceAll(path, "'", "''")
	if _, e := db.ExecContext(ctx, "ATTACH DATABASE '"+escaped+"' AS imported"); e != nil {
		return e
	}
	defer db.ExecContext(context.Background(), "DETACH DATABASE imported")
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, table := range []string{"dynamodb_items", "dynamodb_attributes", "dynamodb_tables", "sqs_messages", "sqs_queues", "s3_objects", "s3_buckets"} {
		if _, e = tx.ExecContext(ctx, "DELETE FROM "+table); e != nil {
			return e
		}
	}
	for _, table := range []string{"s3_buckets", "s3_objects", "sqs_queues", "sqs_messages", "dynamodb_tables", "dynamodb_attributes", "dynamodb_items"} {
		if _, e = tx.ExecContext(ctx, "INSERT INTO "+table+" SELECT * FROM imported."+table); e != nil {
			return e
		}
	}
	return tx.Commit()
}
func (s *Store) rebaseImportedBodies(ctx context.Context) error {
	rows, e := s.db.QueryContext(ctx, `SELECT bucket,key,body_path FROM s3_objects`)
	if e != nil {
		return e
	}
	type row struct{ bucket, key, path string }
	var all []row
	for rows.Next() {
		var r row
		if e = rows.Scan(&r.bucket, &r.key, &r.path); e != nil {
			rows.Close()
			return e
		}
		all = append(all, r)
	}
	if e = rows.Close(); e != nil {
		return e
	}
	for _, r := range all {
		clean := filepath.Clean(r.path)
		base := filepath.Base(clean)
		parent := filepath.Base(filepath.Dir(clean))
		grand := filepath.Base(filepath.Dir(filepath.Dir(clean)))
		path := filepath.Join(s.objectsRoot, grand, parent, base)
		if !within(s.objectsRoot, path) {
			return errors.New("imported body path is unsafe")
		}
		if _, e = os.Stat(path); e != nil {
			return errors.New("imported object body is missing")
		}
		if _, e = s.db.ExecContext(ctx, `UPDATE s3_objects SET body_path=? WHERE bucket=? AND key=?`, path, r.bucket, r.key); e != nil {
			return e
		}
	}
	return nil
}
func fileSHA(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	_, e = io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), e
}
func writeTar(tw *tar.Writer, name string, data []byte, now time.Time) error {
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(data)), ModTime: now.UTC(), Typeflag: tar.TypeReg}); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}
func writeTarFile(tw *tar.Writer, name, path string, now time.Time) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		return e
	}
	if e = tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: info.Size(), ModTime: now.UTC(), Typeflag: tar.TypeReg}); e != nil {
		return e
	}
	_, e = io.Copy(tw, f)
	return e
}
