package sns

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const account = "000000000000"
const ns = "http://sns.amazonaws.com/doc/2010-03-31/"

var nameRE = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

type Handler struct {
	store  *state.Store
	region string
}

func New(s *state.Store, region string) *Handler {
	if region == "" {
		region = "us-east-1"
	}
	return &Handler{store: s, region: region}
}
func (h *Handler) ServeAWS(w http.ResponseWriter, r *awsprovider.Request, id string) {
	in := map[string]string{}
	for k, v := range r.Form {
		if len(v) > 0 {
			in[k] = v[0]
		}
	}
	switch r.Operation {
	case "CreateTopic":
		h.create(w, r, id, in)
	case "ListTopics":
		h.list(w, r, id, in)
	case "GetTopicAttributes":
		h.attrs(w, r, id, in)
	case "DeleteTopic":
		h.delete(w, r, id, in)
	case "Publish":
		h.publish(w, r, id, in)
	case "Subscribe":
		h.subscribe(w, r, id, in)
	case "Unsubscribe":
		h.unsubscribe(w, r, id, in)
	case "ListSubscriptions", "ListSubscriptionsByTopic":
		h.listSubs(w, r, id, in, r.Operation == "ListSubscriptionsByTopic")
	case "GetSubscriptionAttributes":
		h.subAttrs(w, r, id, in)
	case "SetSubscriptionAttributes":
		h.setSubAttrs(w, r, id, in)
	default:
		h.fail(w, id, 400, "InvalidAction", "The action is not valid")
	}
}

type envelope struct {
	XMLName xml.Name `xml:"Response"`
	Xmlns   string   `xml:"xmlns,attr"`
	Body    any      `xml:"-"`
}

