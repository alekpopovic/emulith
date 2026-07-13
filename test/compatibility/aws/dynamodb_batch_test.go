package aws_test

import (
	"context"
	"github.com/alekpopovic/emulith/test/compatibility/aws/compat"
	"github.com/aws/aws-sdk-go-v2/aws"
	sdk "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"testing"
)

func createBatchTables(t *testing.T, c *sdk.Client) (string, string) {
	t.Helper()
	ctx := context.Background()
	names := []string{"batch-one", "batch-two"}
	for _, name := range names {
		_, e := c.CreateTable(ctx, &sdk.CreateTableInput{TableName: &name, BillingMode: types.BillingModePayPerRequest, AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS}}, KeySchema: []types.KeySchemaElement{{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash}}})
		if e != nil {
			t.Fatal(e)
		}
	}
	return names[0], names[1]
}
func TestDynamoDBBatchWriteSDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.BatchWriteItem.multi-table", func(t *testing.T) {
		ctx, c := queryClient(t)
		a, b := createBatchTables(t, c)
		key := func(s string) map[string]types.AttributeValue {
			return map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: s}}
		}
		out, e := c.BatchWriteItem(ctx, &sdk.BatchWriteItemInput{RequestItems: map[string][]types.WriteRequest{a: {{PutRequest: &types.PutRequest{Item: key("a")}}}, b: {{PutRequest: &types.PutRequest{Item: key("b")}}}}})
		if e != nil || len(out.UnprocessedItems) != 0 {
			t.Fatal(out, e)
		}
	})
}
func TestDynamoDBBatchGetSDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.BatchGetItem.multi-table", func(t *testing.T) {
		ctx, c := queryClient(t)
		a, b := createBatchTables(t, c)
		key := func(s string) map[string]types.AttributeValue {
			return map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: s}}
		}
		for table, k := range map[string]string{a: "a", b: "b"} {
			_, e := c.PutItem(ctx, &sdk.PutItemInput{TableName: &table, Item: key(k)})
			if e != nil {
				t.Fatal(e)
			}
		}
		out, e := c.BatchGetItem(ctx, &sdk.BatchGetItemInput{RequestItems: map[string]types.KeysAndAttributes{a: {Keys: []map[string]types.AttributeValue{key("a"), key("missing")}, ConsistentRead: aws.Bool(true)}, b: {Keys: []map[string]types.AttributeValue{key("b")}}}})
		if e != nil || len(out.UnprocessedKeys) != 0 || len(out.Responses[a]) != 1 || len(out.Responses[b]) != 1 {
			t.Fatal(out, e)
		}
	})
}
