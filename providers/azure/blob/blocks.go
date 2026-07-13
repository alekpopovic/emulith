// Package blob contains bounded, provider-independent helpers for Azure Blob
// block staging and HTTP range/condition handling.
package blob

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidBlockID = errors.New("invalid block id")
var ErrInvalidRange = errors.New("invalid range")

// Block is an in-memory representation suitable for persistence by the blob state layer.
type Block struct {
	ID        string
	Data      []byte
	Committed bool
	Order     int
}

// CanonicalBlockID validates and canonicalizes a base64 block identifier.
func CanonicalBlockID(raw string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(b) == 0 || len(b) > 64 {
		return "", ErrInvalidBlockID
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// CommitBlocks resolves a block list atomically from staged blocks and returns the body.
func CommitBlocks(staged map[string][]byte, ids []string, max int64) ([]byte, error) {
	if len(ids) == 0 || len(ids) > 50000 {
		return nil, fmt.Errorf("invalid block list")
	}
	var n int64
	var out bytes.Buffer
	for _, id := range ids {
		c, err := CanonicalBlockID(id)
		if err != nil {
			return nil, err
		}
		b, ok := staged[c]
		if !ok {
			return nil, fmt.Errorf("missing block")
		}
		n += int64(len(b))
		if n > max {
			return nil, fmt.Errorf("blob too large")
		}
		_, _ = out.Write(b)
	}
	return out.Bytes(), nil
}

// StreamBlocks writes a resolved block list in order without buffering the
// resulting blob. It is the preferred primitive for request handlers.
func StreamBlocks(staged map[string][]byte, ids []string, max int64, dst io.Writer) (int64, error) {
	if len(ids) == 0 || len(ids) > 50000 {
		return 0, fmt.Errorf("invalid block list")
	}
	var n int64
	for _, id := range ids {
		c, err := CanonicalBlockID(id)
		if err != nil {
			return n, err
		}
		b, ok := staged[c]
		if !ok {
			return n, fmt.Errorf("missing block")
		}
		n += int64(len(b))
		if n > max {
			return n, fmt.Errorf("blob too large")
		}
		if _, err = dst.Write(b); err != nil {
			return n, err
		}
	}
	return n, nil
}

type Range struct{ Start, End, Length int64 }

// ParseRange parses one Azure/HTTP bytes range against size.
func ParseRange(value string, size int64) (Range, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "bytes=") || strings.Count(value, ",") > 0 || size < 0 {
		return Range{}, ErrInvalidRange
	}
	p := strings.Split(strings.TrimSpace(strings.TrimPrefix(value, "bytes=")), "-")
	if len(p) != 2 {
		return Range{}, ErrInvalidRange
	}
	if p[0] == "" {
		n, e := strconv.ParseInt(p[1], 10, 64)
		if e != nil || n <= 0 || size == 0 {
			return Range{}, ErrInvalidRange
		}
		if n > size {
			n = size
		}
		return Range{size - n, size - 1, n}, nil
	}
	start, e := strconv.ParseInt(p[0], 10, 64)
	if e != nil || start < 0 || start >= size {
		return Range{}, ErrInvalidRange
	}
	end := size - 1
	if p[1] != "" {
		end, e = strconv.ParseInt(p[1], 10, 64)
		if e != nil || end < start {
			return Range{}, ErrInvalidRange
		}
		if end >= size {
			end = size - 1
		}
	}
	return Range{start, end, end - start + 1}, nil
}

func (r Range) ContentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.Start, r.End, size)
}
func (r Range) Reader(src io.ReaderAt) io.Reader { return io.NewSectionReader(src, r.Start, r.Length) }

// Conditions evaluates HTTP preconditions. A non-empty If-Match/If-Unmodified-Since
// failure is 412; If-None-Match/If-Modified-Since failure is 304 for GET/HEAD.
func Conditions(req *http.Request, etag string, modified time.Time) int {
	if v := req.Header.Get("If-Match"); v != "" && v != "*" && !etagListMatch(v, etag) {
		return http.StatusPreconditionFailed
	}
	if v := req.Header.Get("If-Unmodified-Since"); v != "" {
		if t, e := http.ParseTime(v); e == nil && modified.Truncate(time.Second).After(t.Truncate(time.Second)) {
			return http.StatusPreconditionFailed
		}
	}
	if v := req.Header.Get("If-None-Match"); v != "" && (v == "*" || etagListMatch(v, etag)) {
		if req.Method == http.MethodGet || req.Method == http.MethodHead {
			return http.StatusNotModified
		}
		return http.StatusPreconditionFailed
	}
	if v := req.Header.Get("If-Modified-Since"); v != "" && req.Header.Get("If-None-Match") == "" {
		if t, e := http.ParseTime(v); e == nil && !modified.Truncate(time.Second).After(t.Truncate(time.Second)) {
			return http.StatusNotModified
		}
	}
	return 0
}
func etagListMatch(v, etag string) bool {
	for _, x := range strings.Split(v, ",") {
		if strings.Trim(strings.TrimSpace(x), "\"") == strings.Trim(etag, "\"") {
			return true
		}
	}
	return false
}
func Hash(data []byte) [32]byte { return sha256.Sum256(data) }
