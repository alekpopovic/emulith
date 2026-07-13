package aws_test

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"github.com/alekpopovic/emulith/providers/aws/sqs"
	"github.com/alekpopovic/emulith/test/compatibility/aws/compat"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func TestSQSSDKLifecycle(t *testing.T) {
	compat.Run(t, "aws.sqs.lifecycle.basic", func(t *testing.T) {
		ctx := context.Background()
		store, err := state.Open(ctx, t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer store.Close()
		gateway := awsprovider.NewGateway(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
		gateway.SetSQS(sqs.New(store))
		srv := httptest.NewServer(server.New(":0", "dev", gateway).HTTPServer().Handler)
		defer srv.Close()
		cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: srv.Client()}
		client := awssqs.NewFromConfig(cfg, func(o *awssqs.Options) { o.BaseEndpoint = aws.String(srv.URL) })
		created, err := client.CreateQueue(ctx, &awssqs.CreateQueueInput{QueueName: aws.String("sdk-queue")})
		if err != nil {
			t.Fatal(err)
		}
		queueURL := aws.ToString(created.QueueUrl)
		gotURL, err := client.GetQueueUrl(ctx, &awssqs.GetQueueUrlInput{QueueName: aws.String("sdk-queue")})
		if err != nil || aws.ToString(gotURL.QueueUrl) != queueURL {
			t.Fatalf("url=%#v err=%v", gotURL, err)
		}
		listed, err := client.ListQueues(ctx, &awssqs.ListQueuesInput{QueueNamePrefix: aws.String("sdk")})
		if err != nil || len(listed.QueueUrls) != 1 {
			t.Fatalf("list=%#v err=%v", listed, err)
		}
		sent, err := client.SendMessage(ctx, &awssqs.SendMessageInput{QueueUrl: aws.String(queueURL), MessageBody: aws.String("hello")})
		if err != nil || aws.ToString(sent.MessageId) == "" {
			t.Fatalf("send=%#v err=%v", sent, err)
		}
		received, err := client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{QueueUrl: aws.String(queueURL), MaxNumberOfMessages: 1})
		if err != nil || len(received.Messages) != 1 || aws.ToString(received.Messages[0].Body) != "hello" {
			t.Fatalf("receive=%#v err=%v", received, err)
		}
		attrs, err := client.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{QueueUrl: aws.String(queueURL), AttributeNames: []awssqsTypes.QueueAttributeName{awssqsTypes.QueueAttributeNameAll}})
		if err != nil || attrs.Attributes["VisibilityTimeout"] != "30" {
			t.Fatalf("attrs=%#v err=%v", attrs, err)
		}
		if _, err := client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{QueueUrl: aws.String(queueURL), ReceiptHandle: received.Messages[0].ReceiptHandle}); err != nil {
			t.Fatal(err)
		}
		if _, err := client.PurgeQueue(ctx, &awssqs.PurgeQueueInput{QueueUrl: aws.String(queueURL)}); err != nil {
			t.Fatal(err)
		}
	})
}
