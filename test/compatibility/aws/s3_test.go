package aws_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"github.com/alekpopovic/emulith/providers/aws/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestS3SDKLifecycle(t *testing.T) {
	ctx := context.Background()
	store, err := state.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	gateway := awsprovider.NewGateway(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	gateway.SetS3(s3.New(store))
	srv := httptest.NewServer(server.New(":0", "dev", gateway).HTTPServer().Handler)
	defer srv.Close()
	cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: srv.Client()}
	client := awss3.NewFromConfig(cfg, func(o *awss3.Options) { o.BaseEndpoint = aws.String(srv.URL); o.UsePathStyle = true })
	bucket, key := "sdk-bucket", "nested/a.txt"
	if _, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatal(err)
	}
	listed, err := client.ListBuckets(ctx, &awss3.ListBucketsInput{})
	if err != nil || len(listed.Buckets) != 1 {
		t.Fatalf("list: %#v %v", listed, err)
	}
	payload := []byte{0, 1, 2, 'x'}
	if _, err := client.PutObject(ctx, &awss3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(payload), ContentType: aws.String("application/octet-stream")}); err != nil {
		t.Fatal(err)
	}
	head, err := client.HeadObject(ctx, &awss3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil || head.ContentLength == nil || *head.ContentLength != int64(len(payload)) {
		t.Fatalf("head: %#v %v", head, err)
	}
	got, err := client.GetObject(ctx, &awss3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(got.Body)
	got.Body.Close()
	if err != nil || !bytes.Equal(body, payload) {
		t.Fatalf("body=%v err=%v", body, err)
	}
	objects, err := client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String("nested/")})
	if err != nil || len(objects.Contents) != 1 {
		t.Fatalf("objects=%#v err=%v", objects, err)
	}
	if _, err := client.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetObject(ctx, &awss3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}); err == nil {
		t.Fatal("expected missing object")
	}
}
