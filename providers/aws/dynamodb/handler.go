package dynamodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"github.com/google/uuid"
	"io"
	"net/http"
	"regexp"
	"time"
)

type Handler struct {
	store  *state.Store
	region string
}

func New(store *state.Store) *Handler { return &Handler{store: store, region: "us-east-1"} }

type attrDef struct {
	AttributeName string
	AttributeType string
}
type keyDef struct {
	AttributeName string
	KeyType       string
}
type createInput struct {
	TableName              string
	AttributeDefinitions   []attrDef
	KeySchema              []keyDef
	BillingMode            string
	ProvisionedThroughput  json.RawMessage
	GlobalSecondaryIndexes json.RawMessage
	LocalSecondaryIndexes  json.RawMessage
	StreamSpecification    json.RawMessage
	SSESpecification       json.RawMessage
	TableClass             string
	Tags                   json.RawMessage
}

var tableNameRE = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,255}$`)

func (h *Handler) ServeAWS(w http.ResponseWriter, req *awsprovider.Request, id string) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", id)
	switch req.Operation {
	case "CreateTable":
		h.create(w, req)
	case "DescribeTable":
		h.describe(w, req)
	case "ListTables":
		h.list(w, req)
	case "DeleteTable":
		h.delete(w, req)
	case "PutItem":
		h.putItem(w, req)
	case "GetItem":
		h.getItem(w, req)
	case "DeleteItem":
		h.deleteItem(w, req)
	case "UpdateItem":
		h.updateItem(w, req)
	case "Query":
		h.query(w, req)
	case "Scan":
		h.scan(w, req)
	default:
		h.err(w, 400, "UnknownOperationException", "Operation "+req.Operation+" is not implemented")
	}
}
func decode(req *awsprovider.Request, v any) error {
	d := json.NewDecoder(io.LimitReader(req.HTTPRequest.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func (h *Handler) create(w http.ResponseWriter, r *awsprovider.Request) {
	var in createInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	schema, e := validateCreate(in)
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	id := uuid.NewString()
	t := state.DynamoTable{Name: in.TableName, TableID: id, ARN: fmt.Sprintf("arn:aws:dynamodb:%s:000000000000:table/%s", h.region, in.TableName), Status: "ACTIVE", BillingMode: "PAY_PER_REQUEST", CreatedAt: time.Now().UTC(), PartitionKey: schema.PartitionName, PartitionType: schema.PartitionType, SortKey: schema.SortName, SortType: schema.SortType}
	e = h.store.CreateDynamoTable(r.HTTPRequest.Context(), t)
	if errors.Is(e, state.ErrDynamoExists) {
		h.err(w, 400, "ResourceInUseException", "Table already exists: "+in.TableName)
		return
	}
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"TableDescription": description(t)})
}
func validateCreate(in createInput) (KeySchema, error) {
	var s KeySchema
	if !tableNameRE.MatchString(in.TableName) {
		return s, errors.New("invalid TableName")
	}
	if in.BillingMode != "PAY_PER_REQUEST" {
		return s, errors.New("only PAY_PER_REQUEST is supported")
	}
	if len(in.ProvisionedThroughput) > 0 || len(in.GlobalSecondaryIndexes) > 0 || len(in.LocalSecondaryIndexes) > 0 || len(in.StreamSpecification) > 0 || len(in.SSESpecification) > 0 || in.TableClass != "" || len(in.Tags) > 0 {
		return s, errors.New("unsupported table option")
	}
	defs := map[string]string{}
	for _, d := range in.AttributeDefinitions {
		if defs[d.AttributeName] != "" || (d.AttributeType != "S" && d.AttributeType != "N" && d.AttributeType != "B") {
			return s, errors.New("invalid attribute definitions")
		}
		defs[d.AttributeName] = d.AttributeType
	}
	for _, k := range in.KeySchema {
		if defs[k.AttributeName] == "" {
			return s, errors.New("key lacks attribute definition")
		}
		switch k.KeyType {
		case "HASH":
			if s.PartitionName != "" {
				return s, errors.New("exactly one HASH key required")
			}
			s.PartitionName, s.PartitionType = k.AttributeName, defs[k.AttributeName]
		case "RANGE":
			if s.SortName != "" {
				return s, errors.New("at most one RANGE key")
			}
			s.SortName, s.SortType = k.AttributeName, defs[k.AttributeName]
		default:
			return s, errors.New("invalid key type")
		}
	}
	if s.PartitionName == "" || len(defs) != len(in.KeySchema) {
		return s, errors.New("unused or missing attribute definition")
	}
	return s, nil
}
func (h *Handler) describe(w http.ResponseWriter, r *awsprovider.Request) {
	var in struct{ TableName string }
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	t, e := h.store.GetDynamoTable(r.HTTPRequest.Context(), in.TableName)
	if errors.Is(e, state.ErrDynamoNotFound) {
		h.err(w, 400, "ResourceNotFoundException", "Requested resource not found")
		return
	}
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"Table": description(t)})
}
func (h *Handler) list(w http.ResponseWriter, r *awsprovider.Request) {
	var in struct {
		Limit                   int
		ExclusiveStartTableName string
	}
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.Limit == 0 {
		in.Limit = 100
	}
	if in.Limit < 1 || in.Limit > 100 {
		h.err(w, 400, "ValidationException", "Limit must be between 1 and 100")
		return
	}
	ts, e := h.store.ListDynamoTables(r.HTTPRequest.Context(), in.ExclusiveStartTableName, in.Limit+1)
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	names := []string{}
	out := map[string]any{"TableNames": names}
	if len(ts) > in.Limit {
		out["LastEvaluatedTableName"] = ts[in.Limit-1].Name
		ts = ts[:in.Limit]
	}
	names = []string{}
	for _, t := range ts {
		names = append(names, t.Name)
	}
	out["TableNames"] = names
	json.NewEncoder(w).Encode(out)
}
func (h *Handler) delete(w http.ResponseWriter, r *awsprovider.Request) {
	var in struct{ TableName string }
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	t, e := h.store.DeleteDynamoTable(r.HTTPRequest.Context(), in.TableName)
	if errors.Is(e, state.ErrDynamoNotFound) {
		h.err(w, 400, "ResourceNotFoundException", "Requested resource not found")
		return
	}
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	d := description(t)
	d["TableStatus"] = "DELETING"
	json.NewEncoder(w).Encode(map[string]any{"TableDescription": d})
}
func description(t state.DynamoTable) map[string]any {
	ks := []map[string]string{{"AttributeName": t.PartitionKey, "KeyType": "HASH"}}
	ad := []map[string]string{{"AttributeName": t.PartitionKey, "AttributeType": t.PartitionType}}
	if t.SortKey != "" {
		ks = append(ks, map[string]string{"AttributeName": t.SortKey, "KeyType": "RANGE"})
		ad = append(ad, map[string]string{"AttributeName": t.SortKey, "AttributeType": t.SortType})
	}
	return map[string]any{"TableName": t.Name, "TableStatus": t.Status, "CreationDateTime": float64(t.CreatedAt.UnixNano()) / 1e9, "KeySchema": ks, "AttributeDefinitions": ad, "BillingModeSummary": map[string]string{"BillingMode": "PAY_PER_REQUEST"}, "TableArn": t.ARN, "TableId": t.TableID, "ItemCount": 0, "TableSizeBytes": 0}
}
func (h *Handler) err(w http.ResponseWriter, status int, code, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"__type": "com.amazonaws.dynamodb.v20120810#" + code, "message": msg})
}
