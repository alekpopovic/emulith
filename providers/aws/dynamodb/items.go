package dynamodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"net/http"
	"strings"
	"unicode"
)

type itemInput struct {
	TableName                 string
	Item                      map[string]Value
	Key                       map[string]Value
	ReturnValues              string
	ConsistentRead            bool
	UpdateExpression          string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]Value
	ConditionExpression       string
	Expected                  json.RawMessage
	ProjectionExpression      string
	AttributesToGet           json.RawMessage
}

func (h *Handler) tableAndKey(r *awsprovider.Request, name string, item map[string]Value) (state.DynamoTable, []byte, []byte, []byte, error) {
	t, e := h.store.GetDynamoTable(r.HTTPRequest.Context(), name)
	if e != nil {
		return t, nil, nil, nil, e
	}
	schema := KeySchema{t.PartitionKey, t.PartitionType, t.SortKey, t.SortType}
	key, e := EncodeKey(schema, item)
	if e != nil {
		return t, nil, nil, nil, e
	}
	allowed := 1
	if t.SortKey != "" {
		allowed = 2
	}
	if len(item) != allowed {
		return t, nil, nil, nil, errors.New("key contains extra attributes")
	}
	p, _ := item[t.PartitionKey].Canonical()
	var s []byte
	if t.SortKey != "" {
		s, _ = item[t.SortKey].Canonical()
	}
	return t, key, p, s, nil
}
func (h *Handler) putItem(w http.ResponseWriter, r *awsprovider.Request) {
	var in itemInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.ConditionExpression != "" || len(in.Expected) > 0 || (in.ReturnValues != "" && in.ReturnValues != "NONE" && in.ReturnValues != "ALL_OLD") {
		h.err(w, 400, "ValidationException", "unsupported PutItem option")
		return
	}
	t, e := h.store.GetDynamoTable(r.HTTPRequest.Context(), in.TableName)
	if e != nil {
		h.stateErr(w, e)
		return
	}
	keys := map[string]Value{t.PartitionKey: in.Item[t.PartitionKey]}
	if t.SortKey != "" {
		keys[t.SortKey] = in.Item[t.SortKey]
	}
	_, key, p, s, e := h.tableAndKey(r, in.TableName, keys)
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	payload, e := json.Marshal(in.Item)
	if e != nil || len(payload) > MaxItemSize {
		h.err(w, 400, "ValidationException", "item is invalid or too large")
		return
	}
	old, e := h.store.PutDynamoItem(r.HTTPRequest.Context(), in.TableName, key, p, s, payload)
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	out := map[string]any{}
	if in.ReturnValues == "ALL_OLD" && old != nil {
		var x map[string]Value
		json.Unmarshal(old, &x)
		out["Attributes"] = x
	}
	json.NewEncoder(w).Encode(out)
}
func (h *Handler) getItem(w http.ResponseWriter, r *awsprovider.Request) {
	var in itemInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.ProjectionExpression != "" || len(in.AttributesToGet) > 0 {
		h.err(w, 400, "ValidationException", "projection is not supported")
		return
	}
	_, key, _, _, e := h.tableAndKey(r, in.TableName, in.Key)
	if e != nil {
		h.stateOrValidation(w, e)
		return
	}
	b, e := h.store.GetDynamoItem(r.HTTPRequest.Context(), in.TableName, key)
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	out := map[string]any{}
	if b != nil {
		var x map[string]Value
		json.Unmarshal(b, &x)
		out["Item"] = x
	}
	json.NewEncoder(w).Encode(out)
}
func (h *Handler) deleteItem(w http.ResponseWriter, r *awsprovider.Request) {
	var in itemInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.ConditionExpression != "" || len(in.Expected) > 0 || (in.ReturnValues != "" && in.ReturnValues != "NONE" && in.ReturnValues != "ALL_OLD") {
		h.err(w, 400, "ValidationException", "unsupported DeleteItem option")
		return
	}
	_, key, _, _, e := h.tableAndKey(r, in.TableName, in.Key)
	if e != nil {
		h.stateOrValidation(w, e)
		return
	}
	old, e := h.store.DeleteDynamoItem(r.HTTPRequest.Context(), in.TableName, key)
	if e != nil {
		h.err(w, 500, "InternalServerError", e.Error())
		return
	}
	out := map[string]any{}
	if in.ReturnValues == "ALL_OLD" && old != nil {
		var x map[string]Value
		json.Unmarshal(old, &x)
		out["Attributes"] = x
	}
	json.NewEncoder(w).Encode(out)
}

