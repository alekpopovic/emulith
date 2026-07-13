package aws_test

import (
	"context"
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

func queryClient(t *testing.T) (context.Context, *sdk.Client) {
	t.Helper()
	ctx := context.Background()
	store, e := state.Open(ctx, t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { store.Close() })
	g := awsprovider.NewGateway(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	g.SetDynamoDB(emuddb.New(store))
	srv := httptest.NewServer(server.New(":0", "dev", g).HTTPServer().Handler)
	t.Cleanup(srv.Close)
	cfg := aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), HTTPClient: srv.Client()}
	return ctx, sdk.NewFromConfig(cfg, func(o *sdk.Options) { o.BaseEndpoint = aws.String(srv.URL) })
}
func seedQueryTable(t *testing.T, ctx context.Context, c *sdk.Client) string {
	t.Helper()
	name := "query-table"
	_, e := c.CreateTable(ctx, &sdk.CreateTableInput{TableName: &name, BillingMode: types.BillingModePayPerRequest, AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS}, {AttributeName: aws.String("sk"), AttributeType: types.ScalarAttributeTypeN}}, KeySchema: []types.KeySchemaElement{{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash}, {AttributeName: aws.String("sk"), KeyType: types.KeyTypeRange}}})
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"2", "10", "20"} {
		_, e = c.PutItem(ctx, &sdk.PutItemInput{TableName: &name, Item: map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: "p"}, "sk": &types.AttributeValueMemberN{Value: n}, "keep": &types.AttributeValueMemberBOOL{Value: n != "10"}}})
		if e != nil {
			t.Fatal(e)
		}
	}
	return name
}
func TestDynamoDBQuerySDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.Query.primary-index", func(t *testing.T) {
		ctx, c := queryClient(t)
		name := seedQueryTable(t, ctx, c)
		in := &sdk.QueryInput{TableName: &name, KeyConditionExpression: aws.String("#p = :p AND #s BETWEEN :lo AND :hi"), ExpressionAttributeNames: map[string]string{"#p": "pk", "#s": "sk"}, ExpressionAttributeValues: map[string]types.AttributeValue{":p": &types.AttributeValueMemberS{Value: "p"}, ":lo": &types.AttributeValueMemberN{Value: "2"}, ":hi": &types.AttributeValueMemberN{Value: "20"}}, Limit: aws.Int32(2)}
		first, e := c.Query(ctx, in)
		if e != nil || len(first.Items) != 2 || first.Items[0]["sk"].(*types.AttributeValueMemberN).Value != "2" || first.LastEvaluatedKey == nil {
			t.Fatal(first, e)
		}
		in.ExclusiveStartKey = first.LastEvaluatedKey
		second, e := c.Query(ctx, in)
		if e != nil || len(second.Items) != 1 || second.Items[0]["sk"].(*types.AttributeValueMemberN).Value != "20" {
			t.Fatal(second, e)
		}
	})
}
func TestDynamoDBScanSDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.Scan.pagination", func(t *testing.T) {
		ctx, c := queryClient(t)
		name := seedQueryTable(t, ctx, c)
		out, e := c.Scan(ctx, &sdk.ScanInput{TableName: &name, Limit: aws.Int32(2), FilterExpression: aws.String("#k = :yes"), ExpressionAttributeNames: map[string]string{"#k": "keep"}, ExpressionAttributeValues: map[string]types.AttributeValue{":yes": &types.AttributeValueMemberBOOL{Value: true}}, ProjectionExpression: aws.String("pk, sk")})
		if e != nil || out.ScannedCount != 2 || out.Count != 1 || out.LastEvaluatedKey == nil {
			t.Fatal(out, e)
		}
	})
}
