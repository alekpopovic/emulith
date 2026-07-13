package aws_test

import (
	"context"
	"errors"
	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	emuddb "github.com/alekpopovic/emulith/providers/aws/dynamodb"
	"github.com/alekpopovic/emulith/test/compatibility/aws/compat"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	sdk "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestDynamoDBTableLifecycleSDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.table-lifecycle.basic", func(t *testing.T) {
		ctx := context.Background()
		store, e := state.Open(ctx, t.TempDir())
		if e != nil {
			t.Fatal(e)
		}
		defer store.Close()
		g := awsprovider.NewGateway(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
		g.SetDynamoDB(emuddb.New(store))
		srv := httptest.NewServer(server.New(":0", "dev", g).HTTPServer().Handler)
		defer srv.Close()
		cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: srv.Client()}
		c := sdk.NewFromConfig(cfg, func(o *sdk.Options) { o.BaseEndpoint = aws.String(srv.URL) })
		name := "sdk-table"
		_, e = c.CreateTable(ctx, &sdk.CreateTableInput{TableName: &name, BillingMode: types.BillingModePayPerRequest, AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS}}, KeySchema: []types.KeySchemaElement{{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash}}})
		if e != nil {
			t.Fatal(e)
		}
		d, e := c.DescribeTable(ctx, &sdk.DescribeTableInput{TableName: &name})
		if e != nil || d.Table.TableStatus != types.TableStatusActive {
			t.Fatal(d, e)
		}
		l, e := c.ListTables(ctx, &sdk.ListTablesInput{Limit: aws.Int32(1)})
		if e != nil || len(l.TableNames) != 1 {
			t.Fatal(l, e)
		}
		_, e = c.CreateTable(ctx, &sdk.CreateTableInput{TableName: &name, BillingMode: types.BillingModePayPerRequest, AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS}}, KeySchema: []types.KeySchemaElement{{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash}}})
		var inuse *types.ResourceInUseException
		if !errors.As(e, &inuse) {
			t.Fatalf("expected ResourceInUse: %v", e)
		}
		if _, e = c.DeleteTable(ctx, &sdk.DeleteTableInput{TableName: &name}); e != nil {
			t.Fatal(e)
		}
		_, e = c.DescribeTable(ctx, &sdk.DescribeTableInput{TableName: &name})
		var nf *types.ResourceNotFoundException
		if !errors.As(e, &nf) {
			t.Fatalf("expected not found: %v", e)
		}
	})
}
