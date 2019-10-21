package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/narqo/go-badge"
)

var (
	isGitHub = func(p Project) bool { return p.Hoster == "github.com" }
	isGitLab = func(p Project) bool { return p.Hoster == "gitlab.com" }
	isAzure  = func(p Project) bool { return p.AzureDefinitionID != "" }
	cells    = map[string]Cell{
		"icon":         &Icon{},
		"link":         &Link{},
		"travis":       &Badge{"https://travis-ci.org/{{.Namespace}}/{{.Name}}.svg?branch=master", "https://travis-ci.org/{{.Namespace}}/{{.Name}}", nil},
		"gocover":      &Badge{"http://gocover.io/_badge/{{.Hoster}}/{{.Namespace}}/{{.Name}}", "https://gocover.io/{{.Hoster}}/{{.Namespace}}/{{.Name}}", nil},
		"codecov":      &Badge{"https://codecov.io/gh/{{.Namespace}}/{{.Name}}/branch/master/graph/badge.svg", "https://codecov.io/gh/{{.Namespace}}/{{.Name}}", nil},
		"goreportcard": &Badge{"https://goreportcard.com/badge/{{.GoImportPath}}", "https://goreportcard.com/report/{{.GoImportPath}}", nil},
		"golangci":     &Badge{"https://img.shields.io/badge/-golangci--lint-47cad6", "https://golangci.com/r/{{.Hoster}}/{{.Namespace}}/{{.Name}}", nil},
		"godoc":        &Badge{"https://godoc.org/{{.GoImportPath}}?status.svg", "https://godoc.org/{{.GoImportPath}}", nil},

		"azure-pipeline": &Badge{"https://img.shields.io/azure-devops/build/cugu/dfir/{{.AzureDefinitionID}}", "https://dev.azure.com/cugu/dfir/_build?definitionId={{.AzureDefinitionID}}&_a=summary", isAzure},
		"azure-coverage": &Badge{"https://img.shields.io/azure-devops/coverage/cugu/dfir/{{.AzureDefinitionID}}", "{{.URL}}", isAzure},

		"github-branches":     &GithubBranches{},
		"github-forks":        &GithubProject{"forks"},                                                                            // &Badge{"https://img.shields.io/github/forks/{{.Namespace}}/{{.Name}}?label=Fork", "{{.URL}}/network", isGitHub},
		"github-issues":       &GithubProject{"issues"},                                                                           // &Badge{"https://img.shields.io/github/issues/{{.Namespace}}/{{.Name}}", "{{.URL}}/issues", isGitHub},
		"github-lastcommit":   &Badge{"https://img.shields.io/github/last-commit/{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub}, // &GithubProject{"lastcommit"}
		"github-license":      &GithubProject{"license"},                                                                          // &Badge{"https://img.shields.io/github/license/{{.Namespace}}/{{.Name}}", "{{.URL}}/blob/master/LICENSE", isGitHub},
		"github-newcommits":   &Badge{"https://img.shields.io/github/commits-since/{{.Namespace}}/{{.Name}}/latest", "{{.URL}}", isGitHub},
		"github-pipeline":     &Badge{"{{.URL}}/workflows/{{.Workflow}}/badge.svg", "{{.URL}}/actions", isGitHub},
		"github-pullrequests": &GithubMergeRequests{},  // &Badge{"https://img.shields.io/github/issues-pr/{{.Namespace}}/{{.Name}}", "{{.URL}}/pulls", isGitHub},
		"github-size":         &GithubProject{"size"},  // &Badge{"https://img.shields.io/github/repo-size/{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub},
		"github-stars":        &GithubProject{"stars"}, // &Badge{"https://img.shields.io/github/stars/{{.Namespace}}/{{.Name}}", "{{.URL}}/stargazers", isGitHub},
		"github-version":      &GithubTag{},            // &Badge{"https://img.shields.io/github/v/tag/{{.Namespace}}/{{.Name}}?sort=semver", "{{.URL}}", isGitHub},
		"github-visibility":   &GithubProject{"visibility"},
		"github-watchers":     &Badge{"https://img.shields.io/github/watchers/{{.Namespace}}/{{.Name}}?label=Watch", "{{.URL}}/watchers", isGitHub}, //  &GithubProject{"watchers"},

		"gitlab-branches":      &GitLabBranches{},
		"gitlab-coverage":      &Badge{"{{.URL}}/badges/master/coverage.svg", "{{.URL}}/-/jobs/artifacts/master/file/coverage.html?job=unittests", isGitLab},
		"gitlab-forks":         &GitLabProject{"forks"},
		"gitlab-issues":        &GitLabProject{"issues"},
		"gitlab-lastcommit":    &GitLabProject{"lastcommit"},
		"gitlab-mergerequests": &GitLabMergeRequests{},
		"gitlab-pipeline":      &Badge{"{{.URL}}/badges/master/pipeline.svg", "{{.URL}}/pipelines", isGitLab},
		"gitlab-size":          &GitLabProject{"size"},
		"gitlab-stars":         &GitLabProject{"stars"},
		"gitlab-version":       &GitLabTag{},
		"gitlab-visibility":    &GitLabProject{"visibility"},
	}
)

type Cell interface {
	Render(column string, p Project) string
}

type Icon struct{}

func (b *Icon) Render(column string, project Project) string {
	switch project.Hoster {
	case "github.com":
		return "![GitHub Icon](style/github.png)"
	case "gitlab.com":
		return "![Gitlab Icon](style/gitlab.png)"
	default:
		return ""
	}
}

type Link struct{}

func (b *Link) Render(column string, project Project) string {
	return "[" + project.Name + "](" + project.URL + ")"
}

type Badge struct {
	badge     string
	link      string
	condition func(p Project) bool
}

func (b *Badge) Render(column string, project Project) string {
	buf := &bytes.Buffer{}
	if b.condition == nil || b.condition(project) {
		buf.WriteString("[![" + column + "](")
		badgeurl := template.Must(template.New("badge").Parse(b.badge))
		badgeurl.Execute(buf, project)
		buf.WriteString(")](")
		linkurl := template.Must(template.New("link").Parse(b.link))
		linkurl.Execute(buf, project)
		buf.WriteString(")")
	}
	return buf.String()
}

func svgBadge(projectname, name, left, right string, color badge.Color, url string) string {
	b, err := badge.RenderBytes(left, right, color)
	if err != nil {
		panic(err)
	}
	os.MkdirAll(filepath.Join("badges", projectname), 0777)
	ioutil.WriteFile(filepath.Join("badges", projectname, name+".svg"), bytes.ReplaceAll(b, []byte("\n"), []byte("")), 0666)

	return fmt.Sprintf("[![%s](badges/%s/%s.svg)](%s)", name, projectname, name, url)
	// return fmt.Sprintf("[![%s](https://img.shields.io/badge/%s-%s-%s)](%s)", left, left, right, strings.TrimLeft(string(color), "#"), url)
}
