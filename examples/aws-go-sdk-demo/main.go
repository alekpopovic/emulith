package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	awssdksts "github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

const bucket = "emulith-demo-bucket"
const queue = "emulith-demo-queue"

func main() {
	endpoint := flag.String("endpoint", "http://127.0.0.1:4566", "Emulith endpoint")
	verifyReset := flag.Bool("verify-reset", false, "verify demo resources were reset")
	flag.Parse()
	if err := run(context.Background(), *endpoint, *verifyReset); err != nil {
		fmt.Fprintln(os.Stderr, "demo failed:", err)
		os.Exit(1)
	}
}
func run(ctx context.Context, endpoint string, verifyReset bool) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	ip := net.ParseIP(u.Hostname())
	if u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return fmt.Errorf("refusing non-loopback endpoint %q", u.Hostname())
	}
	client := &http.Client{Timeout: 10 * time.Second}
	if err := health(ctx, client, endpoint); err != nil {
		return err
	}
	cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: client}
	s3c := awss3.NewFromConfig(cfg, func(o *awss3.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	sqsc := awssqs.NewFromConfig(cfg, func(o *awssqs.Options) { o.BaseEndpoint = aws.String(endpoint) })
	stsc := awssdksts.NewFromConfig(cfg, func(o *awssdksts.Options) { o.BaseEndpoint = aws.String(endpoint) })
	if verifyReset {
		return verify(ctx, s3c, sqsc)
	}
	identity, err := stsc.GetCallerIdentity(ctx, &awssdksts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}
	fmt.Printf("STS account=%s arn=%s\n", aws.ToString(identity.Account), aws.ToString(identity.Arn))
	if _, err := s3c.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		return err
	}
	objects := map[string][]byte{"demo/text.txt": []byte("hello from Emulith\n"), "demo/binary.bin": {0, 1, 2, 255}}
	for key, body := range objects {
		if _, err := s3c.PutObject(ctx, &awss3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(body)}); err != nil {
			return err
		}
		got, err := s3c.GetObject(ctx, &awss3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
		if err != nil {
			return err
		}
		actual, readErr := io.ReadAll(got.Body)
		got.Body.Close()
		if readErr != nil || !bytes.Equal(actual, body) {
			return fmt.Errorf("object %s mismatch", key)
		}
	}
	listed, err := s3c.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String("demo/")})
	if err != nil || len(listed.Contents) != 2 {
		return fmt.Errorf("S3 list mismatch: %w", err)
	}
	if _, err := s3c.HeadObject(ctx, &awss3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String("demo/text.txt")}); err != nil {
		return err
	}
	if _, err := s3c.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String("demo/binary.bin")}); err != nil {
		return err
	}
	created, err := sqsc.CreateQueue(ctx, &awssqs.CreateQueueInput{QueueName: aws.String(queue)})
	if err != nil {
		return err
	}
	sent, err := sqsc.SendMessage(ctx, &awssqs.SendMessageInput{QueueUrl: created.QueueUrl, MessageBody: aws.String("demo message")})
	if err != nil || aws.ToString(sent.MessageId) == "" {
		return fmt.Errorf("SQS send failed: %w", err)
	}
	received, err := sqsc.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{QueueUrl: created.QueueUrl})
	if err != nil || len(received.Messages) != 1 || aws.ToString(received.Messages[0].Body) != "demo message" {
		return fmt.Errorf("SQS receive mismatch: %w", err)
	}
	attrs, err := sqsc.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{QueueUrl: created.QueueUrl, AttributeNames: []awssqstypes.QueueAttributeName{awssqstypes.QueueAttributeNameAll}})
	if err != nil || attrs.Attributes["VisibilityTimeout"] != "30" {
		return fmt.Errorf("SQS attributes mismatch: %w", err)
	}
	if _, err := sqsc.DeleteMessage(ctx, &awssqs.DeleteMessageInput{QueueUrl: created.QueueUrl, ReceiptHandle: received.Messages[0].ReceiptHandle}); err != nil {
		return err
	}
	fmt.Println("S3 and SQS demo flows passed")
	return nil
}
func health(ctx context.Context, client *http.Client, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/_emulith/health", nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("health HTTP %d", response.StatusCode)
	}
	return nil
}
func verify(ctx context.Context, s3c *awss3.Client, sqsc *awssqs.Client) error {
	listed, err := s3c.ListBuckets(ctx, &awss3.ListBucketsInput{})
	if err != nil {
		return err
	}
	for _, b := range listed.Buckets {
		if aws.ToString(b.Name) == bucket {
			return errors.New("demo bucket survived reset")
		}
	}
	_, err = sqsc.GetQueueUrl(ctx, &awssqs.GetQueueUrlInput{QueueName: aws.String(queue)})
	if err == nil {
		return errors.New("demo queue survived reset")
	}
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) || apiErr.ErrorCode() != "AWS.SimpleQueueService.NonExistentQueue" {
		return fmt.Errorf("unexpected queue reset error: %w", err)
	}
	fmt.Println("Reset verification passed")
	return nil
}
