package badge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/enfipy/locker"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/v28/github"
	"github.com/narqo/go-badge"
	"golang.org/x/oauth2"
)

var isGitHub = func(p Project) bool { return p.Hoster == "github.com" }

func InitGitHubBadges(githubAccessToken string) {
	githubProject := NewGithubProject(githubAccessToken)
	badges["github-branches"] = githubProject.branches
	badges["github-forks"] = githubProject.forks
	badges["github-issues"] = githubProject.issues
	badges["github-lastcommit"] = markdownBadge("https://img.shields.io/github/last-commit/{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub)
	badges["github-license"] = githubProject.license
	badges["github-newcommits"] = githubProject.commitssince
	badges["github-pipeline"] = markdownBadge("{{.URL}}/workflows/{{.Workflow}}/badge.svg", "{{.URL}}/actions", isGitHub)
	badges["github-pullrequests"] = githubProject.pullRequests
	badges["github-size"] = githubProject.size
	badges["github-stars"] = githubProject.stars
	badges["github-version"] = githubProject.tag
	badges["github-visibility"] = githubProject.visibility
	badges["github-watchers"] = githubProject.watchers
	badges["github-sloc"] = markdownBadge("https://sloc.xyz/github/{{.Namespace}}/{{.Name}}/", "{{.URL}}", isGitHub)
	// "github-forks":        markdownBadge("https://img.shields.io/github/forks/{{.Namespace}}/{{.Name}}?label=Fork", "{{.URL}}/network", isGitHub),
	// "github-issues":       markdownBadge("https://img.shields.io/github/issues/{{.Namespace}}/{{.Name}}", "{{.URL}}/issues", isGitHub),
	badges["github-lastcommit"] = githubProject.lastcommit
	// "github-license":      markdownBadge("https://img.shields.io/github/license/{{.Namespace}}/{{.Name}}", "{{.URL}}/blob/master/LICENSE", isGitHub),
	// "github-pullrequests": markdownBadge("https://img.shields.io/github/issues-pr/{{.Namespace}}/{{.Name}}", "{{.URL}}/pulls", isGitHub),
	// "github-size":         markdownBadge("https://img.shields.io/github/repo-size/{{.Namespace}}/{{.Name}}", "{{.URL}}", isGitHub),
	// "github-stars":        markdownBadge("https://img.shields.io/github/stars/{{.Namespace}}/{{.Name}}", "{{.URL}}/stargazers", isGitHub),
	// "github-version":      markdownBadge("https://img.shields.io/github/v/tag/{{.Namespace}}/{{.Name}}?sort=semver", "{{.URL}}", isGitHub),
	// "github-watchers":     markdownBadge("https://img.shields.io/github/watchers/{{.Namespace}}/{{.Name}}?label=Watch", "{{.URL}}/watchers", isGitHub),
}

type GithubProject struct {
	client                *github.Client
	locker                *locker.Locker
	repositoryCache       sync.Map
	pullrequestCountCache sync.Map
}

func NewGithubProject(githubAccessToken string) *GithubProject {
	return &GithubProject{
		client: github.NewClient(
			oauth2.NewClient(
				context.Background(),
				oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: githubAccessToken},
				),
			),
		),
		locker: locker.Initialize(),
	}
}

func (b *GithubProject) getProject(project Project) (*github.Repository, *Badge) {
	if !isGitHub(project) {
		return nil, svgBadge(project.Hoster, project.Name, "github", "github", "Not a GitHub project", badge.ColorLightgrey, project.URL, errors.New("not a GitHub project"))
	}

	b.locker.Lock("repo" + project.URL)
	defer b.locker.Unlock("repo" + project.URL)

	loadedGitHubProject, ok := b.repositoryCache.Load(project.URL)
	if ok {
		return loadedGitHubProject.(*github.Repository), nil
	}
	githubProject, _, err := b.client.Repositories.Get(context.Background(), project.Namespace, project.Name)
	if err != nil {
		return nil, svgBadge(project.Hoster, project.Name, "github", "github", "Error", badge.ColorLightgrey, project.URL, err)
	}

	b.repositoryCache.Store(project.URL, githubProject)
	return githubProject, nil
}

func (b *GithubProject) pullRequestCount(project Project) (int, error) {
	b.locker.Lock("pr" + project.URL)
	defer b.locker.Unlock("pr" + project.URL)

	count, ok := b.pullrequestCountCache.Load(project.URL)
	if !ok {
		opt := &github.PullRequestListOptions{
			ListOptions: github.ListOptions{PerPage: 1},
		}

		pullRequests, response, err := b.client.PullRequests.List(context.Background(), project.Namespace, project.Name, opt)
		if err != nil {
			return 0, err
		}
		prCount := response.LastPage
		if prCount == 0 {
			prCount = len(pullRequests)
		}
		b.pullrequestCountCache.Store(project.URL, prCount)
		return prCount, nil
	}
	return count.(int), nil
}

func (b *GithubProject) pullRequests(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	count, err := b.pullRequestCount(project)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "pullrequests", "pull requests", "Error", badge.ColorRed, project.URL+"/pulls", err)
	}

	color := badge.ColorBrightgreen
	if count > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "pullrequests", "pull requests", fmt.Sprintf("%d", count), color, project.URL+"/pulls", nil)
}

