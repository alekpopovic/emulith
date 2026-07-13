package aws_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/alekpopovic/emulith/test/compatibility/aws/compat"
	"github.com/alekpopovic/emulith/test/compatibility/aws/harness"
	"github.com/aws/aws-sdk-go-v2/aws"
	sdk "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

func TestDynamoDBPersistenceRestartSDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.persistence.restart", func(t *testing.T) {
		ctx := context.Background()
		h := harness.New(t)
		name := "persist-table"
		_, e := h.DynamoDB.CreateTable(ctx, &sdk.CreateTableInput{TableName: &name, BillingMode: types.BillingModePayPerRequest, AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS}}, KeySchema: []types.KeySchemaElement{{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash}}})
		if e != nil {
			t.Fatal(e)
		}
		key := map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: "saved"}}
		if _, e = h.DynamoDB.PutItem(ctx, &sdk.PutItemInput{TableName: &name, Item: key}); e != nil {
			t.Fatal(e)
		}
		h.Restart(t)
		got, e := h.DynamoDB.GetItem(ctx, &sdk.GetItemInput{TableName: &name, Key: key})
		if e != nil || len(got.Item) != 1 {
			t.Fatal(got, e)
		}
		req, e := http.NewRequest(http.MethodPost, h.Endpoint+"/_emulith/reset", bytes.NewReader(nil))
		if e != nil {
			t.Fatal(e)
		}
		resp, e := h.HTTP.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal(resp.Status)
		}
		listed, e := h.DynamoDB.ListTables(ctx, &sdk.ListTablesInput{})
		if e != nil || len(listed.TableNames) != 0 {
			t.Fatal(listed, e)
		}
	})
}
func TestDynamoDBConcurrentConditionalUpdateSDK(t *testing.T) {
	compat.Run(t, "aws.dynamodb.UpdateItem.concurrent-condition", func(t *testing.T) {
		ctx, c := queryClient(t)
		name := "concurrent-table"
		_, e := c.CreateTable(ctx, &sdk.CreateTableInput{TableName: &name, BillingMode: types.BillingModePayPerRequest, AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS}}, KeySchema: []types.KeySchemaElement{{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash}}})
		if e != nil {
			t.Fatal(e)
		}
		key := map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: "one"}}
		item := map[string]types.AttributeValue{"pk": key["pk"], "v": &types.AttributeValueMemberN{Value: "0"}}
		if _, e = c.PutItem(ctx, &sdk.PutItemInput{TableName: &name, Item: item}); e != nil {
			t.Fatal(e)
		}
		var wins atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, er := c.UpdateItem(ctx, &sdk.UpdateItemInput{TableName: &name, Key: key, UpdateExpression: aws.String("SET #v = #v + :one"), ConditionExpression: aws.String("#v = :zero"), ExpressionAttributeNames: map[string]string{"#v": "v"}, ExpressionAttributeValues: map[string]types.AttributeValue{":one": &types.AttributeValueMemberN{Value: "1"}, ":zero": &types.AttributeValueMemberN{Value: "0"}}})
				if er == nil {
					wins.Add(1)
					return
				}
				var conditional *types.ConditionalCheckFailedException
				if !errors.As(er, &conditional) {
					t.Errorf("unexpected error: %v", er)
				}
			}()
		}
		wg.Wait()
		if wins.Load() != 1 {
			t.Fatalf("wins=%d", wins.Load())
		}
	})
}