type updateAction struct{ kind, name, value string }

func parseUpdate(s string) ([]updateAction, error) {
	tokens := lex(s)
	if len(tokens) < 2 {
		return nil, errors.New("invalid update expression")
	}
	var out []updateAction
	kind := ""
	for i := 0; i < len(tokens); {
		u := strings.ToUpper(tokens[i])
		if u == "SET" || u == "REMOVE" {
			kind = u
			i++
			continue
		}
		if kind == "" {
			return nil, errors.New("missing update action")
		}
		a := updateAction{kind: kind, name: tokens[i]}
		i++
		if kind == "SET" {
			if i+1 >= len(tokens) || tokens[i] != "=" {
				return nil, errors.New("invalid SET")
			}
			a.value = tokens[i+1]
			i += 2
		}
		out = append(out, a)
		if i < len(tokens) && tokens[i] == "," {
			i++
		}
	}
	return out, nil
}
func lex(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		if unicode.IsSpace(rune(s[i])) {
			i++
			continue
		}
		if strings.ContainsRune(",=", rune(s[i])) {
			out = append(out, s[i:i+1])
			i++
			continue
		}
		j := i
		for j < len(s) && !unicode.IsSpace(rune(s[j])) && !strings.ContainsRune(",=", rune(s[j])) {
			j++
		}
		out = append(out, s[i:j])
		i = j
	}
	return out
}
func (h *Handler) updateItem(w http.ResponseWriter, r *awsprovider.Request) {
	var in itemInput
	if e := decode(r, &in); e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	if in.ConditionExpression != "" {
		h.err(w, 400, "ValidationException", "conditions unsupported")
		return
	}
	actions, e := parseUpdate(in.UpdateExpression)
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	t, key, p, s, e := h.tableAndKey(r, in.TableName, in.Key)
	if e != nil {
		h.stateOrValidation(w, e)
		return
	}
	old, next, e := h.store.UpdateDynamoItem(r.HTTPRequest.Context(), in.TableName, key, p, s, func(b []byte) ([]byte, error) {
		item := map[string]Value{}
		if b != nil {
			if e := json.Unmarshal(b, &item); e != nil {
				return nil, e
			}
		} else {
			for k, v := range in.Key {
				item[k] = v
			}
		}
		for _, a := range actions {
			name := in.ExpressionAttributeNames[a.name]
			if name == "" {
				return nil, fmt.Errorf("missing name mapping %s", a.name)
			}
			if name == t.PartitionKey || name == t.SortKey {
				return nil, errors.New("cannot update primary key")
			}
			if a.kind == "SET" {
				v, ok := in.ExpressionAttributeValues[a.value]
				if !ok {
					return nil, fmt.Errorf("missing value %s", a.value)
				}
				item[name] = v
			} else {
				delete(item, name)
			}
		}
		return json.Marshal(item)
	})
	if e != nil {
		h.err(w, 400, "ValidationException", e.Error())
		return
	}
	var before, after map[string]Value
	if old != nil {
		json.Unmarshal(old, &before)
	}
	json.Unmarshal(next, &after)
	out := map[string]any{}
	switch in.ReturnValues {
	case "", "NONE":
	case "ALL_OLD":
		if old != nil {
			out["Attributes"] = before
		}
	case "ALL_NEW":
		out["Attributes"] = after
	case "UPDATED_OLD", "UPDATED_NEW":
		picked := map[string]Value{}
		src := after
		if in.ReturnValues == "UPDATED_OLD" {
			src = before
		}
		for _, a := range actions {
			n := in.ExpressionAttributeNames[a.name]
			if v, ok := src[n]; ok {
				picked[n] = v
			}
		}
		out["Attributes"] = picked
	default:
		h.err(w, 400, "ValidationException", "unsupported ReturnValues")
		return
	}
	json.NewEncoder(w).Encode(out)
}
func (h *Handler) stateErr(w http.ResponseWriter, e error) {
	if errors.Is(e, state.ErrDynamoNotFound) {
		h.err(w, 400, "ResourceNotFoundException", "Requested resource not found")
	} else {
		h.err(w, 500, "InternalServerError", e.Error())
	}
}
func (h *Handler) stateOrValidation(w http.ResponseWriter, e error) {
	if errors.Is(e, state.ErrDynamoNotFound) {
		h.stateErr(w, e)
	} else {
		h.err(w, 400, "ValidationException", e.Error())
	}
}
