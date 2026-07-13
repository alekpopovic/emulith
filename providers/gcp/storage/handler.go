package storage

import (
	"encoding/json"
	"fmt"
	"github.com/alekpopovic/emulith/internal/state"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Handler struct {
	Store   *state.Store
	Project string
}

var bucketRE = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,61}[a-z0-9]$`)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("x-goog-request-id", "local-request")
	if r.URL.Path != "/storage/v1/b" && !strings.HasPrefix(r.URL.Path, "/storage/v1/b/") {
		errJSON(w, 404, "notFound", "Not found")
		return
	}
	q := r.URL.Query()
	if r.URL.Path == "/storage/v1/b" {
		if r.Method == "GET" {
			if q.Get("project") != "" && q.Get("project") != h.Project {
				errJSON(w, 400, "invalid", "cross-project")
				return
			}
			bs, _ := h.Store.ListGCPBuckets(r.Context(), h.Project)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"kind": "storage#buckets", "items": bs})
			return
		}
		if r.Method != "POST" {
			errJSON(w, 405, "methodNotAllowed", "unsupported")
			return
		}
		var in struct {
			Name, Location, StorageClass string
			Labels                       map[string]string
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || !bucketRE.MatchString(in.Name) {
			errJSON(w, 400, "invalid", "invalid bucket")
			return
		}
		now := time.Now().UTC()
		b := state.GCPBucket{Project: h.Project, Name: in.Name, Location: in.Location, StorageClass: in.StorageClass, Labels: in.Labels, ETag: "\"local\"", CreatedAt: now, UpdatedAt: now, Metageneration: 1}
		if b.Location == "" {
			b.Location = "US"
		}
		if b.StorageClass == "" {
			b.StorageClass = "STANDARD"
		}
		if e := h.Store.CreateGCPBucket(r.Context(), b); e == state.ErrConflict {
			errJSON(w, 409, "conflict", "bucket exists")
			return
		}
		json.NewEncoder(w).Encode(b)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/storage/v1/b/")
	if !bucketRE.MatchString(name) {
		errJSON(w, 400, "invalid", "invalid bucket")
		return
	}
	b, e := h.Store.GetGCPBucket(r.Context(), h.Project, name)
	if e != nil {
		errJSON(w, 404, "notFound", "bucket not found")
		return
	}
	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(b)
	case "PATCH":
		var in struct{ Labels map[string]string }
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			errJSON(w, 400, "invalid", "invalid json")
			return
		}
		b.Labels = in.Labels
		b.Metageneration++
		b.UpdatedAt = time.Now().UTC()
		h.Store.UpdateGCPBucket(r.Context(), b)
		json.NewEncoder(w).Encode(b)
	case "DELETE":
		if e := h.Store.DeleteGCPBucket(r.Context(), h.Project, name); e != nil {
			errJSON(w, 412, "conditionNotMet", "bucket not empty")
			return
		}
		w.WriteHeader(204)
	default:
		errJSON(w, 405, "methodNotAllowed", "unsupported")
	}
}
func errJSON(w http.ResponseWriter, code int, reason, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": code, "message": msg, "errors": []map[string]string{{"reason": reason, "message": msg}}}})
}

var _ = fmt.Sprint
