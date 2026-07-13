package harness

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"github.com/alekpopovic/emulith/providers/aws/s3"
	"github.com/alekpopovic/emulith/providers/aws/sqs"
	"github.com/alekpopovic/emulith/providers/aws/sts"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssdksts "github.com/aws/aws-sdk-go-v2/service/sts"
)

type Harness struct {
	Endpoint string
	S3       *awss3.Client
	SQS      *awssqs.Client
	STS      *awssdksts.Client
	sequence atomic.Uint64
}

func New(t *testing.T) *Harness {
	t.Helper()
	store, err := state.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gateway := awsprovider.NewGateway(store, logger)
	gateway.SetS3(s3.New(store))
	gateway.SetSQS(sqs.New(store))
	gateway.SetSTS(sts.New())
	app := server.NewWithState(":0", "dev", store, logger, gateway).HTTPServer().Handler
	srv := httptest.NewServer(app)
	parsed, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(parsed.Host)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		srv.Close()
		t.Fatalf("non-loopback compatibility endpoint %q", srv.URL)
	}
	t.Cleanup(func() { srv.Close(); store.Close() })
	client := &http.Client{Transport: loopbackTransport{base: srv.Client().Transport}}
	cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: client}
	h := &Harness{Endpoint: srv.URL}
	h.S3 = awss3.NewFromConfig(cfg, func(o *awss3.Options) { o.BaseEndpoint = aws.String(srv.URL); o.UsePathStyle = true })
	h.SQS = awssqs.NewFromConfig(cfg, func(o *awssqs.Options) { o.BaseEndpoint = aws.String(srv.URL) })
	h.STS = awssdksts.NewFromConfig(cfg, func(o *awssdksts.Options) { o.BaseEndpoint = aws.String(srv.URL) })
	return h
}
func (h *Harness) Name(prefix string) string {
	return fmt.Sprintf("%s-%02d", prefix, h.sequence.Add(1))
}

type loopbackTransport struct{ base http.RoundTripper }

func (t loopbackTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	host := r.URL.Hostname()
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, fmt.Errorf("compatibility transport rejected non-loopback host %q", host)
	}
	return t.base.RoundTrip(r)
}
