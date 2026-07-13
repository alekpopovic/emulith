package main

import (
	"flag"
	"fmt"
	"github.com/alekpopovic/emulith/internal/compatibility"
	"os"
	"runtime"
	"time"
)

func main() {
	catalog := flag.String("catalog", "compatibility/aws.yaml", "")
	results := flag.String("results", "build/compatibility/results.json", "")
	jsonOut := flag.String("json", "build/compatibility/aws.json", "")
	mdOut := flag.String("markdown", "docs/compatibility/aws.md", "")
	check := flag.Bool("check", false, "")
	flag.Parse()
	if err := run(*catalog, *results, *jsonOut, *mdOut, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(cp, rp, jp, mp string, check bool) error {
	c, e := compatibility.Load(cp)
	if e != nil {
		return e
	}
	r, e := compatibility.ReadResults(rp)
	if e != nil {
		return e
	}
	compatibility.Sort(&c, r)
	if e = compatibility.Validate(c, r, true); e != nil {
		return e
	}
	md := []byte(compatibility.Markdown(c))
	if check {
		old, e := os.ReadFile(mp)
		if e != nil {
			return e
		}
		if string(old) != string(md) {
			return compatibility.ErrStale
		}
		return nil
	}
	summary := map[string]int{}
	for _, x := range c.Entries {
		summary[x.Status]++
	}
	rep := compatibility.Report{SchemaVersion: 1, EmulithVersion: env("VERSION", "dev"), Commit: env("COMMIT", "unknown"), GoVersion: runtime.Version(), SDKModules: map[string]string{"github.com/aws/aws-sdk-go-v2": "v1.30.5", "github.com/aws/aws-sdk-go-v2/service/dynamodb": "v1.34.5"}, GeneratedAt: time.Now().UTC().Format(time.RFC3339), Entries: c.Entries, Results: r, Summary: summary}
	b, e := compatibility.JSON(rep)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(dir(jp), 0755); e != nil {
		return e
	}
	if e = os.WriteFile(jp, b, 0644); e != nil {
		return e
	}
	return os.WriteFile(mp, md, 0644)
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func dir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}
