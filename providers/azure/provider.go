package azure

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"github.com/alekpopovic/emulith/internal/state"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const DefaultAccountName = "devstoreaccount1"
const DefaultAccountKey = "RW11bGl0aC1kZXZlbG9wbWVudC1rZXk="

type Handler struct {
	Service, Account string
	Store            *state.Store
}

func New(service, account string) *Handler {
	if account == "" {
		account = DefaultAccountName
	}
	return &Handler{Service: service, Account: account}
}
func NewWithStore(service, account string, s *state.Store) *Handler {
	h := New(service, account)
	h.Store = s
	return h
}
func requestID() string {
	b := make([]byte, 8)
	if _, e := rand.Read(b); e != nil {
		return "local-request"
	}
	return hex.EncodeToString(b)
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestID()
	w.Header().Set("x-ms-request-id", id)
	w.Header().Set("x-ms-version", "2023-11-03")
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != h.Account {
		writeError(w, 404, "AccountNotFound", "The specified account does not exist.", id)
		return
	}
	if h.Store == nil {
		writeError(w, 501, "UnsupportedOperation", "Not configured", id)
		return
	}
	if strings.EqualFold(h.Service, "queue") || strings.EqualFold(h.Service, "queues") {
		h.queue(w, r, parts, id)
		return
	}
	container, _ := url.PathUnescape(parts[1])
	blob := ""
	if len(parts) > 2 {
		blob, _ = url.PathUnescape(strings.Join(parts[2:], "/"))
	}
	if blob != "" {
		h.blob(w, r, container, blob, id)
		return
	}
	h.container(w, r, container, id)
}
func validQueue(n string) bool {
	if len(n) < 3 || len(n) > 63 || n[0] == '-' || n[len(n)-1] == '-' || strings.Contains(n, "--") {
		return false
	}
	for _, c := range n {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}
func (h *Handler) queue(w http.ResponseWriter, r *http.Request, parts []string, id string) {
	if len(parts) > 2 {
		if len(parts) >= 3 && parts[2] == "messages" {
			h.queueMessages(w, r, parts, id)
			return
		}
		writeError(w, 404, "InvalidUri", "Queue path is invalid.", id)
		return
	}
	var n string
	if len(parts) == 2 {
		n, _ = url.PathUnescape(parts[1])
	}
	qv := r.URL.Query()
	if n == "" && qv.Get("comp") == "list" {
		qs, e := h.Store.ListAzureQueues(r.Context(), h.Account, qv.Get("prefix"))
		if e != nil {
			writeError(w, 500, "InternalError", "", id)
			return
		}
		marker := qv.Get("marker")
		max := 100
		if qv.Get("maxresults") != "" {
			fmt.Sscanf(qv.Get("maxresults"), "%d", &max)
			if max < 1 {
				max = 1
			}
			if max > 5000 {
				max = 5000
			}
		}
		start := 0
		if marker != "" {
			for start < len(qs) && qs[start].Name <= marker {
				start++
			}
		}
		end := start + max
		if end > len(qs) {
			end = len(qs)
		}
		type Item struct {
			Name     string            `xml:"Name"`
			Metadata map[string]string `xml:"-"`
		}
		type L struct {
			XMLName xml.Name `xml:"EnumerationResults"`
			Queues  []Item   `xml:"Queues>Queue"`
			Next    string   `xml:"NextMarker,omitempty"`
		}
		out := L{}
		for _, q := range qs[start:end] {
			out.Queues = append(out.Queues, Item{Name: q.Name})
		}
		if end < len(qs) {
			out.Next = qs[end-1].Name
		}
		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(out)
		return
	}
	if !validQueue(n) {
		writeError(w, 400, "InvalidResourceName", "Invalid queue name", id)
		return
	}
	ctx := r.Context()
	switch r.Method {
	case "PUT":
		if qv.Get("comp") == "metadata" {
			q, e := h.Store.GetAzureQueue(ctx, h.Account, n)
			if e != nil {
				writeError(w, 404, "QueueNotFound", "The specified queue does not exist.", id)
				return
			}
			q.Metadata = metadata(r)
			q.ETag = "\"" + requestID() + "\""
			q.LastModified = time.Now().UTC()
			_ = h.Store.UpdateAzureQueue(ctx, q)
			w.WriteHeader(204)
			return
		}
		now := time.Now().UTC()
		e := h.Store.CreateAzureQueue(ctx, state.AzureQueue{Account: h.Account, Name: n, ETag: "\"" + requestID() + "\"", LastModified: now, CreatedAt: now, Metadata: metadata(r)})
		if e == state.ErrConflict {
			writeError(w, 409, "QueueAlreadyExists", "The specified queue already exists.", id)
			return
		}
		w.WriteHeader(201)
	case "DELETE":
		if e := h.Store.DeleteAzureQueue(ctx, h.Account, n); e != nil {
			writeError(w, 404, "QueueNotFound", "The specified queue does not exist.", id)
			return
		}
		w.WriteHeader(204)
	case "GET", "HEAD":
		q, e := h.Store.GetAzureQueue(ctx, h.Account, n)
		if e != nil {
			writeError(w, 404, "QueueNotFound", "The specified queue does not exist.", id)
			return
		}
		w.Header().Set("ETag", q.ETag)
		w.Header().Set("Last-Modified", q.LastModified.UTC().Format(http.TimeFormat))
		w.Header().Set("x-ms-approximate-messages-count", "0")
		for k, v := range q.Metadata {
			w.Header().Set("x-ms-meta-"+k, v)
		}
		w.WriteHeader(200)
	default:
		writeError(w, 405, "UnsupportedOperation", "Unsupported method", id)
	}
}

type queueMessageXML struct {
	MessageText string `xml:"MessageText"`
}

func (h *Handler) queueMessages(w http.ResponseWriter, r *http.Request, parts []string, id string) {
	q := parts[1]
	ctx := r.Context()
	if _, e := h.Store.GetAzureQueue(ctx, h.Account, q); e != nil {
		writeError(w, 404, "QueueNotFound", "Queue not found", id)
		return
	}
	if len(parts) > 3 {
		mid, _ := url.PathUnescape(parts[3])
		if r.Method == "DELETE" {
			if e := h.Store.DeleteAzureMessage(ctx, h.Account, q, mid, r.URL.Query().Get("popreceipt")); e != nil {
				writeError(w, 404, "MessageNotFound", "Message not found", id)
				return
			}
			w.WriteHeader(204)
			return
		}
		if r.Method == "PUT" {
			b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			var x queueMessageXML
			_ = xml.Unmarshal(b, &x)
			m, e := h.Store.UpdateAzureMessage(ctx, h.Account, q, mid, r.URL.Query().Get("popreceipt"), x.MessageText, time.Duration(parseInt(r.URL.Query().Get("visibilitytimeout"), 30))*time.Second)
			if e != nil {
				writeError(w, 404, "MessageNotFound", "Message not found", id)
				return
			}
			writeQueueMessage(w, m, id)
			return
		}
	}
	switch r.Method {
	case "POST":
		b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		var x queueMessageXML
		if xml.Unmarshal(b, &x) != nil {
			writeError(w, 400, "InvalidXmlDocument", "Invalid XML", id)
			return
		}
		m, e := h.Store.PutAzureMessage(ctx, h.Account, q, x.MessageText, time.Duration(parseInt(r.URL.Query().Get("messagettl"), 604800))*time.Second, time.Duration(parseInt(r.URL.Query().Get("visibilitytimeout"), 0))*time.Second)
		if e != nil {
			writeError(w, 400, "MessageTooLarge", "Message rejected", id)
			return
		}
		writeQueueMessage(w, m, id)
	case "GET":
		peek := r.URL.Query().Get("peekonly") == "true"
		n := parseInt(r.URL.Query().Get("numofmessages"), 1)
		if n < 1 || n > 32 {
			writeError(w, 400, "OutOfRangeInput", "numofmessages out of range", id)
			return
		}
		ms, e := h.Store.QueueMessages(ctx, h.Account, q, peek, n, time.Duration(parseInt(r.URL.Query().Get("visibilitytimeout"), 30))*time.Second)
		if e != nil {
			writeError(w, 500, "InternalError", "", id)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		type out struct {
			XMLName  xml.Name                  `xml:"QueueMessagesList"`
			Messages []state.AzureQueueMessage `xml:"QueueMessage"`
		}
		o := out{}
		for _, m := range ms {
			o.Messages = append(o.Messages, m)
		}
		xml.NewEncoder(w).Encode(o)
	case "DELETE":
		if e := h.Store.ClearAzureMessages(ctx, h.Account, q); e != nil {
			writeError(w, 500, "InternalError", "", id)
			return
		}
		w.WriteHeader(204)
	default:
		writeError(w, 405, "UnsupportedOperation", "", id)
	}
}
func parseInt(s string, d int) int {
	if s == "" {
		return d
	}
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
func writeQueueMessage(w http.ResponseWriter, m state.AzureQueueMessage, id string) {
	w.Header().Set("Content-Type", "application/xml")
	type out struct {
		XMLName    xml.Name  `xml:"QueueMessage"`
		ID         string    `xml:"MessageId"`
		Insertion  time.Time `xml:"InsertionTime"`
		Expiration time.Time `xml:"ExpirationTime"`
		Pop        string    `xml:"PopReceipt"`
		Visible    time.Time `xml:"TimeNextVisible"`
		Text       string    `xml:"MessageText"`
		Count      int       `xml:"DequeueCount"`
	}
	xml.NewEncoder(w).Encode(out{ID: m.ID, Insertion: m.InsertedAt, Expiration: m.ExpiresAt, Pop: m.PopReceipt, Visible: m.VisibleAt, Text: m.Body, Count: m.DequeueCount})
}
func validContainer(n string) bool {
	if len(n) < 3 || len(n) > 63 || n[0] == '-' || n[len(n)-1] == '-' || strings.Contains(n, "--") {
		return false
	}
	for _, c := range n {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}
func metadata(r *http.Request) map[string]string {
	m := map[string]string{}
	for k, v := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-ms-meta-") && len(v) > 0 {
			m[strings.ToLower(strings.TrimPrefix(strings.ToLower(k), "x-ms-meta-"))] = v[0]
		}
	}
	return m
}
func (h *Handler) container(w http.ResponseWriter, r *http.Request, n, id string) {
	ctx := r.Context()
	if r.URL.Query().Get("comp") == "list" {
		cs, _ := h.Store.ListAzureContainers(ctx, h.Account, r.URL.Query().Get("prefix"))
		type C struct {
			Name string            `xml:"Name"`
			Meta map[string]string `xml:"-"`
		}
		type L struct {
			XMLName    xml.Name `xml:"EnumerationResults"`
			Containers []C      `xml:"Containers>Container"`
		}
		out := L{}
		for _, c := range cs {
			out.Containers = append(out.Containers, C{Name: c.Name})
		}
		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(out)
		return
	}
	switch r.Method {
	case "PUT":
		if !validContainer(n) {
			writeError(w, 400, "InvalidResourceName", "Invalid container name", id)
			return
		}
		e := h.Store.CreateAzureContainer(ctx, state.AzureContainer{Account: h.Account, Name: n, ETag: "\"" + requestID() + "\"", LastModified: time.Now().UTC(), CreatedAt: time.Now().UTC(), Metadata: metadata(r)})
		if e == state.ErrConflict {
			writeError(w, 409, "ContainerAlreadyExists", "The specified container already exists.", id)
			return
		}
		w.WriteHeader(201)
	case "DELETE":
		e := h.Store.DeleteAzureContainer(ctx, h.Account, n)
		if e != nil {
			writeError(w, 404, "ContainerNotFound", "The specified container does not exist.", id)
			return
		}
		w.WriteHeader(202)
	case "HEAD", "GET":
		c, e := h.Store.GetAzureContainer(ctx, h.Account, n)
		if e != nil {
			writeError(w, 404, "ContainerNotFound", "The specified container does not exist.", id)
			return
		}
		w.Header().Set("ETag", c.ETag)
		w.Header().Set("Last-Modified", c.LastModified.UTC().Format(http.TimeFormat))
		for k, v := range c.Metadata {
			w.Header().Set("x-ms-meta-"+k, v)
		}
		if r.Method == "GET" {
			w.WriteHeader(200)
		}
	default:
		writeError(w, 405, "UnsupportedOperation", "Unsupported method", id)
	}
}
func (h *Handler) blob(w http.ResponseWriter, r *http.Request, c, n, id string) {
	if r.Method != "PUT" && r.Method != "GET" && r.Method != "HEAD" && r.Method != "DELETE" {
		writeError(w, 405, "UnsupportedOperation", "Unsupported method", id)
		return
	}
	if _, e := h.Store.GetAzureContainer(r.Context(), h.Account, c); e != nil {
		writeError(w, 404, "ContainerNotFound", "The specified container does not exist.", id)
		return
	}
	if r.Method == "PUT" {
		if r.Header.Get("x-ms-blob-type") != "BlockBlob" {
			writeError(w, 400, "InvalidHeaderValue", "Blob type required", id)
			return
		}
		p, e := h.Store.StreamObjectBody("azure", "blob", h.Account, c+"/"+n, r.Body)
		if e != nil {
			writeError(w, 500, "InternalError", "upload failed", id)
			return
		}
		now := time.Now().UTC()
		old, e := h.Store.PutAzureBlob(r.Context(), state.AzureBlob{Account: h.Account, Container: c, Name: n, ETag: "\"" + requestID() + "\"", BodyPath: p.FinalPath, Size: p.Size, LastModified: now, CreatedAt: now, ContentType: r.Header.Get("Content-Type"), Metadata: metadata(r)})
		if old != "" {
			_ = h.Store.RemoveBody(old)
		}
		if e != nil {
			writeError(w, 500, "InternalError", "persist failed", id)
			return
		}
		w.WriteHeader(201)
		return
	}
	if r.Method == "DELETE" {
		p, e := h.Store.DeleteAzureBlob(r.Context(), h.Account, c, n)
		if e != nil {
			writeError(w, 404, "BlobNotFound", "The specified blob does not exist.", id)
			return
		}
		if p != "" {
			_ = h.Store.RemoveBody(p)
		}
		w.WriteHeader(202)
		return
	}
	b, e := h.Store.GetAzureBlob(r.Context(), h.Account, c, n)
	if e != nil {
		writeError(w, 404, "BlobNotFound", "The specified blob does not exist.", id)
		return
	}
	w.Header().Set("ETag", b.ETag)
	w.Header().Set("Content-Length", fmt.Sprint(b.Size))
	w.Header().Set("x-ms-blob-type", "BlockBlob")
	if b.ContentType != "" {
		w.Header().Set("Content-Type", b.ContentType)
	}
	if r.Method == "GET" {
		f, e := os.Open(b.BodyPath)
		if e != nil {
			writeError(w, 500, "InternalError", "", id)
			return
		}
		defer f.Close()
		io.Copy(w, f)
	}
}

type errorBody struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId"`
}

func writeError(w http.ResponseWriter, s int, c, m, id string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(s)
	xml.NewEncoder(w).Encode(errorBody{Code: c, Message: m, RequestID: id})
}
