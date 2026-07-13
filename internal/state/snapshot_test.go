package state

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSnapshotRoundTrip(t *testing.T) {
	ctx := context.Background()
	source, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	now := time.Unix(1000, 0).UTC()
	if err := source.CreateS3Bucket(ctx, S3Bucket{Name: "bucket", Region: "us-east-1", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	body, err := source.StreamObjectBody("aws", "s3", "bucket", "key", strings.NewReader("binary\x00body"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = source.PutS3Object(ctx, S3Object{Bucket: "bucket", Key: "key", ETag: "etag", Size: body.Size, LastModified: now, BodyPath: body.FinalPath}); err != nil {
		t.Fatal(err)
	}
	q, err := source.CreateSQSQueue(ctx, SQSQueue{Name: "queue", URLPath: "/000000000000/queue", VisibilityTimeout: 30, CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	source.SendSQSMessage(ctx, q.Name, "message", now)
	var archive bytes.Buffer
	if err = source.Export(ctx, &archive, "dev", now); err != nil {
		t.Fatal(err)
	}
	if _, err = ValidateSnapshot(bytes.NewReader(archive.Bytes())); err != nil {
		t.Fatal(err)
	}
	target, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err = target.Import(ctx, bytes.NewReader(archive.Bytes()), false); err != nil {
		t.Fatal(err)
	}
	object, err := target.GetS3Object(ctx, "bucket", "key")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(object.BodyPath)
	if err != nil || string(data) != "binary\x00body" {
		t.Fatalf("body=%q err=%v", data, err)
	}
	queues, err := target.ListSQSQueues(ctx, "")
	if err != nil || len(queues) != 1 {
		t.Fatalf("queues=%#v err=%v", queues, err)
	}
}
func TestSnapshotRejectsTraversalAndLinks(t *testing.T) {
	for _, kind := range []byte{tar.TypeReg, tar.TypeSymlink} {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		name := "../escape"
		if kind == tar.TypeSymlink {
			name = "emulith-snapshot/link"
		}
		tw.WriteHeader(&tar.Header{Name: name, Typeflag: kind, Linkname: "/tmp", Size: 0})
		tw.Close()
		gz.Close()
		if _, err := ValidateSnapshot(bytes.NewReader(buf.Bytes())); err == nil {
			t.Fatalf("accepted type %d", kind)
		}
	}
}
