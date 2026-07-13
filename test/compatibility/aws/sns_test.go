package aws_test

import (
	"context"
	"errors"
	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"github.com/alekpopovic/emulith/providers/aws/sns"
	"github.com/alekpopovic/emulith/test/compatibility/aws/compat"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestSNSSDKLifecycle(t *testing.T) {
	compat.Run(t, "aws.sns.topic-publish.lifecycle", func(t *testing.T) {
		ctx := context.Background()
		store, e := state.Open(ctx, t.TempDir())
		if e != nil {
			t.Fatal(e)
		}
		defer store.Close()
		g := awsprovider.NewGateway(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
		g.SetSNS(sns.New(store, "us-east-1"))
		srv := httptest.NewServer(server.New(":0", "dev", g).HTTPServer().Handler)
		defer srv.Close()
		cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: srv.Client()}
		c := awssns.NewFromConfig(cfg, func(o *awssns.Options) { o.BaseEndpoint = aws.String(srv.URL) })
		name := "sdk-topic"
		created, e := c.CreateTopic(ctx, &awssns.CreateTopicInput{Name: &name})
		if e != nil || created.TopicArn == nil {
			t.Fatal(created, e)
		}
		again, e := c.CreateTopic(ctx, &awssns.CreateTopicInput{Name: &name})
		if e != nil || aws.ToString(again.TopicArn) != aws.ToString(created.TopicArn) {
			t.Fatal(again, e)
		}
		attrs, e := c.GetTopicAttributes(ctx, &awssns.GetTopicAttributesInput{TopicArn: created.TopicArn})
		if e != nil || attrs.Attributes["DisplayName"] != "sdk-topic" {
			t.Fatal(attrs, e)
		}
		pub, e := c.Publish(ctx, &awssns.PublishInput{TopicArn: created.TopicArn, Message: aws.String("hello")})
		if e != nil || pub.MessageId == nil {
			t.Fatal(pub, e)
		}
		if _, e = c.DeleteTopic(ctx, &awssns.DeleteTopicInput{TopicArn: created.TopicArn}); e != nil {
			t.Fatal(e)
		}
		_, e = c.GetTopicAttributes(ctx, &awssns.GetTopicAttributesInput{TopicArn: created.TopicArn})
		var nf *types.NotFoundException
		if !errors.As(e, &nf) {
			t.Fatalf("expected not found: %v", e)
		}
	})
}
