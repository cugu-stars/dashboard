package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/narqo/go-badge"
)

var (
	cells    = map[string]func(Project) string{}
	isGitHub = func(p Project) bool { return p.Hoster == "github.com" }
	isGitLab = func(p Project) bool { return p.Hoster == "gitlab.com" }
)

func setup() {
	isAzure := func(p Project) bool { return p.AzureDefinitionID != "" }
	githubProject := NewGithubProject()
	gitlabProject := NewGitLabProject()
	cells = map[string]func(Project) string{
		"icon":         icon,
		"link":         link,
		"travis":       markdownBadge("https://travis-ci.org/{{.Namespace}}/{{.Name}}.svg?branch=master", "https://travis-ci.org/{{.Namespace}}/{{.Name}}", nil),
		"gocover":      markdownBadge("http://gocover.io/_badge/{{.Hoster}}/{{.Namespace}}/{{.Name}}", "https://gocover.io/{{.Hoster}}/{{.Namespace}}/{{.Name}}", nil),
		"codecov":      markdownBadge("https://codecov.io/gh/{{.Namespace}}/{{.Name}}/branch/master/graph/badge.svg", "https://codecov.io/gh/{{.Namespace}}/{{.Name}}", nil),
		"goreportcard": markdownBadge("https://goreportcard.com/badge/{{.GoImportPath}}", "https://goreportcard.com/report/{{.GoImportPath}}", nil),
		"golangci":     markdownBadge("https://img.shields.io/badge/-golangci--lint-47cad6", "https://golangci.com/r/{{.Hoster}}/{{.Namespace}}/{{.Name}}", nil),
		"godoc":        markdownBadge("https://godoc.org/{{.GoImportPath}}?status.svg", "https://godoc.org/{{.GoImportPath}}", nil),

		"azure-pipeline": markdownBadge("https://img.shields.io/azure-devops/build/{{.AzureOrganization}}/{{.AzureProject}}/{{.AzureDefinitionID}}", "https://dev.azure.com/{{.AzureOrganization}}/{{.AzureProject}}/_build?definitionId={{.AzureDefinitionID}}&_a=summary", isAzure),
		"azure-coverage": markdownBadge("https://img.shields.io/azure-devops/coverage/{{.AzureOrganization}}/{{.AzureProject}}/{{.AzureDefinitionID}}", "{{.URL}}", isAzure),

		"github-branches":     githubProject.branches,
		"github-forks":        githubProject.forks,
		"github-issues":       githubProject.issues,
		"github-lastcommit":   markdownBadge("https://img.shields.io/github/last-commit/{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub),
		"github-license":      githubProject.license,
		"github-newcommits":   githubProject.commitssince,
		"github-pipeline":     markdownBadge("{{.URL}}/workflows/{{.Workflow}}/badge.svg", "{{.URL}}/actions", isGitHub),
		"github-pullrequests": githubProject.pullRequests,
		"github-size":         githubProject.size,
		"github-stars":        githubProject.stars,
		"github-version":      githubProject.tag,
		"github-visibility":   githubProject.visibility,
		"github-watchers":     githubProject.watchers,
		"github-sloc":         markdownBadge("https://sloc.xyz/github/{{.Namespace}}/{{.Name}}/", "{{.URL}}", isGitHub),
		// "github-forks":        markdownBadge("https://img.shields.io/github/forks/{{.Namespace}}/{{.Name}}?label=Fork", "{{.URL}}/network", isGitHub),
		// "github-issues":       markdownBadge("https://img.shields.io/github/issues/{{.Namespace}}/{{.Name}}", "{{.URL}}/issues", isGitHub),
		// "github-lastcommit":   githubProject.lastcommit,
		// "github-license":      markdownBadge("https://img.shields.io/github/license/{{.Namespace}}/{{.Name}}", "{{.URL}}/blob/master/LICENSE", isGitHub),
		// "github-pullrequests": markdownBadge("https://img.shields.io/github/issues-pr/{{.Namespace}}/{{.Name}}", "{{.URL}}/pulls", isGitHub),
		// "github-size":         markdownBadge("https://img.shields.io/github/repo-size/{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub),
		// "github-stars":        markdownBadge("https://img.shields.io/github/stars/{{.Namespace}}/{{.Name}}", "{{.URL}}/stargazers", isGitHub),
		// "github-version":      markdownBadge("https://img.shields.io/github/v/tag/{{.Namespace}}/{{.Name}}?sort=semver", "{{.URL}}", isGitHub),
		// "github-watchers":     markdownBadge("https://img.shields.io/github/watchers/{{.Namespace}}/{{.Name}}?label=Watch", "{{.URL}}/watchers", isGitHub),

		"gitlab-branches":      gitlabProject.branches,
		"gitlab-coverage":      markdownBadge("{{.URL}}/badges/master/coverage.svg", "{{.URL}}/-/jobs/artifacts/master/file/coverage.html?job=unittests", isGitLab),
		"gitlab-forks":         gitlabProject.forks,
		"gitlab-issues":        gitlabProject.issues,
		"gitlab-lastcommit":    gitlabProject.lastcommit,
		"gitlab-mergerequests": gitlabProject.mergerequests,
		"gitlab-pipeline":      markdownBadge("{{.URL}}/badges/master/pipeline.svg", "{{.URL}}/pipelines", isGitLab),
		"gitlab-size":          gitlabProject.size,
		"gitlab-stars":         gitlabProject.stars,
		"gitlab-version":       gitlabProject.tag,
		"gitlab-visibility":    gitlabProject.visibility,
	}
}

func svgBadge(hoster, projectname, name, left, right string, color badge.Color, url string, err error) string {
	if len(right) > 40 {
		right = right[:35]
	}
	b, err := badge.RenderBytes("", right, color)
	if err != nil {
		panic(err)
	}
	os.MkdirAll(filepath.Join("badges", hoster, projectname), 0777)
	ioutil.WriteFile(filepath.Join("badges", hoster, projectname, name+".svg"), bytes.ReplaceAll(b, []byte("\n"), []byte("")), 0666)

	e := ""
	if err != nil {
		e = err.Error()
	}
	return fmt.Sprintf("[![%s](badges/%s/%s/%s.svg)](%s)<!--%s-->", name, hoster, projectname, name, url, e)
	// fmt.Sprintf("[![%s](https://img.shields.io/badge/%s-%s-%s)](%s)", left, left, right, strings.TrimLeft(string(color), "#"), url)
}
