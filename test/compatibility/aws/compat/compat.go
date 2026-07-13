package compat

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"testing"
)

type Result struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

var mu sync.Mutex
var results []Result
var seen = map[string]bool{}

func Run(t *testing.T, id string, fn func(t *testing.T)) {
	t.Helper()
	mu.Lock()
	if seen[id] {
		mu.Unlock()
		t.Fatalf("duplicate compatibility ID %s", id)
	}
	seen[id] = true
	mu.Unlock()
	ok := t.Run(id, fn)
	status := "fail"
	if ok {
		status = "pass"
	}
	mu.Lock()
	results = append(results, Result{id, status})
	mu.Unlock()
}
func Write() {
	path := os.Getenv("EMULITH_COMPAT_RESULTS")
	if path == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	b, _ := json.MarshalIndent(struct {
		Results []Result `json:"results"`
	}{results}, "", "  ")
	b = append(b, '\n')
	_ = os.WriteFile(path, b, 0644)
}
