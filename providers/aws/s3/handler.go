package s3

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
)

const namespace = "http://s3.amazonaws.com/doc/2006-03-01/"

var bucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

type Handler struct {
	store *state.Store
	now   func() time.Time
}

func New(store *state.Store) *Handler { return &Handler{store: store, now: time.Now} }

func (h *Handler) ServeAWS(w http.ResponseWriter, req *awsprovider.Request, id string) {
	switch req.Operation {
	case "CreateBucket":
		h.createBucket(w, req, id)
	case "ListBuckets":
		h.listBuckets(w, req, id)
	case "PutObject":
		h.putObject(w, req, id)
	case "GetObject":
		h.getObject(w, req, id, false)
	case "HeadObject":
		h.getObject(w, req, id, true)
	case "DeleteObject":
		h.deleteObject(w, req, id)
	case "ListObjectsV2":
		h.listObjects(w, req, id)
	default:
		writeError(w, id, http.StatusNotImplemented, "NotImplemented", "The requested operation is not implemented")
	}
}
func bucketKey(r *http.Request) (string, string) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
func validBucket(name string) bool {
	return bucketPattern.MatchString(name) && !strings.Contains(name, "..") && !looksLikeIPv4(name)
}
func looksLikeIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		n, e := strconv.Atoi(part)
		if e != nil || n < 0 || n > 255 {
			return false
		}
	}
	return true
}
func (h *Handler) createBucket(w http.ResponseWriter, req *awsprovider.Request, id string) {
	bucket, _ := bucketKey(req.HTTPRequest)
	if !validBucket(bucket) {
		writeError(w, id, 400, "InvalidBucketName", "The specified bucket is not valid")
		return
	}
	region := "us-east-1"
	if req.HTTPRequest.ContentLength != 0 {
		var input struct {
			Location string `xml:"LocationConstraint"`
		}
		if err := xml.NewDecoder(http.MaxBytesReader(w, req.HTTPRequest.Body, 64<<10)).Decode(&input); err != nil {
			writeError(w, id, 400, "MalformedXML", "The XML is not well-formed")
			return
		}
		if input.Location != "" {
			region = input.Location
		}
	}
	if err := h.store.CreateS3Bucket(req.HTTPRequest.Context(), state.S3Bucket{Name: bucket, Region: region, CreatedAt: h.now().UTC()}); err != nil {
		writeError(w, id, 409, "BucketAlreadyExists", "The requested bucket name already exists")
		return
	}
	w.Header().Set("Location", "/"+bucket)
	requestHeader(w, id)
	w.WriteHeader(http.StatusOK)
}

type listBucketsResult struct {
	XMLName xml.Name   `xml:"ListAllMyBucketsResult"`
	XMLNS   string     `xml:"xmlns,attr"`
	Owner   owner      `xml:"Owner"`
	Buckets bucketList `xml:"Buckets"`
}
type owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}
type bucketList struct {
	Buckets []bucketXML `xml:"Bucket"`
}
type bucketXML struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

