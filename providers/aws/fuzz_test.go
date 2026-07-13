package aws

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func FuzzClassify(f *testing.F) {
	for _, seed := range []string{"Action=GetCallerIdentity", "Action=CreateQueue", "list-type=2", "%zz", strings.Repeat("a", 1024)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, query string) {
		if len(query) > 4096 {
			t.Skip()
		}
		r := httptest.NewRequest("GET", "/?"+query, nil)
		_, _ = classify(r)
	})
}
