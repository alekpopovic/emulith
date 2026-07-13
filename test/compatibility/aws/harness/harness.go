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
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"github.com/alekpopovic/emulith/providers/aws/dynamodb"
	"github.com/alekpopovic/emulith/providers/aws/s3"
	"github.com/alekpopovic/emulith/providers/aws/sqs"
	"github.com/alekpopovic/emulith/providers/aws/sts"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssdksts "github.com/aws/aws-sdk-go-v2/service/sts"
)

type Harness struct {
	Endpoint string
	S3       *awss3.Client
	SQS      *awssqs.Client
	STS      *awssdksts.Client
	DynamoDB *awsdynamodb.Client
	HTTP     *http.Client
	sequence atomic.Uint64
	mu       sync.Mutex
	dataDir  string
	store    *state.Store
	server   *httptest.Server
}

func New(t *testing.T) *Harness {
	t.Helper()
	h := &Harness{dataDir: t.TempDir()}
	if err := h.start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.close)
	return h
}
func (h *Harness) start() error {
	store, err := state.Open(context.Background(), h.dataDir)
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gateway := awsprovider.NewGateway(store, logger)
	gateway.SetS3(s3.New(store))
	gateway.SetSQS(sqs.New(store))
	gateway.SetSTS(sts.New())
	gateway.SetDynamoDB(dynamodb.New(store))
	app := server.NewWithState(":0", "dev", store, logger, gateway).HTTPServer().Handler
	srv := httptest.NewServer(app)
	parsed, err := url.Parse(srv.URL)
	if err != nil {
		srv.Close()
		store.Close()
		return err
	}
	host, _, err := net.SplitHostPort(parsed.Host)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		srv.Close()
		store.Close()
		return fmt.Errorf("non-loopback compatibility endpoint %q", srv.URL)
	}
	client := &http.Client{Transport: loopbackTransport{base: srv.Client().Transport}}
	cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: client}
	h.Endpoint = srv.URL
	h.store = store
	h.server = srv
	h.HTTP = client
	h.S3 = awss3.NewFromConfig(cfg, func(o *awss3.Options) { o.BaseEndpoint = aws.String(srv.URL); o.UsePathStyle = true })
	h.SQS = awssqs.NewFromConfig(cfg, func(o *awssqs.Options) { o.BaseEndpoint = aws.String(srv.URL) })
	h.STS = awssdksts.NewFromConfig(cfg, func(o *awssdksts.Options) { o.BaseEndpoint = aws.String(srv.URL) })
	h.DynamoDB = awsdynamodb.NewFromConfig(cfg, func(o *awsdynamodb.Options) { o.BaseEndpoint = aws.String(srv.URL) })
	return nil
}
func (h *Harness) close() { h.mu.Lock(); defer h.mu.Unlock(); h.stop() }
func (h *Harness) stop() {
	if h.server != nil {
		h.server.Close()
		h.server = nil
	}
	if h.store != nil {
		_ = h.store.Close()
		h.store = nil
	}
}
func (h *Harness) Restart(t *testing.T) {
	t.Helper()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stop()
	if err := h.start(); err != nil {
		t.Fatal(err)
	}
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
