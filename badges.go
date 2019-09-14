package main

import (
	"bytes"
	"html/template"
)

var (
	isGitHub = func(p Project) bool { return p.Hoster == "github.com" }
	isGitLab = func(p Project) bool { return p.Hoster == "gitlab.com" }
	isAzure  = func(p Project) bool { return p.AzureDefinitionID != "" }
	cells    = map[string]Cell{
		"icon":                &Icon{},
		"link":                &Link{},
		"azure":               &Badge{"https://img.shields.io/azure-devops/build/cugu/dfir/{{.AzureDefinitionID}}", "https://dev.azure.com/cugu/dfir/_build?definitionId={{.AzureDefinitionID}}&_a=summary", isAzure},
		"gitlab":              &Badge{"{{.URL}}/badges/master/pipeline.svg", "{{.URL}}/pipelines", isGitLab},
		"travis":              &Badge{"https://img.shields.io/travis{{.Namespace}}/{{.Name}}", "https://travis-ci.org{{.Namespace}}/{{.Name}}", nil},
		"azure-coverage":      &Badge{"https://img.shields.io/azure-devops/coverage/cugu/dfir/{{.AzureDefinitionID}}", "{{.URL}}", isAzure},
		"gitlab-coverage":     &Badge{"{{.URL}}/badges/master/coverage.svg", "{{.URL}}/-/jobs/artifacts/master/file/coverage.html?job=unittests", isGitLab},
		"gocover":             &Badge{"http://gocover.io/_badge/{{.Hoster}}{{.Namespace}}/{{.Name}}", "https://gocover.io/{{.Hoster}}{{.Namespace}}/{{.Name}}", nil},
		"codecov":             &Badge{"https://codecov.io/gh{{.Namespace}}/{{.Name}}/branch/master/graph/badge.svg", "https://codecov.io/gh{{.Namespace}}/{{.Name}}", nil},
		"goreportcard":        &Badge{"https://goreportcard.com/badge/{{.GoImportPath}}", "https://goreportcard.com/report/{{.GoImportPath}}", nil},
		"golangci":            &Badge{"https://img.shields.io/badge/-golangci--lint-47cad6", "https://golangci.com/r/{{.Hoster}}{{.Namespace}}/{{.Name}}", nil},
		"godoc":               &Badge{"https://godoc.org/{{.GoImportPath}}?status.svg", "https://godoc.org/{{.GoImportPath}}", nil},
		"github-issues":       &Badge{"https://img.shields.io/github/issues{{.Namespace}}/{{.Name}}", "{{.URL}}/issues", isGitHub},
		"github-pullrequests": &Badge{"https://img.shields.io/github/issues-pr{{.Namespace}}/{{.Name}}", "{{.URL}}/pulls", isGitHub},
		"github-version":      &Badge{"https://img.shields.io/github/v/tag{{.Namespace}}/{{.Name}}?sort=semver", "{{.URL}}", isGitHub},
		"github-lastcommit":   &Badge{"https://img.shields.io/github/last-commit{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub},
		"github-newcommits":   &Badge{"https://img.shields.io/github/commits-since{{.Namespace}}/{{.Name}}/latest", "{{.URL}}", isGitHub},
		"github-watchers":     &Badge{"https://img.shields.io/github/watchers{{.Namespace}}/{{.Name}}?label=Watch", "{{.URL}}/watchers", isGitHub},
		"github-stars":        &Badge{"https://img.shields.io/github/stars{{.Namespace}}/{{.Name}}", "{{.URL}}/stargazers", isGitHub},
		"github-forks":        &Badge{"https://img.shields.io/github/forks{{.Namespace}}/{{.Name}}?label=Fork", "{{.URL}}/network", isGitHub},
		"github-size":         &Badge{"https://img.shields.io/github/repo-size{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub},
		"github-license":      &Badge{"https://img.shields.io/github/license{{.Namespace}}/{{.Name}}", "{{.URL}}/blob/master/LICENSE", isGitHub},
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
