package compatibility

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var statuses = map[string]bool{"supported": true, "partial": true, "experimental": true, "unsupported": true}

type Catalog struct {
	Entries []Entry `yaml:"entries" json:"entries"`
}
type Entry struct {
	Provider        string   `yaml:"provider" json:"provider"`
	Service         string   `yaml:"service" json:"service"`
	Operation       string   `yaml:"operation" json:"operation"`
	Status          string   `yaml:"status" json:"status"`
	Protocol        string   `yaml:"protocol" json:"protocol"`
	TestID          string   `yaml:"test_id" json:"test_id"`
	Notes           string   `yaml:"notes" json:"notes"`
	KnownDeviations []string `yaml:"known_deviations" json:"known_deviations"`
	Since           string   `yaml:"since" json:"since"`
}
type Result struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
type Results struct {
	Results []Result `json:"results"`
}
type Report struct {
	SchemaVersion  int               `json:"schema_version"`
	EmulithVersion string            `json:"emulith_version"`
	Commit         string            `json:"commit"`
	GoVersion      string            `json:"go_version"`
	SDKModules     map[string]string `json:"sdk_modules"`
	GeneratedAt    string            `json:"generated_at"`
	Entries        []Entry           `json:"catalog_entries"`
	Results        []Result          `json:"test_results"`
	Summary        map[string]int    `json:"summary"`
}

func Load(path string) (Catalog, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, err
	}
	var c Catalog
	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	if err = d.Decode(&c); err != nil {
		return c, err
	}
	return c, Validate(c, nil, false)
}
func Validate(c Catalog, results []Result, requirePassing bool) error {
	seen, tests := map[string]bool{}, map[string]string{}
	for _, e := range c.Entries {
		key := e.Provider + "/" + e.Service + "/" + e.Operation
		if seen[key] {
			return fmt.Errorf("duplicate operation %s", key)
		}
		seen[key] = true
		if !statuses[e.Status] {
			return fmt.Errorf("invalid status %q", e.Status)
		}
		if e.Provider == "" || e.Service == "" || e.Operation == "" || e.Protocol == "" || e.Notes == "" || e.Since == "" {
			return fmt.Errorf("incomplete catalog entry %s", key)
		}
		if e.Status == "supported" && e.TestID == "" {
			return fmt.Errorf("supported entry %s lacks test_id", key)
		}
		if e.TestID != "" {
			if tests[e.TestID] != "" {
				return fmt.Errorf("duplicate test_id %s", e.TestID)
			}
			tests[e.TestID] = key
		}
	}
	if results == nil {
		return nil
	}
	executed := map[string]string{}
	for _, r := range results {
		if _, ok := executed[r.ID]; ok {
			return fmt.Errorf("duplicate executed test %s", r.ID)
		}
		executed[r.ID] = r.Status
		if tests[r.ID] == "" {
			return fmt.Errorf("executed compatibility test %s missing from catalog", r.ID)
		}
	}
	for id, key := range tests {
		if _, ok := executed[id]; !ok {
			return fmt.Errorf("catalog test %s for %s does not exist in results", id, key)
		}
	}
	if requirePassing {
		for _, e := range c.Entries {
			if e.Status == "supported" && executed[e.TestID] != "pass" {
				return fmt.Errorf("supported test %s is %s", e.TestID, executed[e.TestID])
			}
		}
	}
	return nil
}
func Sort(c *Catalog, r []Result) {
	sort.Slice(c.Entries, func(i, j int) bool {
		a, b := c.Entries[i], c.Entries[j]
		return a.Provider+a.Service+a.Operation < b.Provider+b.Service+b.Operation
	})
	sort.Slice(r, func(i, j int) bool { return r[i].ID < r[j].ID })
}
func Markdown(c Catalog) string {
	var b strings.Builder
	b.WriteString("# AWS compatibility\n\nGenerated from `compatibility/aws.yaml`. Statuses: supported (default SDK test passes), partial (documented subset), experimental (may change), unsupported (not implemented).\n\n| Service | Operation | Status | Protocol | Test ID | Notes | Known deviations | Since |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, e := range c.Entries {
		vals := []string{e.Service, e.Operation, e.Status, e.Protocol, e.TestID, e.Notes, strings.Join(e.KnownDeviations, "; "), e.Since}
		for i := range vals {
			vals[i] = strings.ReplaceAll(strings.ReplaceAll(vals[i], "|", "\\|"), "\n", " ")
		}
		fmt.Fprintf(&b, "| %s |\n", strings.Join(vals, " | "))
	}
	return b.String()
}
func ReadResults(path string) ([]Result, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var r Results
	if e = json.Unmarshal(b, &r); e != nil {
		return nil, e
	}
	return r.Results, nil
}
func JSON(r Report) ([]byte, error) {
	b, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return nil, e
	}
	return append(b, '\n'), nil
}

var ErrStale = errors.New("generated compatibility document is stale")
