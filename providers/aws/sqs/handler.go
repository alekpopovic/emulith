package sqs

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
)

const accountID = "000000000000"
const queryNamespace = "http://queue.amazonaws.com/doc/2012-11-05/"

var queuePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)

type Handler struct {
	store *state.Store
	now   func() time.Time
}

func New(store *state.Store) *Handler { return &Handler{store: store, now: time.Now} }
func (h *Handler) ServeAWS(w http.ResponseWriter, req *awsprovider.Request, id string) {
	input, err := decodeInput(req)
	if err != nil {
		h.fail(w, req, id, 400, "InvalidParameterValue", "Malformed request")
		return
	}
	switch req.Operation {
	case "CreateQueue":
		h.create(w, req, id, input)
	case "GetQueueUrl":
		h.getURL(w, req, id, input)
	case "ListQueues":
		h.list(w, req, id, input)
	case "SendMessage":
		h.send(w, req, id, input)
	case "ReceiveMessage":
		h.receive(w, req, id, input)
	case "DeleteMessage":
		h.delete(w, req, id, input)
	case "PurgeQueue":
		h.purge(w, req, id, input)
	case "GetQueueAttributes":
		h.attributes(w, req, id, input)
	default:
		h.fail(w, req, id, 400, "InvalidAction", "The action is not valid")
	}
}
func decodeInput(req *awsprovider.Request) (map[string]any, error) {
	if req.Protocol == awsprovider.ProtocolSQSJSON {
		var input map[string]any
		decoder := json.NewDecoder(req.HTTPRequest.Body)
		if err := decoder.Decode(&input); err != nil {
			return nil, err
		}
		return input, nil
	}
	input := map[string]any{}
	for k, v := range req.Form {
		if len(v) > 0 {
			input[k] = v[0]
		}
	}
	return input, nil
}
func str(input map[string]any, key string) string { v, _ := input[key].(string); return v }
func integer(input map[string]any, key string, fallback int) (int, error) {
	v, ok := input[key]
	if !ok {
		return fallback, nil
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case string:
		return strconv.Atoi(n)
	default:
		return 0, errors.New("invalid integer")
	}
}
func queueNameFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != accountID {
		return ""
	}
	return parts[1]
}
func publicBase(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
func (h *Handler) queue(req *awsprovider.Request, input map[string]any) (state.SQSQueue, error) {
	name := queueNameFromURL(str(input, "QueueUrl"))
	if name == "" {
		return state.SQSQueue{}, state.ErrNotFound
	}
	return h.store.GetSQSQueue(req.HTTPRequest.Context(), name)
}
func (h *Handler) create(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	name := str(input, "QueueName")
	if name == "" {
		h.fail(w, req, id, 400, "MissingParameter", "QueueName is required")
		return
	}
	if strings.HasSuffix(name, ".fifo") {
		h.fail(w, req, id, 400, "UnsupportedOperation", "FIFO queues are not supported")
		return
	}
	if !queuePattern.MatchString(name) {
		h.fail(w, req, id, 400, "InvalidParameterValue", "Invalid queue name")
		return
	}
	visibility := 30
	var err error
	if attrs, ok := input["Attributes"].(map[string]any); ok && attrs["VisibilityTimeout"] != nil {
		visibility, err = strconv.Atoi(fmt.Sprint(attrs["VisibilityTimeout"]))
	} else if str(input, "Attribute.1.Name") == "VisibilityTimeout" {
		visibility, err = strconv.Atoi(str(input, "Attribute.1.Value"))
	}
	if err != nil || visibility < 0 || visibility > 43200 {
		h.fail(w, req, id, 400, "InvalidAttributeValue", "Invalid visibility timeout")
		return
	}
	if existing, getErr := h.store.GetSQSQueue(req.HTTPRequest.Context(), name); getErr == nil {
		if existing.VisibilityTimeout != visibility {
			h.fail(w, req, id, 400, "QueueNameExists", "A queue with different attributes already exists")
			return
		}
		queueURL := publicBase(req.HTTPRequest) + existing.URLPath
		h.respond(w, req, id, map[string]any{"QueueUrl": queueURL}, "<QueueUrl>"+escape(queueURL)+"</QueueUrl>")
		return
	} else if !errors.Is(getErr, state.ErrNotFound) {
		h.internal(w, req, id)
		return
	}
	path := "/" + accountID + "/" + name
	q, err := h.store.CreateSQSQueue(req.HTTPRequest.Context(), state.SQSQueue{Name: name, URLPath: path, VisibilityTimeout: visibility, CreatedAt: h.now().UTC()})
	if err != nil {
		h.internal(w, req, id)
		return
	}
	h.respond(w, req, id, map[string]any{"QueueUrl": publicBase(req.HTTPRequest) + q.URLPath}, "<QueueUrl>"+escape(publicBase(req.HTTPRequest)+q.URLPath)+"</QueueUrl>")
}
func (h *Handler) getURL(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	name := str(input, "QueueName")
	if name == "" {
		h.fail(w, req, id, 400, "MissingParameter", "QueueName is required")
		return
	}
	q, err := h.store.GetSQSQueue(req.HTTPRequest.Context(), name)
	if errors.Is(err, state.ErrNotFound) {
		h.nonexistent(w, req, id)
		return
	}
	if err != nil {
		h.internal(w, req, id)
		return
	}
	url := publicBase(req.HTTPRequest) + q.URLPath
	h.respond(w, req, id, map[string]any{"QueueUrl": url}, "<QueueUrl>"+escape(url)+"</QueueUrl>")
}
func (h *Handler) list(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	queues, err := h.store.ListSQSQueues(req.HTTPRequest.Context(), str(input, "QueueNamePrefix"))
	if err != nil {
		h.internal(w, req, id)
		return
	}
	urls := make([]string, 0, len(queues))
	inner := ""
	for _, q := range queues {
		u := publicBase(req.HTTPRequest) + q.URLPath
		urls = append(urls, u)
		inner += "<QueueUrl>" + escape(u) + "</QueueUrl>"
	}
	h.respond(w, req, id, map[string]any{"QueueUrls": urls}, inner)
}
func (h *Handler) send(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	q, err := h.queue(req, input)
	if errors.Is(err, state.ErrNotFound) {
		h.nonexistent(w, req, id)
		return
	}
	if err != nil {
		h.internal(w, req, id)
		return
	}
	body, ok := input["MessageBody"].(string)
	if !ok {
		h.fail(w, req, id, 400, "MissingParameter", "MessageBody is required")
		return
	}
	if len(body) > 256<<10 || !utf8.ValidString(body) {
		h.fail(w, req, id, 400, "InvalidParameterValue", "Message body is invalid or too large")
		return
	}
	m, err := h.store.SendSQSMessage(req.HTTPRequest.Context(), q.Name, body, h.now().UTC())
	if err != nil {
		h.internal(w, req, id)
		return
	}
	h.respond(w, req, id, map[string]any{"MessageId": m.ID, "MD5OfMessageBody": m.MD5}, "<MD5OfMessageBody>"+m.MD5+"</MD5OfMessageBody><MessageId>"+m.ID+"</MessageId>")
}
func (h *Handler) receive(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	q, err := h.queue(req, input)
	if errors.Is(err, state.ErrNotFound) {
		h.nonexistent(w, req, id)
		return
	}
	if err != nil {
		h.internal(w, req, id)
		return
	}
	max, err := integer(input, "MaxNumberOfMessages", 1)
	if err != nil || max < 1 || max > 10 {
		h.fail(w, req, id, 400, "InvalidParameterValue", "MaxNumberOfMessages must be 1 through 10")
		return
	}
	visibility, err := integer(input, "VisibilityTimeout", q.VisibilityTimeout)
	if err != nil || visibility < 0 || visibility > 43200 {
		h.fail(w, req, id, 400, "InvalidParameterValue", "Invalid visibility timeout")
		return
	}
	messages, err := h.store.ReceiveSQSMessages(req.HTTPRequest.Context(), q.Name, max, visibility, h.now().UTC())
	if err != nil {
		h.internal(w, req, id)
		return
	}
	jsonMessages := make([]map[string]any, 0, len(messages))
	inner := ""
	for _, m := range messages {
		jsonMessages = append(jsonMessages, map[string]any{"MessageId": m.ID, "ReceiptHandle": m.ReceiptHandle, "MD5OfBody": m.MD5, "Body": m.Body})
		inner += "<Message><MessageId>" + m.ID + "</MessageId><ReceiptHandle>" + m.ReceiptHandle + "</ReceiptHandle><MD5OfBody>" + m.MD5 + "</MD5OfBody><Body>" + escape(m.Body) + "</Body></Message>"
	}
	h.respond(w, req, id, map[string]any{"Messages": jsonMessages}, inner)
}
func (h *Handler) delete(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	q, err := h.queue(req, input)
	if errors.Is(err, state.ErrNotFound) {
		h.nonexistent(w, req, id)
		return
	}
	if err != nil {
		h.internal(w, req, id)
		return
	}
	receipt := str(input, "ReceiptHandle")
	if receipt == "" {
		h.fail(w, req, id, 400, "MissingParameter", "ReceiptHandle is required")
		return
	}
	if err := h.store.DeleteSQSMessage(req.HTTPRequest.Context(), q.Name, receipt); errors.Is(err, state.ErrNotFound) {
		h.fail(w, req, id, 400, "ReceiptHandleIsInvalid", "The receipt handle is invalid")
		return
	} else if err != nil {
		h.internal(w, req, id)
		return
	}
	h.respond(w, req, id, map[string]any{}, "")
}
func (h *Handler) purge(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	q, err := h.queue(req, input)
	if errors.Is(err, state.ErrNotFound) {
		h.nonexistent(w, req, id)
		return
	}
	if err != nil {
		h.internal(w, req, id)
		return
	}
	if err := h.store.PurgeSQSQueue(req.HTTPRequest.Context(), q.Name); err != nil {
		h.internal(w, req, id)
		return
	}
	h.respond(w, req, id, map[string]any{}, "")
}
func (h *Handler) attributes(w http.ResponseWriter, req *awsprovider.Request, id string, input map[string]any) {
	q, err := h.queue(req, input)
	if errors.Is(err, state.ErrNotFound) {
		h.nonexistent(w, req, id)
		return
	}
	if err != nil {
		h.internal(w, req, id)
		return
	}
	visible, hidden, err := h.store.SQSMessageCounts(req.HTTPRequest.Context(), q.Name, h.now().UTC())
	if err != nil {
		h.internal(w, req, id)
		return
	}
	attrs := map[string]string{"ApproximateNumberOfMessages": strconv.Itoa(visible), "ApproximateNumberOfMessagesNotVisible": strconv.Itoa(hidden), "QueueArn": "arn:aws:sqs:us-east-1:" + accountID + ":" + q.Name, "CreatedTimestamp": strconv.FormatInt(q.CreatedAt.Unix(), 10), "VisibilityTimeout": strconv.Itoa(q.VisibilityTimeout)}
	inner := ""
	for _, name := range []string{"ApproximateNumberOfMessages", "ApproximateNumberOfMessagesNotVisible", "QueueArn", "CreatedTimestamp", "VisibilityTimeout"} {
		inner += "<Attribute><Name>" + name + "</Name><Value>" + attrs[name] + "</Value></Attribute>"
	}
	h.respond(w, req, id, map[string]any{"Attributes": attrs}, inner)
}
func (h *Handler) respond(w http.ResponseWriter, req *awsprovider.Request, id string, jsonValue any, queryInner string) {
	w.Header().Set("x-amzn-RequestId", id)
	if req.Protocol == awsprovider.ProtocolSQSJSON {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(jsonValue)
		return
	}
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, "<%sResponse xmlns=%q><%sResult>%s</%sResult><ResponseMetadata><RequestId>%s</RequestId></ResponseMetadata></%sResponse>", req.Operation, queryNamespace, req.Operation, queryInner, req.Operation, id, req.Operation)
}
func (h *Handler) fail(w http.ResponseWriter, req *awsprovider.Request, id string, status int, code, message string) {
	w.Header().Set("x-amzn-RequestId", id)
	if req.Protocol == awsprovider.ProtocolSQSJSON {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]string{"__type": code, "message": message})
		return
	}
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, "<ErrorResponse xmlns=%q><Error><Type>Sender</Type><Code>%s</Code><Message>%s</Message></Error><RequestId>%s</RequestId></ErrorResponse>", queryNamespace, escape(code), escape(message), id)
}
func (h *Handler) nonexistent(w http.ResponseWriter, req *awsprovider.Request, id string) {
	h.fail(w, req, id, 400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist")
}
func (h *Handler) internal(w http.ResponseWriter, req *awsprovider.Request, id string) {
	h.fail(w, req, id, 500, "InternalError", "An internal storage error occurred")
}
func escape(value string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}
