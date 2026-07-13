package aws

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emulith/emulith/internal/state"
)

const maxProtocolBody = 1 << 20

type Protocol string

const (
	ProtocolSQSJSON Protocol = "sqs-json"
	ProtocolQuery   Protocol = "query"
	ProtocolS3      Protocol = "s3"
	ProtocolUnknown Protocol = "unknown"
)

type Request struct {
	HTTPRequest *http.Request
	Protocol    Protocol
	Service     string
	Operation   string
	Form        url.Values
}

type Handler interface {
	ServeAWS(http.ResponseWriter, *Request, string)
}

type Gateway struct {
	store        *state.Store
	logger       *slog.Logger
	sts, s3, sqs Handler
}

func NewGateway(store *state.Store, logger *slog.Logger) *Gateway {
	p := placeholder{}
	return &Gateway{store: store, logger: logger, sts: p, s3: p, sqs: p}
}

func (g *Gateway) SetSTS(handler Handler) { g.sts = handler }
func (g *Gateway) SetS3(handler Handler)  { g.s3 = handler }
func (g *Gateway) SetSQS(handler Handler) { g.sqs = handler }

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	id := requestID()
	tracked := &statusWriter{ResponseWriter: w, status: http.StatusOK}
	defer func() {
		if recovered := recover(); recovered != nil {
			g.logger.Error("recovered AWS handler panic", "request_id", id, "method", r.Method, "path", r.URL.EscapedPath())
			if !tracked.wroteHeader {
				writeJSONError(tracked, id, http.StatusInternalServerError, "InternalError", "An internal error occurred")
			}
		}
	}()
	req, err := classify(r)
	if err != nil {
		writeJSONError(tracked, id, http.StatusRequestEntityTooLarge, "InvalidParameterValue", err.Error())
		g.log(r, req, tracked.status, id, started)
		return
	}
	switch req.Service {
	case "sts":
		g.sts.ServeAWS(tracked, req, id)
	case "sqs":
		g.sqs.ServeAWS(tracked, req, id)
	case "s3":
		g.s3.ServeAWS(tracked, req, id)
	default:
		writeQueryError(tracked, id, http.StatusBadRequest, "InvalidAction", "Unsupported AWS operation")
	}
	g.log(r, req, tracked.status, id, started)
}

func (g *Gateway) log(r *http.Request, req *Request, status int, id string, started time.Time) {
	service, operation := "unknown", "unknown"
	protocol := ProtocolUnknown
	if req != nil {
		service, operation, protocol = req.Service, req.Operation, req.Protocol
	}
	g.logger.Info("aws request", "method", r.Method, "path", r.URL.EscapedPath(), "provider", "aws", "service", service, "operation", operation, "protocol", protocol, "status", status, "duration_ms", time.Since(started).Milliseconds(), "request_id", id)
}

func classify(r *http.Request) (*Request, error) {
	req := &Request{HTTPRequest: r, Protocol: ProtocolUnknown}
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if target := r.Header.Get("X-Amz-Target"); r.Method == http.MethodPost && strings.EqualFold(mediaType, "application/x-amz-json-1.0") && strings.HasPrefix(strings.ToLower(target), "amazonsqs.") {
		req.Protocol, req.Service, req.Operation = ProtocolSQSJSON, "sqs", target[strings.LastIndex(target, ".")+1:]
		body, err := io.ReadAll(io.LimitReader(r.Body, maxProtocolBody+1))
		if err != nil {
			return req, err
		}
		if len(body) > maxProtocolBody {
			return req, fmt.Errorf("JSON body exceeds %d bytes", maxProtocolBody)
		}
		if len(bytes.TrimSpace(body)) > 0 && !json.Valid(body) {
			return req, fmt.Errorf("malformed JSON body")
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		return req, nil
	}
	form := r.URL.Query()
	if r.Method == http.MethodPost && strings.EqualFold(mediaType, "application/x-www-form-urlencoded") {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxProtocolBody+1))
		if err != nil {
			return req, err
		}
		if len(body) > maxProtocolBody {
			return req, fmt.Errorf("form body exceeds %d bytes", maxProtocolBody)
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		parsed, err := url.ParseQuery(string(body))
		if err != nil {
			return req, fmt.Errorf("parse form: %w", err)
		}
		form = merge(form, parsed)
		r.Body = io.NopCloser(bytes.NewReader(body))
	}
	if action := form.Get("Action"); action != "" {
		req.Protocol, req.Operation, req.Form = ProtocolQuery, action, form
		if isSTS(action) {
			req.Service = "sts"
		} else {
			req.Service = "sqs"
		}
		return req, nil
	}
	if plausibleS3(r) {
		req.Protocol, req.Service, req.Operation = ProtocolS3, "s3", s3Operation(r)
		return req, nil
	}
	return req, nil
}

func merge(a, b url.Values) url.Values {
	out := make(url.Values, len(a)+len(b))
	for key, values := range a {
		out[key] = append([]string(nil), values...)
	}
	for key, values := range b {
		for _, value := range values {
			out.Add(key, value)
		}
	}
	return out
}
func isSTS(action string) bool {
	switch action {
	case "GetCallerIdentity", "AssumeRole", "GetAccessKeyInfo", "GetSessionToken":
		return true
	}
	return false
}
func plausibleS3(r *http.Request) bool {
	if r.URL.Path != "" {
		return true
	}
	for key := range r.Header {
		if strings.HasPrefix(strings.ToLower(key), "x-amz-") {
			return true
		}
	}
	return false
}
func s3Operation(r *http.Request) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2" && len(parts) >= 1 {
		return "ListObjectsV2"
	}
	if r.URL.Path == "/" && r.Method == http.MethodGet {
		return "ListBuckets"
	}
	if len(parts) == 1 {
		return map[string]string{http.MethodPut: "CreateBucket", http.MethodDelete: "DeleteBucket", http.MethodHead: "HeadBucket"}[r.Method]
	}
	if len(parts) > 1 {
		return map[string]string{http.MethodPut: "PutObject", http.MethodGet: "GetObject", http.MethodDelete: "DeleteObject", http.MethodHead: "HeadObject"}[r.Method]
	}
	return "Unknown"
}

type placeholder struct{}

func (placeholder) ServeAWS(w http.ResponseWriter, req *Request, id string) {
	switch req.Protocol {
	case ProtocolS3:
		writeS3Error(w, id, http.StatusNotImplemented, "NotImplemented", "The requested operation is not implemented")
	case ProtocolSQSJSON:
		writeJSONError(w, id, http.StatusBadRequest, "InvalidAction", "The requested operation is not implemented")
	default:
		writeQueryError(w, id, http.StatusBadRequest, "InvalidAction", "The requested operation is not implemented")
	}
}

func requestID() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return fmt.Sprintf("local-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(data[:])
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

type s3Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId"`
}

func writeS3Error(w http.ResponseWriter, id string, status int, code, message string) {
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amz-request-id", id)
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(s3Error{Code: code, Message: message, RequestID: id})
}

type queryEnvelope struct {
	XMLName   xml.Name   `xml:"ErrorResponse"`
	Error     queryError `xml:"Error"`
	RequestID string     `xml:"RequestId"`
}
type queryError struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

func writeQueryError(w http.ResponseWriter, id string, status int, code, message string) {
	w.Header().Set("Content-Type", "text/xml")
	w.Header().Set("x-amzn-RequestId", id)
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(queryEnvelope{Error: queryError{Type: "Sender", Code: code, Message: message}, RequestID: id})
}
func writeJSONError(w http.ResponseWriter, id string, status int, code, message string) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", id)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"__type": code, "message": message})
}
