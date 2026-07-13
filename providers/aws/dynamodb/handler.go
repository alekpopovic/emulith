package dynamodb

import (
	"encoding/json"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"net/http"
)

type Handler struct{}

func New() *Handler { return &Handler{} }
func (h *Handler) ServeAWS(w http.ResponseWriter, req *awsprovider.Request, id string) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", id)
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"__type": "com.amazonaws.dynamodb.v20120810#UnknownOperationException", "message": "Operation " + req.Operation + " is not implemented"})
}
