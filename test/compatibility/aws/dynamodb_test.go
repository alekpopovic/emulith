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

func TestDynamoDBItemCRUDSDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.item-crud.basic", func(t *testing.T) {
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
		name := "crud-table"
		_, e = c.CreateTable(ctx, &sdk.CreateTableInput{TableName: &name, BillingMode: types.BillingModePayPerRequest, AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS}}, KeySchema: []types.KeySchemaElement{{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash}}})
		if e != nil {
			t.Fatal(e)
		}
		key := map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: "one"}}
		item := map[string]types.AttributeValue{"pk": key["pk"], "precise": &types.AttributeValueMemberN{Value: "12345678901234567890.001"}, "nested": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{"ok": &types.AttributeValueMemberBOOL{Value: true}, "nil": &types.AttributeValueMemberNULL{Value: true}, "bin": &types.AttributeValueMemberB{Value: []byte{0, 1}}}}}
		if _, e = c.PutItem(ctx, &sdk.PutItemInput{TableName: &name, Item: item}); e != nil {
			t.Fatal(e)
		}
		got, e := c.GetItem(ctx, &sdk.GetItemInput{TableName: &name, Key: key, ConsistentRead: aws.Bool(true)})
		if e != nil || got.Item["precise"].(*types.AttributeValueMemberN).Value != "12345678901234567890.001" {
			t.Fatal(got, e)
		}
		rv := types.ReturnValueAllNew
		_, e = c.UpdateItem(ctx, &sdk.UpdateItemInput{TableName: &name, Key: key, UpdateExpression: aws.String("SET #n = :v"), ExpressionAttributeNames: map[string]string{"#n": "name"}, ExpressionAttributeValues: map[string]types.AttributeValue{":v": &types.AttributeValueMemberS{Value: "updated"}}, ReturnValues: rv})
		if e != nil {
			t.Fatal(e)
		}
		old := types.ReturnValueAllOld
		d, e := c.DeleteItem(ctx, &sdk.DeleteItemInput{TableName: &name, Key: key, ReturnValues: old})
		if e != nil || len(d.Attributes) == 0 {
			t.Fatal(d, e)
		}
		missing, e := c.GetItem(ctx, &sdk.GetItemInput{TableName: &name, Key: key})
		if e != nil || missing.Item != nil {
			t.Fatal(missing, e)
		}
	})
}
