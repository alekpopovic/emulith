package state

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOpenMigrationsReopenAndReset(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	s, err := Open(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{s.DataDir(), s.ObjectsRoot(), filepath.Join(root, "tmp"), filepath.Join(root, "emulith.db")} {
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO s3_buckets(name, region, created_at) VALUES('bucket','us-east-1',?)`, now); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM s3_buckets`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if err := s.Reset(ctx); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"s3_buckets", "s3_objects", "sqs_queues", "sqs_messages"} {
		if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM `+table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s count=%d err=%v", table, count, err)
		}
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO s3_buckets(name, region, created_at) VALUES('after','us-east-1',?)`, now); err != nil {
		t.Fatal(err)
	}
}

func TestObjectPathsAndResetSymlink(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, key := range []string{"../../escape", "a/b", "..", "ü nicode", "%2e%2e"} {
		path, err := s.NewObjectBodyPath("aws", "s3", "bucket", key)
		if err != nil || !within(s.ObjectsRoot(), path) || strings.Contains(path, key) {
			t.Fatalf("key=%q path=%q err=%v", key, path, err)
		}
	}
	pending, err := s.StreamObjectBody("aws", "s3", "bucket", "key", strings.NewReader("body"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(pending.FinalPath); err != nil {
		t.Fatal(err)
	}
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep")
	if err := os.WriteFile(sentinel, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(s.ObjectsRoot(), "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := s.Reset(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("external file removed: %v", err)
	}
	if _, err := os.Stat(s.DataDir()); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentOpen(t *testing.T) {
	root := t.TempDir()
	ctx := context.Background()
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := Open(ctx, root)
			if err == nil {
				err = s.Close()
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRejectUnsafeRoots(t *testing.T) {
	for _, root := range []string{"", string(filepath.Separator)} {
		if s, err := Open(context.Background(), root); err == nil {
			s.Close()
			t.Fatalf("accepted %q", root)
		}
	}
}
