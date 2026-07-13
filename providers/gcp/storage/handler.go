package storage

import (
	"encoding/json"
	"fmt"
	"github.com/alekpopovic/emulith/internal/state"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	if strings.HasPrefix(r.URL.Path, "/upload/storage/v1/b/") && r.Method == "POST" {
		h.upload(w, r)
		return
	}
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
	if strings.Contains(name, "/o/") {
		p := strings.SplitN(name, "/o/", 2)
		if len(p) != 2 {
			errJSON(w, 400, "invalid", "invalid object")
			return
		}
		o, e := h.Store.GetGCPObject(r.Context(), h.Project, p[0], p[1])
		if e != nil {
			errJSON(w, 404, "notFound", "object not found")
			return
		}
		if r.Method == "DELETE" {
			h.Store.DeleteGCPObject(r.Context(), h.Project, p[0], p[1])
			os.Remove(o.BodyPath)
			w.WriteHeader(204)
			return
		}
		if r.Method == "GET" && q.Get("alt") == "media" {
			f, e := os.Open(o.BodyPath)
			if e != nil {
				errJSON(w, 404, "notFound", "object body missing")
				return
			}
			defer f.Close()
			w.Header().Set("Content-Type", o.ContentType)
			io.Copy(w, f)
			return
		}
		json.NewEncoder(w).Encode(o)
		return
	}
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

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/upload/storage/v1/b/")
	parts := strings.SplitN(rest, "/o/", 2)
	if len(parts) != 2 || !bucketRE.MatchString(parts[0]) {
		errJSON(w, 400, "invalid", "invalid upload path")
		return
	}
	if _, e := h.Store.GetGCPBucket(r.Context(), h.Project, parts[0]); e != nil {
		errJSON(w, 404, "notFound", "bucket not found")
		return
	}
	obj := parts[1]
	if obj == "" {
		errJSON(w, 400, "invalid", "object required")
		return
	}
	path, e := h.Store.NewObjectBodyPath("gcp", "storage", parts[0], obj)
	if e != nil {
		errJSON(w, 400, "invalid", "invalid object")
		return
	}
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		errJSON(w, 500, "internal", "storage failure")
		return
	}
	tmp := path + ".tmp"
	f, e := os.Create(tmp)
	if e != nil {
		errJSON(w, 500, "internal", "storage failure")
		return
	}
	n, e := io.Copy(f, io.LimitReader(r.Body, 64<<20))
	f.Close()
	if e != nil {
		os.Remove(tmp)
		errJSON(w, 400, "invalid", "upload failed")
		return
	}
	if e = os.Rename(tmp, path); e != nil {
		os.Remove(tmp)
		errJSON(w, 500, "internal", "storage failure")
		return
	}
	now := time.Now().UTC()
	o := state.GCPObject{Project: h.Project, Bucket: parts[0], Name: obj, BodyPath: path, ContentType: r.Header.Get("Content-Type"), ETag: "\"local\"", Generation: now.UnixNano(), Metageneration: 1, Size: n, CreatedAt: now, UpdatedAt: now}
	if e = h.Store.PutGCPObject(r.Context(), o); e != nil {
		errJSON(w, 500, "internal", "storage failure")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}
func errJSON(w http.ResponseWriter, code int, reason, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": code, "message": msg, "errors": []map[string]string{{"reason": reason, "message": msg}}}})
}

var _ = fmt.Sprint
