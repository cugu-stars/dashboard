package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/v28/github"
	"github.com/narqo/go-badge"
	"golang.org/x/oauth2"
)

type GithubMergeRequests struct{}

func (b *GithubMergeRequests) Render(column string, project Project) string {
	if !isGitHub(project) {
		return ""
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_ACCESS_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	gl := github.NewClient(tc)

	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	}

	_, response, err := gl.PullRequests.List(context.Background(), project.Namespace, project.Name, opt)
	if err != nil {
		return svgBadge(project.Name, "mergerequests", "merge requests", err.Error(), badge.ColorLightgrey, project.URL)
	}

	color := badge.ColorBrightgreen
	if response.LastPage > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Name, "mergerequests", "merge requests", fmt.Sprintf("%d open", response.LastPage), color, project.URL)
}

type GithubBranches struct{}

func (b *GithubBranches) Render(column string, project Project) string {
	if !isGitHub(project) {
		return ""
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_ACCESS_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	gl := github.NewClient(tc)

	// opt := github.BranchListOptions{}
	_, response, err := gl.Repositories.ListBranches(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Name, "branches", "branches", err.Error(), badge.ColorLightgrey, project.URL)
	}

	branchesCount := response.LastPage
	if branchesCount == 0 {
		branchesCount = 1
	}

	color := badge.ColorBrightgreen
	if branchesCount > 1 {
		color = badge.ColorGreen
	}
	if branchesCount > 2 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Name, "branches", "branches", fmt.Sprintf("%d open", branchesCount), color, project.URL)
}

type GithubTag struct{}

func (b *GithubTag) Render(column string, project Project) string {
	if !isGitHub(project) {
		return ""
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_ACCESS_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	gl := github.NewClient(tc)

	tags, _, err := gl.Repositories.ListTags(context.Background(), project.Namespace, project.Name, &github.ListOptions{PerPage: 1})
	if err != nil {
		return svgBadge(project.Name, "tag", "tag", err.Error(), badge.ColorLightgrey, project.URL)
	}
	if len(tags) == 0 {
		return ""
	}
	tag := tags[0]
	return svgBadge(project.Name, "tag", "tag", *tag.Name, badge.ColorBlue, project.URL)
}

type GithubProject struct {
	field string
}

func (b *GithubProject) Render(column string, project Project) string {
	if !isGitHub(project) {
		return ""
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_ACCESS_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	gl := github.NewClient(tc)

	githubProject, _, err := gl.Repositories.Get(context.Background(), project.Namespace, project.Name)
	if err != nil {
		return svgBadge(project.Name, "github", "github", err.Error(), badge.ColorLightgrey, project.URL)
	}

	switch b.field {
	case "issues":
		color := badge.ColorBrightgreen
		if githubProject.GetOpenIssuesCount() > 0 {
			color = badge.ColorYellow
		}
		return svgBadge(project.Name, "issues", "issues", fmt.Sprintf("%d open", *githubProject.OpenIssuesCount), color, project.URL)
	case "lastcommit":
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
		return svgBadge(project.Name, "lastcommit", "last commit", humanize.Time(githubProject.UpdatedAt.Time), color, project.URL)
	case "stars":
		return svgBadge(project.Name, "stars", "stars", fmt.Sprint(*githubProject.StargazersCount), badge.ColorBlue, project.URL)
	case "visibility":
		color := badge.ColorGreen
		left := "public"
		if *githubProject.Private {
			color = badge.ColorLightgray
			left = "private"
		}
		return svgBadge(project.Name, "visibility", "visibility", left, color, project.URL)
	case "forks":
		return svgBadge(project.Name, "fork", "Fork", fmt.Sprint(*githubProject.ForksCount), badge.ColorBlue, project.URL)
	case "size":
		return svgBadge(project.Name, "reposize", "repo size", humanize.Bytes(uint64(*githubProject.Size)), badge.ColorBlue, project.URL)
	case "watchers":
		return svgBadge(project.Name, "watchers", "watchers", fmt.Sprint(*githubProject.SubscribersCount), badge.ColorBlue, project.URL)
	case "license":
		return svgBadge(project.Name, "license", "license", fmt.Sprint(githubProject.License.SPDXID), badge.ColorBlue, project.URL)

	}

	return svgBadge(project.Name, "error", "error", "unknown field", badge.ColorLightgrey, project.URL)
}