func (h *Handler) listBuckets(w http.ResponseWriter, req *awsprovider.Request, id string) {
	buckets, err := h.store.ListS3Buckets(req.HTTPRequest.Context())
	if err != nil {
		internal(w, id)
		return
	}
	out := listBucketsResult{XMLNS: namespace, Owner: owner{ID: "EMULITHUSER", DisplayName: "emulith"}}
	for _, b := range buckets {
		out.Buckets.Buckets = append(out.Buckets.Buckets, bucketXML{Name: b.Name, CreationDate: b.CreatedAt.UTC().Format(time.RFC3339)})
	}
	writeXML(w, id, http.StatusOK, out)
}
func (h *Handler) putObject(w http.ResponseWriter, req *awsprovider.Request, id string) {
	bucket, key := bucketKey(req.HTTPRequest)
	if key == "" {
		writeError(w, id, 400, "InvalidArgument", "Object key is required")
		return
	}
	if !h.bucketExists(w, req, id, bucket) {
		return
	}
	pending, err := h.store.StreamObjectBody("aws", "s3", bucket, key, req.HTTPRequest.Body)
	if err != nil {
		internal(w, id)
		return
	}
	etag := "\"" + pending.MD5 + "\""
	old, err := h.store.PutS3Object(req.HTTPRequest.Context(), state.S3Object{Bucket: bucket, Key: key, ETag: etag, Size: pending.Size, ContentType: req.HTTPRequest.Header.Get("Content-Type"), LastModified: h.now().UTC(), BodyPath: pending.FinalPath})
	if err != nil {
		_ = h.store.RemoveBody(pending.FinalPath)
		internal(w, id)
		return
	}
	if old != "" && old != pending.FinalPath {
		_ = h.store.RemoveBody(old)
	}
	requestHeader(w, id)
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) getObject(w http.ResponseWriter, req *awsprovider.Request, id string, head bool) {
	bucket, key := bucketKey(req.HTTPRequest)
	if req.HTTPRequest.Header.Get("Range") != "" {
		writeError(w, id, 501, "NotImplemented", "Range requests are not implemented")
		return
	}
	if !h.bucketExists(w, req, id, bucket) {
		return
	}
	object, err := h.store.GetS3Object(req.HTTPRequest.Context(), bucket, key)
	if errors.Is(err, state.ErrNotFound) {
		writeError(w, id, 404, "NoSuchKey", "The specified key does not exist")
		return
	}
	if err != nil {
		internal(w, id)
		return
	}
	file, err := os.Open(object.BodyPath)
	if err != nil {
		internal(w, id)
		return
	}
	defer file.Close()
	requestHeader(w, id)
	w.Header().Set("ETag", object.ETag)
	w.Header().Set("Content-Length", strconv.FormatInt(object.Size, 10))
	w.Header().Set("Last-Modified", object.LastModified.UTC().Format(http.TimeFormat))
	if object.ContentType != "" {
		w.Header().Set("Content-Type", object.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.WriteHeader(200)
	if !head {
		_, _ = io.Copy(w, file)
	}
}
func (h *Handler) deleteObject(w http.ResponseWriter, req *awsprovider.Request, id string) {
	bucket, key := bucketKey(req.HTTPRequest)
	if !h.bucketExists(w, req, id, bucket) {
		return
	}
	path, err := h.store.DeleteS3Object(req.HTTPRequest.Context(), bucket, key)
	if err != nil {
		internal(w, id)
		return
	}
	if path != "" {
		_ = h.store.RemoveBody(path)
	}
	requestHeader(w, id)
	w.WriteHeader(204)
}

type listResult struct {
	XMLName     xml.Name     `xml:"ListBucketResult"`
	XMLNS       string       `xml:"xmlns,attr"`
	Name        string       `xml:"Name"`
	Prefix      string       `xml:"Prefix"`
	KeyCount    int          `xml:"KeyCount"`
	MaxKeys     int          `xml:"MaxKeys"`
	IsTruncated bool         `xml:"IsTruncated"`
	Contents    []contentXML `xml:"Contents"`
}
type contentXML struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

func (h *Handler) listObjects(w http.ResponseWriter, req *awsprovider.Request, id string) {
	bucket, _ := bucketKey(req.HTTPRequest)
	if !h.bucketExists(w, req, id, bucket) {
		return
	}
	max := 1000
	if raw := req.HTTPRequest.URL.Query().Get("max-keys"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n > 1000 {
			writeError(w, id, 400, "InvalidArgument", "max-keys must be between 0 and 1000")
			return
		}
		max = n
	}
	prefix := req.HTTPRequest.URL.Query().Get("prefix")
	objects, err := h.store.ListS3Objects(req.HTTPRequest.Context(), bucket, prefix, max)
	if err != nil {
		internal(w, id)
		return
	}
	out := listResult{XMLNS: namespace, Name: bucket, Prefix: prefix, KeyCount: len(objects), MaxKeys: max, IsTruncated: false}
	for _, o := range objects {
		out.Contents = append(out.Contents, contentXML{Key: o.Key, LastModified: o.LastModified.UTC().Format(time.RFC3339), ETag: o.ETag, Size: o.Size, StorageClass: "STANDARD"})
	}
	writeXML(w, id, 200, out)
}
func (h *Handler) bucketExists(w http.ResponseWriter, req *awsprovider.Request, id, bucket string) bool {
	ok, err := h.store.S3BucketExists(req.HTTPRequest.Context(), bucket)
	if err != nil {
		internal(w, id)
		return false
	}
	if !ok {
		writeError(w, id, 404, "NoSuchBucket", "The specified bucket does not exist")
		return false
	}
	return true
}
func requestHeader(w http.ResponseWriter, id string) { w.Header().Set("x-amz-request-id", id) }
func writeXML(w http.ResponseWriter, id string, status int, value any) {
	requestHeader(w, id)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, id string, status int, code, message string) {
	writeXML(w, id, status, struct {
		XMLName   xml.Name `xml:"Error"`
		Code      string   `xml:"Code"`
		Message   string   `xml:"Message"`
		RequestID string   `xml:"RequestId"`
	}{Code: code, Message: message, RequestID: id})
}
func internal(w http.ResponseWriter, id string) {
	writeError(w, id, 500, "InternalError", "An internal storage error occurred")
}
