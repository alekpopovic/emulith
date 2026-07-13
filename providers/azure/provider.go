package azure

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"net/http"
	"strings"
	"time"
)

const DefaultAccountName = "devstoreaccount1"
const DefaultAccountKey = "RW11bGl0aC1kZXZlbG9wbWVudC1rZXk="

type Handler struct {
	Service string
	Account string
}

func New(service, account string) *Handler {
	if account == "" {
		account = DefaultAccountName
	}
	return &Handler{Service: service, Account: account}
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestID()
	w.Header().Set("x-ms-request-id", id)
	w.Header().Set("x-ms-version", "2023-11-03")
	w.Header().Set("Date", httpTime())
	account := strings.Trim(strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)[0], " ")
	if account != h.Account {
		writeError(w, 404, "AccountNotFound", "The specified account does not exist.", id)
		return
	}
	writeError(w, 501, "UnsupportedOperation", "This Azure Storage operation is not implemented in the local POC.", id)
}
func requestID() string {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return "local-request"
	}
	return hex.EncodeToString(b)
}
func httpTime() string { return time.Now().UTC().Format(http.TimeFormat) }

type errorBody struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId"`
}

func writeError(w http.ResponseWriter, status int, code, msg, id string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(errorBody{Code: code, Message: msg, RequestID: id})
}
