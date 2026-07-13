package logs

import (
	"encoding/json"
	"errors"
	"github.com/alekpopovic/emulith/internal/state"
	aws "github.com/alekpopovic/emulith/providers/aws"
	"net/http"
	"time"
)

type Handler struct{ Store *state.Store }

func New(s *state.Store) *Handler { return &Handler{Store: s} }
func (h *Handler) ServeAWS(w http.ResponseWriter, r *aws.Request, id string) {
	var b map[string]any
	dec := json.NewDecoder(r.HTTPRequest.Body)
	if dec.Decode(&b) != nil {
		h.err(w, id, 400, "InvalidParameterException")
		return
	}
	op := r.Operation
	var out any
	var e error
	switch op {
	case "CreateLogGroup":
		e = h.Store.LogCreateGroup(r.HTTPRequest.Context(), str(b["logGroupName"]))
	case "DeleteLogGroup":
		e = h.Store.LogDeleteGroup(r.HTTPRequest.Context(), str(b["logGroupName"]))
	case "DescribeLogGroups":
		var a []string
		a, e = h.Store.LogGroups(r.HTTPRequest.Context(), str(b["logGroupNamePrefix"]), lim(b["limit"]))
		out = map[string]any{"logGroups": groups(a)}
	case "CreateLogStream":
		e = h.Store.LogCreateStream(r.HTTPRequest.Context(), str(b["logGroupName"]), str(b["logStreamName"]))
	case "DescribeLogStreams":
		a, ee := h.Store.LogStreams(r.HTTPRequest.Context(), str(b["logGroupName"]))
		e = ee
		out = map[string]any{"logStreams": streams(a)}
	case "PutLogEvents":
		g, st := str(b["logGroupName"]), str(b["logStreamName"])
		var ev []state.LogEvent
		if x, ok := b["logEvents"].([]any); ok {
			for _, z := range x {
				m := z.(map[string]any)
				ev = append(ev, state.LogEvent{Timestamp: int64(m["timestamp"].(float64)), Message: str(m["message"]), Ingested: time.Now().UnixMilli()})
			}
		}
		e = h.Store.LogPut(r.HTTPRequest.Context(), g, st, ev)
		out = map[string]any{"nextSequenceToken": "local"}
	case "GetLogEvents":
		a, ee := h.Store.LogEvents(r.HTTPRequest.Context(), str(b["logGroupName"]), str(b["logStreamName"]), 0, 1<<62, lim(b["limit"]))
		e = ee
		var x []any
		for _, v := range a {
			x = append(x, map[string]any{"timestamp": v.Timestamp, "message": v.Message, "ingestionTime": v.Ingested})
		}
		out = map[string]any{"events": x}
	default:
		e = errors.New("unsupported")
	}
	if e != nil {
		h.err(w, id, 400, "ResourceNotFoundException")
		return
	}
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.Header().Set("x-amzn-RequestId", id)
	json.NewEncoder(w).Encode(out)
}
func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func lim(v any) int {
	if n, ok := v.(float64); ok && n > 0 && n < 10000 {
		return int(n)
	}
	return 1000
}
func groups(a []string) []any {
	var x []any
	for _, n := range a {
		x = append(x, map[string]any{"logGroupName": n})
	}
	return x
}
func streams(a []string) []any {
	var x []any
	for _, n := range a {
		x = append(x, map[string]any{"logStreamName": n})
	}
	return x
}
func (h *Handler) err(w http.ResponseWriter, id string, s int, c string) {
	w.Header().Set("x-amzn-RequestId", id)
	w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]any{"__type": c, "message": c})
}