func (h *Handler) write(w http.ResponseWriter, id string, name string, body any) {
	w.Header().Set("Content-Type", "text/xml")
	w.Header().Set("x-amzn-RequestId", id)
	w.WriteHeader(200)
	_, _ = w.Write([]byte(xml.Header))
	b, _ := xml.Marshal(body)
	_, _ = w.Write(b)
}
func (h *Handler) fail(w http.ResponseWriter, id string, status int, code, msg string) {
	w.Header().Set("Content-Type", "text/xml")
	w.Header().Set("x-amzn-RequestId", id)
	w.WriteHeader(status)
	type er struct {
		XMLName xml.Name `xml:"ErrorResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
		Error   struct {
			Type    string `xml:"Type"`
			Code    string `xml:"Code"`
			Message string `xml:"Message"`
		} `xml:"Error"`
	}
	x := er{Xmlns: ns}
	x.Error.Type = "Sender"
	x.Error.Code = code
	x.Error.Message = msg
	_ = xml.NewEncoder(w).Encode(x)
}

type response struct {
	XMLName xml.Name `xml:"Response"`
	Xmlns   string   `xml:"xmlns,attr"`
	Result  any      `xml:"-"`
}

func (h *Handler) create(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	n := in["Name"]
	if !nameRE.MatchString(n) || strings.HasSuffix(n, ".fifo") || in["Attributes.entry.1.key"] == "FifoTopic" {
		h.fail(w, id, 400, "InvalidParameter", "invalid topic name")
		return
	}
	arn := fmt.Sprintf("arn:aws:sns:%s:%s:%s", h.region, account, n)
	t, e := h.store.CreateSNSTopic(r.HTTPRequest.Context(), state.SNSTopic{Name: n, ARN: arn, DisplayName: n, CreatedAt: time.Now().UTC()})
	if e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	type result struct {
		TopicArn string `xml:"TopicArn"`
	}
	h.write(w, id, "CreateTopicResponse", struct {
		XMLName xml.Name `xml:"CreateTopicResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
		Result  result   `xml:"CreateTopicResult"`
	}{Xmlns: ns, Result: result{t.ARN}})
}
func (h *Handler) list(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	start := in["NextToken"]
	ts, e := h.store.ListSNSTopics(r.HTTPRequest.Context(), start, 101)
	if e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	type member struct {
		Arn string `xml:"TopicArn"`
	}
	members := []member{}
	for _, t := range ts {
		members = append(members, member{t.ARN})
	}
	var token string
	if len(members) > 100 {
		token = ts[99].ARN
		members = members[:100]
	}
	type result struct {
		Topics struct {
			Members []member `xml:"member"`
		} `xml:"Topics"`
		Next string `xml:"NextToken,omitempty"`
	}
	x := result{}
	x.Topics.Members = members
	x.Next = token
	h.write(w, id, "ListTopicsResponse", struct {
		XMLName xml.Name `xml:"ListTopicsResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
		Result  result   `xml:"ListTopicsResult"`
	}{Xmlns: ns, Result: x})
}
func (h *Handler) attrs(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	t, e := h.store.GetSNSTopic(r.HTTPRequest.Context(), in["TopicArn"])
	if errors.Is(e, state.ErrSNSTopicNotFound) {
		h.fail(w, id, 404, "NotFound", "topic not found")
		return
	}
	if e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	type entry struct {
		Key   string `xml:"key"`
		Value string `xml:"value"`
	}
	vals := []entry{{"TopicArn", t.ARN}, {"DisplayName", t.DisplayName}, {"SubscriptionsConfirmed", "0"}, {"SubscriptionsPending", "0"}, {"SubscriptionsDeleted", "0"}}
	type result struct {
		Attrs struct {
			Entries []entry `xml:"entry"`
		} `xml:"Attributes"`
	}
	x := result{}
	x.Attrs.Entries = vals
	h.write(w, id, "GetTopicAttributesResponse", struct {
		XMLName xml.Name `xml:"GetTopicAttributesResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
		Result  result   `xml:"GetTopicAttributesResult"`
	}{Xmlns: ns, Result: x})
}
func (h *Handler) delete(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	if e := h.store.DeleteSNSTopic(r.HTTPRequest.Context(), in["TopicArn"]); e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	h.write(w, id, "DeleteTopicResponse", struct {
		XMLName xml.Name `xml:"DeleteTopicResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
	}{Xmlns: ns})
}
func (h *Handler) publish(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	if in["Message"] == "" || !utf8.ValidString(in["Message"]) || len([]byte(in["Message"])) > 262144 {
		h.fail(w, id, 400, "InvalidParameter", "invalid message")
		return
	}
	if in["MessageStructure"] != "" || in["Subject"] != "" && strings.ContainsAny(in["Subject"], "\r\n") {
		h.fail(w, id, 400, "InvalidParameter", "unsupported message options")
		return
	}
	topic, e := h.store.GetSNSTopic(r.HTTPRequest.Context(), in["TopicArn"])
	if errors.Is(e, state.ErrSNSTopicNotFound) {
		h.fail(w, id, 404, "NotFound", "topic not found")
		return
	} else if e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	type result struct {
		MessageId string `xml:"MessageId"`
	}
	msgID := hex.EncodeToString(b[:])
	subs, _ := h.store.ListSNSSubscriptions(r.HTTPRequest.Context(), topic.ARN)
	for _, sub := range subs {
		if sub.Protocol != "sqs" {
			continue
		}
		body := in["Message"]
		if !sub.RawDelivery {
			env := map[string]any{"Type": "Notification", "MessageId": msgID, "TopicArn": topic.ARN, "Message": in["Message"], "Timestamp": time.Now().UTC().Format(time.RFC3339)}
			if in["Subject"] != "" {
				env["Subject"] = in["Subject"]
			}
			bodyBytes, _ := json.Marshal(env)
			body = string(bodyBytes)
		}
		qname := strings.TrimPrefix(sub.Endpoint, "arn:aws:sqs:")
		parts := strings.SplitN(qname, ":", 4)
		if len(parts) == 4 {
			_, _ = h.store.SendSQSMessage(r.HTTPRequest.Context(), parts[3], body, time.Now().UTC())
		}
	}
	h.write(w, id, "PublishResponse", struct {
		XMLName xml.Name `xml:"PublishResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
		Result  result   `xml:"PublishResult"`
	}{Xmlns: ns, Result: result{msgID}})
}

