package dynamodb

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
)

type batchGetTable struct {
	Keys                     []map[string]Value
	ConsistentRead           bool
	ProjectionExpression     string
	ExpressionAttributeNames map[string]string
}
type batchGetInput struct {
	RequestItems           map[string]batchGetTable
	ReturnConsumedCapacity string
}

func (h *Handler) batchGet(w http.ResponseWriter, r *awsprovider.Request) {
	var in batchGetInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.ReturnConsumedCapacity != "" && in.ReturnConsumedCapacity != "NONE" {
		h.err(w, 400, "ValidationException", "consumed capacity is unsupported")
		return
	}
	total := 0
	tables := make([]string, 0, len(in.RequestItems))
	for name, x := range in.RequestItems {
		total += len(x.Keys)
		tables = append(tables, name)
	}
	if total < 1 || total > 100 {
		h.err(w, 400, "ValidationException", "BatchGetItem requires 1 to 100 keys")
		return
	}
	sort.Strings(tables)
	type prepared struct {
		name  string
		keys  [][]byte
		paths [][]pathPart
	}
	var all []prepared
	seen := map[string]bool{}
	for _, name := range tables {
		x := in.RequestItems[name]
		paths, e := parseProjection(x.ProjectionExpression, x.ExpressionAttributeNames)
		if e != nil {
			h.err(w, 400, "ValidationException", e.Error())
			return
		}
		p := prepared{name: name, paths: paths}
		for _, key := range x.Keys {
			_, encoded, _, _, e := h.tableAndKey(r, name, key)
			if e != nil {
				h.stateOrValidation(w, e)
				return
			}
			id := name + "\x00" + string(encoded)
			if seen[id] {
				h.err(w, 400, "ValidationException", "duplicate batch key")
				return
			}
			seen[id] = true
			p.keys = append(p.keys, encoded)
		}
		all = append(all, p)
	}
	responses := map[string][]map[string]Value{}
	for _, p := range all {
		responses[p.name] = []map[string]Value{}
		for _, key := range p.keys {
			b, e := h.store.GetDynamoItem(r.HTTPRequest.Context(), p.name, key)
			if e != nil {
				h.err(w, 500, "InternalServerError", e.Error())
				return
			}
			if b == nil {
				continue
			}
			var item map[string]Value
			if e = json.Unmarshal(b, &item); e != nil {
				h.err(w, 500, "InternalServerError", "corrupt item")
				return
			}
			if len(p.paths) > 0 {
				item = projectItem(item, p.paths)
			}
			responses[p.name] = append(responses[p.name], item)
		}
	}
	json.NewEncoder(w).Encode(map[string]any{"Responses": responses, "UnprocessedKeys": map[string]any{}})
}

type putRequest struct{ Item map[string]Value }
type deleteRequest struct{ Key map[string]Value }
type batchWriteEntry struct {
	PutRequest    *putRequest
	DeleteRequest *deleteRequest
}
type batchWriteInput struct {
	RequestItems                map[string][]batchWriteEntry
	ReturnConsumedCapacity      string
	ReturnItemCollectionMetrics string
}

func (h *Handler) batchWrite(w http.ResponseWriter, r *awsprovider.Request) {
	var in batchWriteInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if (in.ReturnConsumedCapacity != "" && in.ReturnConsumedCapacity != "NONE") || (in.ReturnItemCollectionMetrics != "" && in.ReturnItemCollectionMetrics != "NONE") {
		h.err(w, 400, "ValidationException", "capacity and metrics are unsupported")
		return
	}
	total := 0
	for _, x := range in.RequestItems {
		total += len(x)
	}
	if total < 1 || total > 25 {
		h.err(w, 400, "ValidationException", "BatchWriteItem requires 1 to 25 writes")
		return
	}
	tables := make([]string, 0, len(in.RequestItems))
	for name := range in.RequestItems {
		tables = append(tables, name)
	}
	sort.Strings(tables)
	var ops []state.DynamoWrite
	seen := map[string]bool{}
	for _, name := range tables {
		table, e := h.store.GetDynamoTable(r.HTTPRequest.Context(), name)
		if e != nil {
			h.stateErr(w, e)
			return
		}
		for _, entry := range in.RequestItems[name] {
			if (entry.PutRequest == nil) == (entry.DeleteRequest == nil) {
				h.err(w, 400, "ValidationException", "write entry must contain exactly one action")
				return
			}
			if entry.PutRequest != nil {
				item := entry.PutRequest.Item
				keys := map[string]Value{table.PartitionKey: item[table.PartitionKey]}
				if table.SortKey != "" {
					keys[table.SortKey] = item[table.SortKey]
				}
				_, key, p, s, e := h.tableAndKey(r, name, keys)
				if e != nil {
					h.err(w, 400, "ValidationException", e.Error())
					return
				}
				payload, e := json.Marshal(item)
				if e != nil || len(payload) > MaxItemSize {
					h.err(w, 400, "ValidationException", "invalid or oversized item")
					return
				}
				id := name + "\x00" + string(key)
				if seen[id] {
					h.err(w, 400, "ValidationException", "duplicate batch action")
					return
				}
				seen[id] = true
				ops = append(ops, state.DynamoWrite{Table: name, Key: key, Partition: p, SortKey: s, Payload: payload})
			} else {
				_, key, p, s, e := h.tableAndKey(r, name, entry.DeleteRequest.Key)
				if e != nil {
					h.stateOrValidation(w, e)
					return
				}
				id := name + "\x00" + string(key)
				if seen[id] {
					h.err(w, 400, "ValidationException", "duplicate batch action")
					return
				}
				seen[id] = true
				ops = append(ops, state.DynamoWrite{Table: name, Key: key, Partition: p, SortKey: s, Delete: true})
			}
		}
	}
	if e := h.store.BatchWriteDynamoItems(r.HTTPRequest.Context(), ops); e != nil {
		if errors.Is(e, state.ErrDynamoNotFound) {
			h.stateErr(w, e)
		} else {
			h.err(w, 500, "InternalServerError", e.Error())
		}
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"UnprocessedItems": map[string]any{}})
}
