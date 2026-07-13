package common

import (
	"fmt"
	"regexp"
	"strings"
)

var projectRE = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)

func ValidProject(p string) bool { return projectRE.MatchString(p) }
func ParseProjectResource(s string) (string, string, error) {
	p := strings.Split(s, "/")
	if len(p) != 2 || p[0] != "projects" || !ValidProject(p[1]) {
		return "", "", fmt.Errorf("invalid project resource")
	}
	return p[1], "", nil
}
func Resource(project, kind, id string) (string, error) {
	if !ValidProject(project) || id == "" || strings.Contains(id, "/") {
		return "", fmt.Errorf("invalid resource")
	}
	return fmt.Sprintf("projects/%s/%s/%s", project, kind, id), nil
}