func (h *Handler) subscribe(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	if in["Protocol"] != "sqs" {
		h.fail(w, id, 400, "InvalidParameter", "only sqs protocol supported")
		return
	}
	if _, e := h.store.GetSNSTopic(r.HTTPRequest.Context(), in["TopicArn"]); e != nil {
		h.fail(w, id, 404, "NotFound", "topic not found")
		return
	}
	if !strings.HasPrefix(in["Endpoint"], "arn:aws:sqs:") {
		h.fail(w, id, 400, "InvalidParameter", "invalid queue endpoint")
		return
	}
	p := strings.SplitN(strings.TrimPrefix(in["Endpoint"], "arn:aws:sqs:"), ":", 4)
	if len(p) != 4 {
		h.fail(w, id, 400, "InvalidParameter", "invalid queue endpoint")
		return
	}
	if _, e := h.store.GetSQSQueue(r.HTTPRequest.Context(), p[3]); e != nil {
		h.fail(w, id, 404, "NotFound", "queue not found")
		return
	}
	sid := fmt.Sprintf("arn:aws:sns:%s:%s:%s", h.region, account, hex.EncodeToString([]byte(in["TopicArn"] + in["Endpoint"]))[:32])
	v, e := h.store.CreateSNSSubscription(r.HTTPRequest.Context(), state.SNSSubscription{ID: sid, TopicARN: in["TopicArn"], Protocol: "sqs", Endpoint: in["Endpoint"], CreatedAt: time.Now().UTC()})
	if e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	h.write(w, id, "SubscribeResponse", struct {
		XMLName xml.Name `xml:"SubscribeResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
		Result  struct {
			ARN string `xml:"SubscriptionArn"`
		} `xml:"SubscribeResult"`
	}{Xmlns: ns, Result: struct {
		ARN string `xml:"SubscriptionArn"`
	}{v.ID}})
}
func (h *Handler) unsubscribe(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	_ = h.store.DeleteSNSSubscription(r.HTTPRequest.Context(), in["SubscriptionArn"])
	h.write(w, id, "UnsubscribeResponse", struct {
		XMLName xml.Name `xml:"UnsubscribeResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
	}{Xmlns: ns})
}
func (h *Handler) listSubs(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string, byTopic bool) {
	topic := ""
	if byTopic {
		topic = in["TopicArn"]
	}
	ss, e := h.store.ListSNSSubscriptions(r.HTTPRequest.Context(), topic)
	if e != nil {
		h.fail(w, id, 500, "InternalError", e.Error())
		return
	}
	type m struct{ Arn, Owner, Protocol, Endpoint, Topic string }
	var ms []m
	for _, s := range ss {
		ms = append(ms, m{s.ID, account, s.Protocol, s.Endpoint, s.TopicARN})
	}
	h.write(w, id, "ListSubscriptionsResponse", struct {
		XMLName xml.Name `xml:"ListSubscriptionsResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
		Result  struct {
			Members []m `xml:"Subscriptions>member"`
		} `xml:"ListSubscriptionsResult"`
	}{Xmlns: ns, Result: struct {
		Members []m `xml:"Subscriptions>member"`
	}{ms}})
}
func (h *Handler) subAttrs(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	s, e := h.store.GetSNSSubscription(r.HTTPRequest.Context(), in["SubscriptionArn"])
	if e != nil {
		h.fail(w, id, 404, "NotFound", "subscription not found")
		return
	}
	_ = s
	h.write(w, id, "GetSubscriptionAttributesResponse", struct {
		XMLName xml.Name `xml:"GetSubscriptionAttributesResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
	}{Xmlns: ns})
}
func (h *Handler) setSubAttrs(w http.ResponseWriter, r *awsprovider.Request, id string, in map[string]string) {
	if in["AttributeName"] != "RawMessageDelivery" {
		h.fail(w, id, 400, "InvalidParameter", "unsupported attribute")
		return
	}
	e := h.store.SetSNSRawDelivery(r.HTTPRequest.Context(), in["SubscriptionArn"], in["AttributeValue"] == "true")
	if e != nil {
		h.fail(w, id, 404, "NotFound", "subscription not found")
		return
	}
	h.write(w, id, "SetSubscriptionAttributesResponse", struct {
		XMLName xml.Name `xml:"SetSubscriptionAttributesResponse"`
		Xmlns   string   `xml:"xmlns,attr"`
	}{Xmlns: ns})
}