func (b *GithubProject) branches(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	_, response, err := b.client.Repositories.ListBranches(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "branches", "branches", "Error", badge.ColorLightgrey, project.URL+"/branches", err)
	}

	branchesCount := response.LastPage
	color := badge.ColorBrightgreen
	switch response.LastPage {
	case 0:
		fallthrough
	case 1:
		branchesCount = 1
	case 2:
		color = badge.ColorGreen
	default:
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "branches", "branches", fmt.Sprintf("%d", branchesCount), color, project.URL+"/branches", nil)
}

func (b *GithubProject) tag(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	tags, _, err := b.client.Repositories.ListTags(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "tag", "tag", "Error", badge.ColorLightgrey, project.URL+"/releases", err)
	}
	if len(tags) == 0 {
		return nil
	}
	return svgBadge(project.Hoster, project.Name, "tag", "tag", *(tags[0]).Name, badge.ColorBlue, project.URL+"/releases", nil)
}

func (b *GithubProject) commitssince(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	tags, _, err := b.client.Repositories.ListTags(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "tag", "tag", "Error", badge.ColorLightgrey, project.URL+"/releases", nil)
	}
	if len(tags) == 0 {
		return nil
	}
	return markdownBadge("https://img.shields.io/github/commits-since/{{.Namespace}}/{{.Name}}/latest", "{{.URL}}", isGitHub)(project)
}

func (b *GithubProject) issues(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return errBadge
	}

	if !*githubProject.HasIssues {
		return svgBadge(project.Hoster, project.Name, "issues", "issues", "disabled", badge.ColorLightgray, project.URL, nil)
	}

	pullRequestCount, err := b.pullRequestCount(project)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "issues", "issues", "Error", badge.ColorRed, project.URL+"/pulls", err)
	}

	issueCount := *githubProject.OpenIssuesCount - pullRequestCount

	color := badge.ColorBrightgreen
	if issueCount > 0 {
		color = badge.ColorYellow
	}

	return svgBadge(project.Hoster, project.Name, "issues", "issues", fmt.Sprintf("%d", issueCount), color, project.URL+"/issues", nil)
}

func (b *GithubProject) lastcommit(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return errBadge
	}

	color := badge.ColorRed
	switch {
	case time.Now().Add(-time.Hour * 24 * 30).Before(githubProject.UpdatedAt.Time):
		color = badge.ColorBrightgreen
	case time.Now().Add(-time.Hour * 24 * 60).Before(githubProject.UpdatedAt.Time):
		color = badge.ColorGreen
	case time.Now().Add(-time.Hour * 24 * 185).Before(githubProject.UpdatedAt.Time):
		color = badge.ColorYellowgreen
	case time.Now().Add(-time.Hour * 24 * 365).Before(githubProject.UpdatedAt.Time):
		color = badge.ColorYellow
	case time.Now().Add(-time.Hour * 24 * 730).Before(githubProject.UpdatedAt.Time):
		color = badge.ColorOrange
	}
	return svgBadge(project.Hoster, project.Name, "lastcommit", "last commit", humanize.Time(githubProject.UpdatedAt.Time), color, project.URL, nil)
}

func (b *GithubProject) stars(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return errBadge
	}
	return svgBadge(project.Hoster, project.Name, "stars", "stars", fmt.Sprint(*githubProject.StargazersCount), badge.ColorBlue, project.URL+"/stargazers", nil)
}

func (b *GithubProject) visibility(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return errBadge
	}

	color := badge.ColorGreen
	text := "public"
	if *githubProject.Private {
		text = "private"
		color = badge.ColorYellow
	}

	archived := ""
	if *githubProject.Archived {
		archived = " archived"
		color = badge.ColorLightgray
	}

	return svgBadge(project.Hoster, project.Name, "visibility", "visibility", text+archived, color, project.URL, nil)
}

func (b *GithubProject) forks(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return errBadge
	}
	return svgBadge(project.Hoster, project.Name, "fork", "Fork", fmt.Sprint(*githubProject.ForksCount), badge.ColorBlue, project.URL+"/network/members", nil)
}

func (b *GithubProject) size(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return errBadge
	}

	color := badge.ColorBrightgreen
	switch {
	case uint64(*githubProject.Size) > 1024*100:
		color = badge.ColorRed
	case uint64(*githubProject.Size) > 1024*50:
		color = badge.ColorOrange
	case uint64(*githubProject.Size) > 1024*10:
		color = badge.ColorYellow
	case uint64(*githubProject.Size) > 1024*5:
		color = badge.ColorYellowgreen
	case uint64(*githubProject.Size) > 1024:
		color = badge.ColorGreen
	}
	return svgBadge(project.Hoster, project.Name, "reposize", "repo size", humanize.Bytes(uint64(*githubProject.Size)*1024), color, project.URL, nil)
}

func (b *GithubProject) watchers(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return errBadge
	}
	return svgBadge(project.Hoster, project.Name, "watchers", "watchers", fmt.Sprint(*githubProject.SubscribersCount), badge.ColorBlue, project.URL, nil)
}

func (b *GithubProject) license(project Project) *Badge {
	if !isGitHub(project) {
		return nil
	}

	githubProject, errBadge := b.getProject(project)

	switch {
	case errBadge != nil:
		return errBadge
	case githubProject.License == nil:
		return svgBadge(project.Hoster, project.Name, "license", "license", "no License", badge.ColorRed, project.URL, nil)
	case *githubProject.License.SPDXID == "NOASSERTION":
		return svgBadge(project.Hoster, project.Name, "license", "license", "not recognized", badge.ColorLightgray, project.URL, nil)
	default:
		return svgBadge(project.Hoster, project.Name, "license", "license", fmt.Sprint(*githubProject.License.SPDXID), badge.ColorBlue, project.URL, nil)
	}
}
