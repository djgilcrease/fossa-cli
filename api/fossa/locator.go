package fossa

import (
	"regexp"
	"strings"

	"github.com/fossas/fossa-cli/pkg"
)

type Locator struct {
	Fetcher  string `json:"fetcher"`
	Project  string `json:"package"`
	Revision string `json:"revision"`
}

func (l Locator) String() string {
	if l.Fetcher != "git" {
		return l.Fetcher + "+" + l.Project + "$" + l.Revision
	}
	return "git+" + NormalizeGitURL(l.Project) + "$" + l.Revision
}

func (l Locator) QueryString() string {
	if l.Fetcher == "go" {
		l.Fetcher = "git"
	}
	return l.String()
}

func NormalizeGitURL(project string) string {
	// Remove fetcher prefix (in case project is derived from splitting a locator on '$')
	noFetcherPrefix := strings.TrimPrefix(project, "git+")

	// Normalize Git URL format
	noGitExtension := strings.TrimSuffix(noFetcherPrefix, ".git")
	handleGitHubSSH := strings.Replace(noGitExtension, "git@github.com:", "github.com/", 1)

	// Remove protocols
	noHTTPPrefix := strings.TrimPrefix(handleGitHubSSH, "http://")
	noHTTPSPrefix := strings.TrimPrefix(noHTTPPrefix, "https://")

	return noHTTPSPrefix
}

func (l Locator) IsResolved() bool {
	return l.Revision != ""
}

func ReadLocator(s string) Locator {
	r := regexp.MustCompile("^(.*?)\\+(.*?)\\$(.*?)$")
	matches := r.FindStringSubmatch(s)
	return Locator{
		Fetcher:  matches[1],
		Project:  matches[2],
		Revision: matches[3],
	}
}

type ImportPath []Locator
type ImportPathString string

func (p ImportPath) String() ImportPathString {
	var out []string
	for _, locator := range p {
		out = append(out, locator.String())
	}
	return ImportPathString(strings.Join(out, " "))
}

func ReadImportPath(s ImportPathString) ImportPath {
	parts := strings.Split(string(s), " ")
	var out []Locator
	for _, part := range parts {
		out = append(out, ReadLocator(part))
	}
	return out
}

func LocatorOf(id pkg.ID) *Locator {
	return &Locator{
		Fetcher:  LocatorType(id.Type),
		Project:  id.Name,
		Revision: id.Revision,
	}
}

func LocatorType(t pkg.Type) string {
	switch t {
	case pkg.Gradle:
		return "mvn"
	case pkg.Ant:
		return "mvn"
	default:
		return t.String()
	}
}
