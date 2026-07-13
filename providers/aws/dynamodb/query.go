package dynamodb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"net/http"
	"sort"
)

const maxEvaluatedItems = 10000

type queryInput struct {
	TableName                 string
	IndexName                 string
	KeyConditionExpression    string
	FilterExpression          string
	ProjectionExpression      string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]Value
	ExclusiveStartKey         map[string]Value
	Limit                     *int
	ScanIndexForward          *bool
	Select                    string
	ConsistentRead            bool
	KeyConditions             json.RawMessage
	ScanFilter                json.RawMessage
	Segment                   *int
	TotalSegments             *int
	ReturnConsumedCapacity    string
}

func (h *Handler) query(w http.ResponseWriter, r *awsprovider.Request) {
	var in queryInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.IndexName != "" || len(in.KeyConditions) > 0 {
		h.err(w, 400, "ValidationException", "secondary indexes and legacy KeyConditions are unsupported")
		return
	}
	t, e := h.store.GetDynamoTable(r.HTTPRequest.Context(), in.TableName)
	if e != nil {
		h.stateErr(w, e)
		return
	}
	keyExpr, e := parseCondition(in.KeyConditionExpression, in.ExpressionAttributeNames, in.ExpressionAttributeValues)
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if e = validateKeyCondition(keyExpr, t); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	filter, e := optionalCondition(in.FilterExpression, in.ExpressionAttributeNames, in.ExpressionAttributeValues)
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	h.readItems(w, r, in, t, keyExpr, filter, true)
}
func (h *Handler) scan(w http.ResponseWriter, r *awsprovider.Request) {
	var in queryInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.IndexName != "" || len(in.ScanFilter) > 0 || in.Segment != nil || in.TotalSegments != nil {
		h.err(w, 400, "ValidationException", "indexes legacy filters and parallel scan are unsupported")
		return
	}
	t, e := h.store.GetDynamoTable(r.HTTPRequest.Context(), in.TableName)
	if e != nil {
		h.stateErr(w, e)
		return
	}
	filter, e := optionalCondition(in.FilterExpression, in.ExpressionAttributeNames, in.ExpressionAttributeValues)
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	h.readItems(w, r, in, t, nil, filter, false)
}
func optionalCondition(s string, n map[string]string, v map[string]Value) (*expression, error) {
	if s == "" {
		return nil, nil
	}
	return parseCondition(s, n, v)
}
func validateKeyCondition(e *expression, t state.DynamoTable) error {
	if e == nil {
		return errors.New("KeyConditionExpression is required")
	}
	seenPK := false
	var walk func(*expression) error
	walk = func(x *expression) error {
		if x == nil {
			return nil
		}
		if x.op == "AND" {
			if e := walk(x.left); e != nil {
				return e
			}
			return walk(x.right)
		}
		if x.op == "=" && x.left != nil && x.left.op == "PATH" && len(x.left.path) == 1 && x.left.path[0].name == t.PartitionKey && x.right.op == "VALUE" {
			if seenPK {
				return errors.New("duplicate partition condition")
			}
			seenPK = true
			return nil
		}
		allowed := map[string]bool{"=": true, "<": true, "<=": true, ">": true, ">=": true, "BETWEEN": true, "begins_with": true}
		path := x.left
		if x.op == "begins_with" && len(x.args) > 0 {
			path = x.args[0]
		}
		if t.SortKey != "" && allowed[x.op] && path != nil && path.op == "PATH" && len(path.path) == 1 && path.path[0].name == t.SortKey {
			return nil
		}
		return errors.New("invalid key condition")
	}
	if e := walk(e); e != nil {
		return e
	}
	if !seenPK {
		return errors.New("partition key equality is required")
	}
	return nil
}
func (h *Handler) readItems(w http.ResponseWriter, r *awsprovider.Request, in queryInput, t state.DynamoTable, keyExpr, filter *expression, isQuery bool) {
	limit := 100
	if in.Limit != nil {
		limit = *in.Limit
		if limit < 1 || limit > 1000 {
			h.err(w, 400, "ValidationException", "Limit must be between 1 and 1000")
			return
		}
	}
	if in.Select != "" && in.Select != "ALL_ATTRIBUTES" && in.Select != "COUNT" && in.Select != "SPECIFIC_ATTRIBUTES" {
		h.err(w, 400, "ValidationException", "unsupported Select")
		return
	}
	if in.Select == "SPECIFIC_ATTRIBUTES" && in.ProjectionExpression == "" {
		h.err(w, 400, "ValidationException", "SPECIFIC_ATTRIBUTES requires ProjectionExpression")
		return
	}
	if in.Select == "COUNT" && in.ProjectionExpression != "" {
		h.err(w, 400, "ValidationException", "COUNT cannot use ProjectionExpression")
		return
	}
	paths, e := parseProjection(in.ProjectionExpression, in.ExpressionAttributeNames)
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	records, e := h.store.ListAllDynamoItems(r.HTTPRequest.Context(), in.TableName, maxEvaluatedItems)
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	type decoded struct {
		rec  state.DynamoItem
		item map[string]Value
	}
	var all []decoded
	for _, rec := range records {
		var item map[string]Value
		if json.Unmarshal(rec.Payload, &item) != nil {
			continue
		}
		if keyExpr != nil {
			ok, er := evalCondition(keyExpr, item)
			if er != nil {
				h.err(w, 400, "ValidationException", er.Error())
				return
			}
			if !ok {
				continue
			}
		}
		all = append(all, decoded{rec, item})
	}
	sort.Slice(all, func(i, j int) bool { return compareItems(all[i].item, all[j].item, t) < 0 })
	forward := in.ScanIndexForward == nil || *in.ScanIndexForward
	if isQuery && !forward {
		for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
			all[i], all[j] = all[j], all[i]
		}
	}
	start := 0
	if in.ExclusiveStartKey != nil {
		_, startKey, _, _, er := h.tableAndKey(r, in.TableName, in.ExclusiveStartKey)
		if er != nil {
			h.stateOrValidation(w, er)
			return
		}
		found := false
		for i, x := range all {
			if bytes.Equal(x.rec.PrimaryKey, startKey) {
				start = i + 1
				found = true
				break
			}
		}
		if !found {
			for start < len(all) && compareItems(all[start].item, in.ExclusiveStartKey, t) <= 0 {
				start++
			}
		}
	}
	evaluated := 0
	items := []map[string]Value{}
	var last map[string]Value
	for i := start; i < len(all) && evaluated < limit; i++ {
		evaluated++
		last = keyMap(all[i].item, t)
		pass := true
		if filter != nil {
			pass, e = evalCondition(filter, all[i].item)
			if e != nil {
				h.err(w, 400, "ValidationException", e.Error())
				return
			}
		}
		if pass && in.Select != "COUNT" {
			item := all[i].item
			if len(paths) > 0 {
				item = projectItem(item, paths)
			}
			items = append(items, item)
		}
	}
	out := map[string]any{"Count": len(items), "ScannedCount": evaluated}
	if in.Select == "COUNT" {
		count := 0
		for i := start; i < start+evaluated; i++ {
			pass := true
			if filter != nil {
				pass, _ = evalCondition(filter, all[i].item)
			}
			if pass {
				count++
			}
		}
		out["Count"] = count
	} else {
		out["Items"] = items
	}
	if start+evaluated < len(all) && last != nil {
		out["LastEvaluatedKey"] = last
	}
	json.NewEncoder(w).Encode(out)
}
func compareItems(a, b map[string]Value, t state.DynamoTable) int {
	if c, _ := compareValue(a[t.PartitionKey], b[t.PartitionKey]); c != 0 {
		return c
	}
	if t.SortKey != "" {
		c, _ := compareValue(a[t.SortKey], b[t.SortKey])
		return c
	}
	return 0
}
func keyMap(item map[string]Value, t state.DynamoTable) map[string]Value {
	m := map[string]Value{t.PartitionKey: item[t.PartitionKey]}
	if t.SortKey != "" {
		m[t.SortKey] = item[t.SortKey]
	}
	return m
}
func parseProjection(s string, names map[string]string) ([][]pathPart, error) {
	if s == "" {
		return nil, nil
	}
	tokens, e := lexExpression(s)
	if e != nil {
		return nil, e
	}
	p := expressionParser{tokens: tokens, names: names}
	var out [][]pathPart
	for {
		path, e := p.parsePath()
		if e != nil {
			return nil, e
		}
		for _, x := range out {
			if pathsOverlap(x, path) {
				return nil, errors.New("projection paths overlap")
			}
		}
		out = append(out, path)
		if p.peek().kind == tEOF {
			break
		}
		if p.peek().kind != tComma {
			return nil, fmt.Errorf("expected comma at position %d", p.peek().pos)
		}
		p.take()
	}
	return out, nil
}
func projectItem(item map[string]Value, paths [][]pathPart) map[string]Value {
	out := map[string]Value{}
	for _, p := range paths {
		v, ok, _ := getPath(item, p)
		if !ok {
			continue
		}
		putProjected(out, p, v)
	}
	return out
}
func putProjected(out map[string]Value, path []pathPart, v Value) {
	if len(path) == 1 {
		out[path[0].name] = v
		return
	}
	root := out[path[0].name]
	if root.Kind == "" {
		root = Value{Kind: "M", M: map[string]Value{}}
	}
	if !path[1].isIdx && root.Kind == "M" {
		root.M[path[1].name] = v
	}
	out[path[0].name] = root
}
