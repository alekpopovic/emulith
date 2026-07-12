package aws_test

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssdksts "github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/emulith/emulith/internal/server"
	"github.com/emulith/emulith/internal/state"
	awsprovider "github.com/emulith/emulith/providers/aws"
	"github.com/emulith/emulith/providers/aws/sts"
)

func TestSTSGetCallerIdentitySDK(t *testing.T) {
	store, err := state.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	gateway := awsprovider.NewGateway(store, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	gateway.SetSTS(sts.New())
	httpServer := httptest.NewServer(server.New(":0", "dev", gateway).HTTPServer().Handler)
	defer httpServer.Close()
	cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: httpServer.Client()}
	client := awssdksts.NewFromConfig(cfg, func(o *awssdksts.Options) { o.BaseEndpoint = aws.String(httpServer.URL) })
	got, err := client.GetCallerIdentity(context.Background(), &awssdksts.GetCallerIdentityInput{})
	if err != nil {
		t.Fatal(err)
	}
	if aws.ToString(got.Account) != "000000000000" || aws.ToString(got.Arn) != "arn:aws:iam::000000000000:user/emulith" || aws.ToString(got.UserId) != "EMULITHUSER" {
		t.Fatalf("unexpected identity: %#v", got)
	}
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) { w.t.Log(string(p)); return len(p), nil }
