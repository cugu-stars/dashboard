package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/enfipy/locker"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/v28/github"
	"github.com/narqo/go-badge"
	"golang.org/x/oauth2"
)

type GithubProject struct {
	client                *github.Client
	locker                *locker.Locker
	repositoryCache       sync.Map
	pullrequestCountCache sync.Map
}

func NewGithubProject() *GithubProject {
	return &GithubProject{
		client: github.NewClient(
			oauth2.NewClient(
				context.Background(),
				oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: GITHUB_ACCESS_TOKEN},
				),
			),
		),
		locker: locker.Initialize(),
	}
}

func (b *GithubProject) getProject(project Project) (*github.Repository, *string) {
	if !isGitHub(project) {
		s := ""
		return nil, &s
	}

	b.locker.Lock("repo" + project.URL)
	defer b.locker.Unlock("repo" + project.URL)

	loadedGitHubProject, ok := b.repositoryCache.Load(project.URL)
	if ok {
		return loadedGitHubProject.(*github.Repository), nil
	}
	githubProject, _, err := b.client.Repositories.Get(context.Background(), project.Namespace, project.Name)
	if err != nil {
		badge := svgBadge(project.Hoster, project.Name, "github", "github", err.Error(), badge.ColorLightgrey, project.URL)
		return nil, &badge
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

func (b *GithubProject) pullRequests(project Project) string {
	if !isGitHub(project) {
		return ""
	}

	count, err := b.pullRequestCount(project)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "pullrequests", "pull requests", err.Error(), badge.ColorRed, project.URL+"/pulls")
	}

	color := badge.ColorBrightgreen
	if count > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "pullrequests", "pull requests", fmt.Sprintf("%d", count), color, project.URL+"/pulls")
}

func (b *GithubProject) branches(project Project) string {
	if !isGitHub(project) {
		return ""
	}

	_, response, err := b.client.Repositories.ListBranches(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "branches", "branches", err.Error(), badge.ColorLightgrey, project.URL+"/branches")
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
	return svgBadge(project.Hoster, project.Name, "branches", "branches", fmt.Sprintf("%d", branchesCount), color, project.URL+"/branches")
}

func (b *GithubProject) tag(project Project) string {
	if !isGitHub(project) {
		return ""
	}

	tags, _, err := b.client.Repositories.ListTags(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "tag", "tag", err.Error(), badge.ColorLightgrey, project.URL+"/releases")
	}
	if len(tags) == 0 {
		return ""
	}
	return svgBadge(project.Hoster, project.Name, "tag", "tag", *(tags[0]).Name, badge.ColorBlue, project.URL+"/releases")
}

func (b *GithubProject) commitssince(project Project) string {
	if !isGitHub(project) {
		return ""
	}

	tags, _, err := b.client.Repositories.ListTags(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "tag", "tag", err.Error(), badge.ColorLightgrey, project.URL+"/releases")
	}
	if len(tags) == 0 {
		return ""
	}
	return markdownBadge("https://img.shields.io/github/commits-since/{{.Namespace}}/{{.Name}}/latest", "{{.URL}}", isGitHub)(project)
}

func (b *GithubProject) issues(project Project) string {
	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}

	if !*githubProject.HasIssues {
		return svgBadge(project.Hoster, project.Name, "issues", "issues", "disabled", badge.ColorLightgray, project.URL)
	}

	pullRequestCount, err := b.pullRequestCount(project)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "issues", "issues", err.Error(), badge.ColorRed, project.URL+"/pulls")
	}

	issueCount := *githubProject.OpenIssuesCount - pullRequestCount

	color := badge.ColorBrightgreen
	if issueCount > 0 {
		color = badge.ColorYellow
	}

	return svgBadge(project.Hoster, project.Name, "issues", "issues", fmt.Sprintf("%d", issueCount), color, project.URL+"/issues")
}

func (b *GithubProject) lastcommit(project Project) string {
	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
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
	return svgBadge(project.Hoster, project.Name, "lastcommit", "last commit", humanize.Time(githubProject.UpdatedAt.Time), color, project.URL)
}

func (b *GithubProject) stars(project Project) string {
	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}
	return svgBadge(project.Hoster, project.Name, "stars", "stars", fmt.Sprint(*githubProject.StargazersCount), badge.ColorBlue, project.URL+"/stargazers")
}

func (b *GithubProject) visibility(project Project) string {
	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
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
		color = badge.ColorGray
	}

	return svgBadge(project.Hoster, project.Name, "visibility", "visibility", text+archived, color, project.URL)
}

func (b *GithubProject) forks(project Project) string {
	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}
	return svgBadge(project.Hoster, project.Name, "fork", "Fork", fmt.Sprint(*githubProject.ForksCount), badge.ColorBlue, project.URL+"/network/members")
}

func (b *GithubProject) size(project Project) string {
	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}
	return svgBadge(project.Hoster, project.Name, "reposize", "repo size", humanize.Bytes(uint64(*githubProject.Size)*1024), badge.ColorBlue, project.URL)
}

func (b *GithubProject) watchers(project Project) string {
	githubProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}
	return svgBadge(project.Hoster, project.Name, "watchers", "watchers", fmt.Sprint(*githubProject.SubscribersCount), badge.ColorBlue, project.URL)
}

func (b *GithubProject) license(project Project) string {
	githubProject, errBadge := b.getProject(project)

	switch {
	case errBadge != nil:
		return *errBadge
	case githubProject.License == nil:
		return svgBadge(project.Hoster, project.Name, "license", "license", "no License", badge.ColorRed, project.URL)
	case *githubProject.License.SPDXID == "NOASSERTION":
		return svgBadge(project.Hoster, project.Name, "license", "license", "not recognized", badge.ColorLightgray, project.URL)
	default:
		return svgBadge(project.Hoster, project.Name, "license", "license", fmt.Sprint(*githubProject.License.SPDXID), badge.ColorBlue, project.URL)
	}
}
