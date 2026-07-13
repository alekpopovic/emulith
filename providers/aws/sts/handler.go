package sts

import (
	"encoding/xml"
	"net/http"

	awsprovider "github.com/alekpopovic/emulith/providers/aws"
)

const namespace = "https://sts.amazonaws.com/doc/2011-06-15/"

type Handler struct{}

func New() *Handler { return &Handler{} }

type callerIdentityResponse struct {
	XMLName  xml.Name             `xml:"GetCallerIdentityResponse"`
	XMLNS    string               `xml:"xmlns,attr"`
	Result   callerIdentityResult `xml:"GetCallerIdentityResult"`
	Metadata responseMetadata     `xml:"ResponseMetadata"`
}
type callerIdentityResult struct {
	UserID  string `xml:"UserId"`
	Account string `xml:"Account"`
	ARN     string `xml:"Arn"`
}
type responseMetadata struct {
	RequestID string `xml:"RequestId"`
}
type errorResponse struct {
	XMLName   xml.Name   `xml:"ErrorResponse"`
	XMLNS     string     `xml:"xmlns,attr"`
	Error     queryError `xml:"Error"`
	RequestID string     `xml:"RequestId"`
}
type queryError struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

func (h *Handler) ServeAWS(w http.ResponseWriter, req *awsprovider.Request, requestID string) {
	if req.Protocol != awsprovider.ProtocolQuery || req.Operation != "GetCallerIdentity" {
		writeError(w, requestID, http.StatusBadRequest, "InvalidAction", "The action is not valid for this endpoint")
		return
	}
	if version := req.Form.Get("Version"); version != "" && version != "2011-06-15" {
		writeError(w, requestID, http.StatusBadRequest, "InvalidParameterValue", "Unsupported STS API version")
		return
	}
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("x-amzn-RequestId", requestID)
	w.WriteHeader(http.StatusOK)
	_ = xml.NewEncoder(w).Encode(callerIdentityResponse{XMLNS: namespace, Result: callerIdentityResult{UserID: "EMULITHUSER", Account: "000000000000", ARN: "arn:aws:iam::000000000000:user/emulith"}, Metadata: responseMetadata{RequestID: requestID}})
}

func writeError(w http.ResponseWriter, requestID string, status int, code, message string) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("x-amzn-RequestId", requestID)
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(errorResponse{XMLNS: namespace, Error: queryError{Type: "Sender", Code: code, Message: message}, RequestID: requestID})
}
