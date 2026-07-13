package manifest

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/emulith/emulith/test/compatibility/aws/harness"
)

func TestParse(t *testing.T) {
	_, err := Parse([]byte("version: 1\nproject: demo\nresources:\n- type: aws.s3.bucket\n  name: invoices\n- type: aws.sqs.queue\n  name: events\n"))
	if err != nil {
		t.Fatal(err)
	}
}
func TestValidationFailures(t *testing.T) {
	cases := []string{"version: 2\nproject: demo\n", "version: 1\nproject: ''\n", "version: 1\nproject: demo\nunknown: true\n", "version: 1\nproject: demo\nresources:\n- type: aws.unknown\n  name: x\n", "version: 1\nproject: demo\nresources:\n- type: aws.sqs.queue\n  name: x.fifo\n", "version: 1\nproject: demo\nresources:\n- type: aws.s3.bucket\n  name: Bad\n", "version: 1\nproject: demo\nresources:\n- type: aws.sqs.queue\n  name: x\n- type: aws.sqs.queue\n  name: x\n"}
	for _, input := range cases {
		if _, err := Parse([]byte(input)); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}
func TestEndpointGuard(t *testing.T) {
	if err := ValidateEndpoint("https://s3.amazonaws.com"); err == nil {
		t.Fatal("accepted public endpoint")
	}
	for _, v := range []string{"http://localhost:4566", "http://127.0.0.1:1", "http://[::1]:2"} {
		if err := ValidateEndpoint(v); err != nil {
			t.Fatal(err)
		}
	}
}

func TestApplyIsIdempotent(t *testing.T) {
	h := harness.New(t)
	m, err := Parse([]byte("version: 1\nproject: demo\nresources:\n- type: aws.s3.bucket\n  name: manifest-bucket\n- type: aws.sqs.queue\n  name: manifest-queue\n"))
	if err != nil {
		t.Fatal(err)
	}
	a := &Applier{S3: h.S3, SQS: h.SQS}
	var out bytes.Buffer
	for i := 0; i < 2; i++ {
		if err := a.Apply(context.Background(), m, &out); err != nil {
			t.Fatal(err)
		}
	}
	buckets, err := h.S3.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil || len(buckets.Buckets) != 1 {
		t.Fatalf("buckets=%#v err=%v", buckets, err)
	}
	queues, err := h.SQS.ListQueues(context.Background(), &sqs.ListQueuesInput{})
	if err != nil || len(queues.QueueUrls) != 1 {
		t.Fatalf("queues=%#v err=%v", queues, err)
	}
}
